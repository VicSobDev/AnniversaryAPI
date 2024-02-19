package server

import (
	"fmt"

	"github.com/VicSobDev/anniversaryAPI/pkg/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (a *Api) validateToken(tokenString string, c *gin.Context) (db.User, bool) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's algorithm matches what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.jwtKey, nil
	})

	if err != nil {
		a.logger.Error("failed to parse token", zap.Error(err))
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return db.User{}, false
	}

	// Check if the token is valid
	if !token.Valid {
		a.logger.Error("invalid token")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return db.User{}, false
	}

	// Extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		a.logger.Error("failed to extract claims from token")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return db.User{}, false
	}

	a.logger.Info("claims", zap.Any("claims", claims))

	// Extract user ID and username from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		a.logger.Error("user_id not found in claims")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return db.User{}, false
	}

	username, ok := claims["username"].(string)
	if !ok {
		a.logger.Error("username not found in claims")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return db.User{}, false
	}

	// Construct the user object
	user := db.User{
		ID:       int(userID),
		Username: username,
	}

	a.logger.Info("user validated", zap.String("username", user.Username))
	return user, true
}
