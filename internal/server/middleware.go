package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *Api) AuthMiddleware(c *gin.Context) {
	// Extract the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(401, gin.H{"error": "Authorization header is missing"})
		c.Abort()
		return
	}

	// Expect the header to be "Bearer <token>"
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.JSON(401, gin.H{"error": "Authorization header must start with Bearer"})
		c.Abort()
		return
	}

	// Extract the JWT token from the header
	jwtToken := authHeader[len(prefix):]
	if jwtToken == "" {
		c.JSON(401, gin.H{"error": "Token is missing"})
		c.Abort()
		return
	}

	// Validate the token
	user, ok := a.validateToken(jwtToken, c)
	if !ok {
		// validateToken method handles response to client
		return
	}

	// Set the user object in the context
	c.Set("user_id", user.ID)
	c.Set("username", user.Username)

	c.Next()
}

func (a *Api) PrometheusAuthMiddleware(c *gin.Context) {
	apiKey := c.GetHeader("authorization")

	if !strings.Contains(apiKey, "Bearer") {
		c.JSON(401, gin.H{"error": "Authorization header must start with Bearer"})
		c.Abort()
		return
	}

	apiKey = strings.Split(apiKey, "Bearer ")[1]

	if apiKey != a.prometheusKey {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

}
