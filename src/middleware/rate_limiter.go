package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// --- Token Bucket Lua Script ---
// KEYS[1]: The key for the user (e.g., rate_limit:<ip>)
// ARGV[1]: Rate (tokens recovery per second)
// ARGV[2]: Burst (bucket capacity)
// ARGV[3]: Current Unix timestamp (seconds)
// Returns: 1 if allowed, 0 if denied
const tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1 -- Consume 1 token per request

local data = redis.call('HMGET', key, 'tokens', 'ts')
local tokens = tonumber(data[1])
local last_ts = tonumber(data[2])

if tokens == nil or last_ts == nil then
  -- First request, initialize bucket
  tokens = capacity
  last_ts = now
end

-- Calculate tokens to add based on elapsed time
local elapsed = math.max(0, now - last_ts)
local added_tokens = elapsed * rate
tokens = math.min(capacity, tokens + added_tokens)
last_ts = now -- Update last timestamp regardless of allow/deny

local allowed = 0
if tokens >= requested then
  tokens = tokens - requested
  allowed = 1
end

-- Save new state and set expiration (e.g., 24 hours)
redis.call('HMSET', key, 'tokens', tokens, 'ts', last_ts)
redis.call('EXPIRE', key, 86400) -- Keep state for 24h

return allowed
`

var scriptSHA string // Stores the SHA hash of the loaded Lua script

// loadTokenBucketScript loads the Lua script into Redis and stores its SHA hash.
func loadTokenBucketScript(ctx context.Context, rdb *redis.Client) {
	var err error
	scriptSHA, err = rdb.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		log.Fatalf("fatal: failed to load rate limiter lua script: %v", err)
	}
	log.Printf("rate limiter lua script loaded successfully. SHA: %s", scriptSHA)
}

// RateLimiter creates a custom token bucket rate limiting middleware using Redis.
// Expects RATE_LIMIT env var like "100-M" (100 requests per Minute).
func RateLimiter() gin.HandlerFunc {
	// Initialize script on startup if not already done
	if scriptSHA == "" {
		// Use a background context for script loading during initialization
		loadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if RDB == nil {
			log.Fatal("fatal: redis client (RDB) not initialized before loading rate limiter script.")
		}
		loadTokenBucketScript(loadCtx, RDB)
	}

	// --- Parse Configuration ---
	rateStr := os.Getenv("RATE_LIMIT")
	if rateStr == "" {
		rateStr = "100-M" // Default if not set
		log.Printf("waning: RATE_LIMIT environment variable not set, defaulting to %s", rateStr)
	}

	parts := strings.SplitN(rateStr, "-", 2)
	if len(parts) != 2 {
		log.Fatalf("fatal: invalid RATE_LIMIT format '%s'. expected format like '100-M' or '5-S'.", rateStr)
	}

	limit, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		log.Fatalf("fatal: invalid number '%s' in RATE_LIMIT: %v", parts[0], err)
	}

	var ratePerSecond float64
	unit := strings.ToUpper(parts[1])
	switch unit {
	case "S":
		ratePerSecond = limit
	case "M":
		ratePerSecond = limit / 60.0
	case "H":
		ratePerSecond = limit / 3600.0
	default:
		log.Fatalf("invalid unit '%s' in RATE_LIMIT. use S (Second), M (Minute), or H (Hour).", unit)
	}

	// Burst capacity is typically set equal to the limit specified for the interval
	burst := limit

	// Ensure rate and burst are valid numbers
	if ratePerSecond <= 0 || burst <= 0 {
		log.Fatalf("fatal: RATE_LIMIT '%s' results in non-positive rate/burst values.", rateStr)
	}

	log.Printf("rate limiter configured: rate %.2f tokens/sec, burst %.0f tokens", ratePerSecond, burst)

	// --- Return Gin Handler ---
	return func(c *gin.Context) {
		if RDB == nil {
			log.Println("error: rate limiter cannot function, redis client is nil.")
			// Fail open or closed? Failing closed (500 error) is safer.
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter configuration error"})
			return
		}

		// Use client IP as the key. Consider X-Forwarded-For if behind a trusted proxy.
		// TODO: Add configuration to trust X-Forwarded-For if needed.
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		now := float64(time.Now().Unix())

		// Execute the Lua script atomically
		// Use EvalSha for performance, falling back to Eval if the script isn't loaded (e.g., Redis restart)
		result, err := RDB.EvalSha(c.Request.Context(), scriptSHA, []string{key}, ratePerSecond, burst, now).Result()

		if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT") {
			// Script not found, try loading and executing again with EVAL
			log.Printf("warning: rate limiter script SHA %s not found. reloading script.", scriptSHA)
			loadTokenBucketScript(c.Request.Context(), RDB) // Reload script
			result, err = RDB.Eval(c.Request.Context(), tokenBucketScript, []string{key}, ratePerSecond, burst, now).Result()
		}

		if err != nil {
			log.Printf("error: rate limiter redis command failed for IP %s: %v", ip, err)
			// Fail open? Let the request through if Redis fails? Or deny?
			// Denying might block legitimate users if Redis is flaky.
			// Let's fail open here but log aggressively. Consider your security posture.
			c.Next()
			return
		}

		// Check Lua script result (1 for allowed, 0 for denied)
		allowed, ok := result.(int64)
		if !ok || allowed == 0 {
			// Deny request
			// Optionally add Retry-After header (calculating this accurately needs more state)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "You have exceeded the request rate limit.",
			})
			return
		}

		// Allow request
		c.Next()
	}
}
