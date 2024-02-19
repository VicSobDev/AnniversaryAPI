package auth

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/VicSobDev/anniversaryAPI/pkg/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorHandler increments a Prometheus counter for tracking errors and logs the error with additional fields.
func (svc *AuthService) ErrorHandler(cv *prometheus.CounterVec, err error, fields ...zapcore.Field) {
	cv.WithLabelValues("error").Inc()        // Increment the error counter for metrics
	svc.logger.Error(err.Error(), fields...) // Log the error with additional contextual fields
}

// verifyPassword checks if the provided password matches the stored hash.
func (svc *AuthService) verifyPassword(storedHash, providedPassword string) (bool, error) {
	// Use Argon2 to verify the password against the stored hash
	valid, err := svc.argon.Verify(storedHash, providedPassword)
	if err != nil {
		return false, err // Return false and the error if verification fails
	}
	return valid, nil // Return true if the password is valid, else false
}

// attemptLogin processes a login request, verifying the user's credentials.
func (svc *AuthService) attemptLogin(req LoginRequest) (*db.User, error) {
	username := strings.ToLower(req.Username) // Normalize the username to lowercase
	// Retrieve the user from the database using the username provided in the login request
	user, err := svc.db.GetUser(username)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no user is found, return an error indicating that the user was not found
			return nil, errUserNotFound
		}
		// Log the error and return if there's a database or other error
		message := "Invalid username or password"
		svc.ErrorHandler(loginAttempts, err, zap.String("error", message))
		return nil, err
	}

	// Verify the password provided in the login request against the stored hash
	valid, err := svc.verifyPassword(user.Password, req.Password)
	if err != nil {
		// Log and return an error if there's an issue during password verification
		svc.ErrorHandler(loginAttempts, err, zap.String("error", "Internal server error"))
		return nil, err
	}

	if !valid {
		// If the password does not match, log the error and return an "invalid password" error
		return nil, errInvalidPassword
	}

	// Return the user object if the login is successful
	return user, nil
}

// generateAndSendToken creates a JWT token for the authenticated user and returns it.
func (svc *AuthService) generateAndSendToken(user *db.User, c *gin.Context) (string, error) {
	// Initialize a new JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Cast the token claims to a MapClaims object and set username and user ID
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username // Include the username in the token payload
	claims["user_id"] = user.ID        // Include the user ID in the token payload

	// Sign the token with the server's secret key to generate the token string
	tokenString, err := token.SignedString(svc.jwtKey)
	if err != nil {
		// If an error occurs during token signing, log the error and return it
		svc.ErrorHandler(loginAttempts, err, zap.String("error", "failed to sign token"))
		return "", err // Return an empty string and the error
	}

	// Return the signed token string and nil for the error if token generation was successful
	return tokenString, nil
}

// validateToken parses and validates a JWT token string, extracting user information if valid.
func (a *AuthService) validateToken(tokenString string, c *gin.Context) (db.User, bool) {
	// Parse the token using the specified callback function to provide the key for verification
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token uses the expected HMAC signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtKey, nil // Provide the secret key for the HMAC method
	})

	// Handle parsing errors by logging and responding with an unauthorized status
	if err != nil {
		a.logger.Error("failed to parse token", zap.Error(err))
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort() // Stop further handlers from executing
		return db.User{}, false
	}

	// Check if the token is actually valid
	if !token.Valid {
		a.logger.Error("invalid token")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort() // Stop further handlers from executing
		return db.User{}, false
	}

	// Attempt to extract claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		a.logger.Error("failed to extract claims from token")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort() // Stop further handlers from executing
		return db.User{}, false
	}

	// Log the extracted claims for debugging purposes
	a.logger.Info("claims", zap.Any("claims", claims))

	// Extract the user ID from the claims
	userID, ok := claims["user_id"].(float64) // JWT standard encodes numeric IDs as floats
	if !ok {
		a.logger.Error("user_id not found in claims")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort() // Stop further handlers from executing
		return db.User{}, false
	}

	// Extract the username from the claims
	username, ok := claims["username"].(string)
	if !ok {
		a.logger.Error("username not found in claims")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort() // Stop further handlers from executing
		return db.User{}, false
	}

	// Reconstruct the user object from the claims
	user := db.User{
		ID:       int(userID), // Convert the float64 ID back to an int
		Username: username,
	}

	// Log the successful validation of the user
	a.logger.Info("user validated", zap.String("username", user.Username))
	return user, true // Return the reconstructed user and true indicating the token is valid
}
