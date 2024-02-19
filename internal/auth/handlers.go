package auth

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Login handles user login requests.
func (svc *AuthService) Login(c *gin.Context) {
	// Log that the login function was called
	svc.logger.Info("Login called")

	var req LoginRequest
	// Bind the incoming JSON request to a LoginRequest struct; handle errors
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the error and respond with a bad request status if request binding fails
		svc.ErrorHandler(loginAttempts, err, zap.String("error", "invalid request"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Attempt to login with the provided credentials
	user, err := svc.attemptLogin(req)
	if err != nil {
		if err == errUserNotFound || err == errInvalidPassword {
			loginAttempts.WithLabelValues("invalid_credentials").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		// Log the error and respond with an unauthorized status if login fails
		svc.ErrorHandler(loginAttempts, err, zap.String("error", "could not login"))
		return
	}

	// Check if the user was not found or invalid credentials were provided
	if user == nil {
		// Log the error and respond with an unauthorized status if user is nil
		loginAttempts.WithLabelValues("invalid_credentials").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Generate and send a token for the logged-in user
	token, err := svc.generateAndSendToken(user, c)
	if err != nil {
		// Log the error and respond with an internal server error status if token generation fails
		svc.ErrorHandler(loginAttempts, err, zap.String("error", "failed to generate token"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	loginAttempts.WithLabelValues("successful").Inc()
	// Respond with the generated token upon successful login
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Register handles user registration requests.
func (svc *AuthService) Register(c *gin.Context) {
	// Log the invocation of the Register function
	svc.logger.Info("Register called")

	var req RegisterRequest
	// Bind the incoming JSON request to a RegisterRequest struct; handle errors
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the error and respond with a bad request status if request binding fails
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "invalid request"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Check if the provided registration key exists in the database
	if exists, err := svc.db.ContainsKey(req.Key); err != nil {
		// Log the error and respond with an internal server error status if a database error occurs
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "internal server error"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	} else if !exists {
		// Respond with an unauthorized status if the registration key is invalid
		registerAttempts.WithLabelValues("invalid_key").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
		return
	}

	username := strings.ToLower(req.Username)

	// Check if the username is already taken
	if _, err := svc.db.GetUser(username); err == nil {
		// Respond with a conflict status if the username already exists
		registerAttempts.WithLabelValues("username_exists").Inc()
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// Hash the provided password
	hash, err := svc.argon.Hash(req.Password)
	if err != nil {
		// Log the error and respond with an internal server error if password hashing fails
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "failed to hash password"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Create a new user with the hashed password
	user, err := svc.db.CreateUser(username, hash)
	if err != nil {
		// Log the error and respond with an internal server error if user creation fails
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "failed to create user"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Delete the registration key used for signing up
	if err := svc.db.DeleteKey(req.Key); err != nil {
		// Log the error and respond with an internal server error if key deletion fails
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "failed to delete key"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Initialize a new JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims for the JWT token (e.g., username and user ID)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID

	// Generate a signed JWT token string
	tokenString, err := token.SignedString(svc.jwtKey)
	if err != nil {
		// Log the error and respond with an internal server error if token generation fails
		svc.ErrorHandler(registerAttempts, err, zap.String("error", "failed to generate token"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Increment the register attempts metric for successful registration
	registerAttempts.WithLabelValues("successful").Inc()
	// Log the successful registration
	svc.logger.Info("user registered", zap.String("username", user.Username), zap.Int("user_id", user.ID))
	// Respond with the generated token upon successful registration
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Refresh handles token refresh requests, generating a new token for authenticated users.
func (svc *AuthService) Refresh(c *gin.Context) {
	// Log the invocation of the Refresh function
	svc.logger.Info("Refresh called")

	// Retrieve the Authorization header from the request
	authHeader := c.GetHeader("Authorization")
	// Check if the Authorization header is missing
	if authHeader == "" {
		// Log the error and respond with an unauthorized status if the header is missing
		svc.ErrorHandler(refreshAttempts, nil, zap.String("error", "Authorization header is missing"))
		c.JSON(401, gin.H{"error": "Authorization header is missing"})
		return
	}

	const prefix = "Bearer "
	// Verify the Authorization header format starts with "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		// Log the error and respond with an unauthorized status if the header format is incorrect
		svc.ErrorHandler(refreshAttempts, nil, zap.String("error", "Authorization header must start with Bearer"))
		c.JSON(401, gin.H{"error": "Authorization header must start with Bearer"})
		return
	}

	// Extract the JWT token from the Authorization header
	jwtToken := authHeader[len(prefix):]
	// Check if the token is missing after removing the prefix
	if jwtToken == "" {
		// Log the error and respond with an unauthorized status if the token is missing
		svc.ErrorHandler(refreshAttempts, nil, zap.String("error", "Token is missing"))
		c.JSON(401, gin.H{"error": "Token is missing"})
		return
	}

	// Validate the extracted token and retrieve the associated user if valid
	user, valid := svc.validateToken(jwtToken, c)
	// Proceed only if the token is valid; otherwise, respond with an unauthorized status
	if !valid {
		svc.ErrorHandler(refreshAttempts, nil, zap.String("error", "Unauthorized"))
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Initialize a new JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims for the new token (e.g., username and user ID)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID

	// Sign the new token with the server's secret key
	tokenString, err := token.SignedString(svc.jwtKey)
	if err != nil {
		// Log the error and respond with an unauthorized status if signing the token fails
		svc.logger.Error("failed to sign token", zap.Error(err))
		c.JSON(401, gin.H{"error": "failed to sign token"})
		return
	}

	// Log the successful token refresh
	svc.logger.Info("token refreshed")

	refreshAttempts.WithLabelValues("successful").Inc()

	// Respond with the new token upon successful refresh
	c.JSON(200, gin.H{"token": tokenString})
}

func (svc *AuthService) GetKeys(c *gin.Context) {

	api_key := c.GetHeader("api_key")

	if api_key != svc.apiKey {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	keys, err := svc.db.GetKeys()
	if err != nil {
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(200, gin.H{"keys": keys})
}

func (svc *AuthService) AddKey(c *gin.Context) {
	api_key := c.GetHeader("api_key")

	if api_key != svc.apiKey {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	var req AddKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	err := svc.db.AddKey(req.Key)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(200, gin.H{"message": "key added"})
}
