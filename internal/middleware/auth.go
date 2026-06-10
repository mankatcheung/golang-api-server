// Package middleware provides HTTP middleware functions for request processing.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/pkg/jwt"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("auth rejected", "reason", "missing authorization header", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			slog.Warn("auth rejected", "reason", "invalid header format", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(secret, parts[1])
		if err != nil {
			slog.Warn("auth rejected", "reason", "invalid or expired token", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			slog.Warn("authz rejected", "reason", "missing role claim", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			c.Abort()
			return
		}

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		slog.Warn("authz rejected",
			"reason", "insufficient permissions",
			"role", role,
			"required", roles,
			"ip", c.ClientIP(),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}
