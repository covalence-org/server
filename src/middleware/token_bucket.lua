-- token_bucket.lua
-- KEYS[1]: The key for the user (e.g., rate_limit:<ip>)
-- ARGV[1]: Rate (tokens recovery per second)
-- ARGV[2]: Burst (bucket capacity)
-- ARGV[3]: Current Unix timestamp (seconds)
-- Returns: 1 if allowed, 0 if denied

local key       = KEYS[1]
local rate      = tonumber(ARGV[1])
local capacity  = tonumber(ARGV[2])
local now       = tonumber(ARGV[3])
local requested = 1

local data    = redis.call('HMGET', key, 'tokens', 'ts')
local tokens  = tonumber(data[1])
local last_ts = tonumber(data[2])

if not tokens or not last_ts then
  tokens  = capacity
  last_ts = now
end

local elapsed      = math.max(0, now - last_ts)
local added_tokens = elapsed * rate
tokens             = math.min(capacity, tokens + added_tokens)
last_ts            = now

local allowed = 0
if tokens >= requested then
  tokens  = tokens - requested
  allowed = 1
end

redis.call('HMSET', key, 'tokens', tokens, 'ts', last_ts)
redis.call('EXPIRE', key, 86400)

return allowed
