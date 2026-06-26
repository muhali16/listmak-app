package middlewares

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/pkg/utils"
	"golang.org/x/time/rate"
)

// RateLimiter limits requests per client IP. r = tokens/sec, burst = max burst.
// ponytail: per-instance sync.Map grows per unique IP; add TTL eviction if deployed at scale
func RateLimiter(r rate.Limit, burst int) gin.HandlerFunc {
	var visitors sync.Map
	return func(c *gin.Context) {
		actual, _ := visitors.LoadOrStore(c.ClientIP(), rate.NewLimiter(r, burst))
		if !actual.(*rate.Limiter).Allow() {
			utils.SendResponse(c, http.StatusTooManyRequests, false, "Rate limit exceeded", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
