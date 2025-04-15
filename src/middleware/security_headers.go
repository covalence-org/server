package middleware

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

// SecureHeaders applies production-focused security headers.
func SecureHeaders() gin.HandlerFunc {
	isDevelopment := os.Getenv("GIN_MODE") == "debug"

	// Base CSP - **MUST be customized for your specific application needs**
	// This example allows loading resources only from the same origin.
	// You will likely need to add domains for CDNs, APIs, fonts, etc.
	// Use https://csp-evaluator.withgoogle.com/ to test policies.
	csp := "default-src 'self'; script-src 'self'; style-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none';"

	secureMiddleware := secure.New(secure.Options{
		// Assuming deployment behind a TLS-terminating proxy (like Nginx, Caddy, LB)
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},                    // Tell secure it's behind a proxy
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"}, // Header indicating HTTPS
		STSSeconds:            31536000,                                        // 1 year
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,          // Provides *some* protection in older browsers
		ContentSecurityPolicy: csp,           // **Customize Heavily**
		IsDevelopment:         isDevelopment, // Allows more relaxed settings in dev if needed
	})

	return func(c *gin.Context) {
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			// Log errors in production, but don't necessarily block requests
			// unless header injection is absolutely critical.
			log.Printf("warning: failed to process security headers: %v", err)
		}
		c.Next() // Always call next, even if headers failed
	}
}
