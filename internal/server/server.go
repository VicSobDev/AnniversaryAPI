package server

import (
	"github.com/VicSobDev/anniversaryAPI/internal/auth"
	"github.com/VicSobDev/anniversaryAPI/internal/pictures"
	"github.com/VicSobDev/anniversaryAPI/pkg/crypto"
	"github.com/VicSobDev/anniversaryAPI/pkg/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Api struct definition
type Api struct {
	listenAddr    string
	jwtKey        []byte
	prometheusKey string
	apiKey        string
	logger        *zap.Logger
}

// NewApi constructor
func NewApi(listenAddr string, jwtKey []byte, prometheusKey string, apiKey string) *Api {
	return &Api{
		listenAddr:    listenAddr,
		jwtKey:        jwtKey,
		prometheusKey: prometheusKey,
		apiKey:        apiKey,
	}
}

// Start initializes and starts the API server
func (a *Api) Start() error {
	// Initialize dependencies: Logger and Database
	logger, err := a.initializeLogger()
	if err != nil {
		return err
	}

	a.logger = logger

	sqliteDB, err := a.initializeDatabase()
	if err != nil {
		return err
	}

	// Initialize services
	argon := a.initializeCryptoService()
	picturesService, authService := a.initializeServices(sqliteDB, argon, logger)

	// Setup and start the API server
	r := a.setupServer(logger, picturesService, authService)
	return r.Run(a.listenAddr)
}

// initializeLogger sets up the application logger
func (a *Api) initializeLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

// initializeDatabase sets up and migrates the database
func (a *Api) initializeDatabase() (*db.SQLiteDB, error) {
	sqliteDB, err := db.NewSQLiteDB("db.sqlite")
	if err != nil {
		return nil, err
	}

	if err := sqliteDB.Migrate(); err != nil {
		return nil, err
	}

	return sqliteDB, nil
}

// initializeCryptoService sets up the crypto service
func (a *Api) initializeCryptoService() *crypto.Argon2 {
	return crypto.NewArgon2(crypto.Argon2Config{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
}

// initializeServices sets up the application services
func (a *Api) initializeServices(sqliteDB *db.SQLiteDB, argon *crypto.Argon2, logger *zap.Logger) (*pictures.PicturesService, *auth.AuthService) {
	picturesService := pictures.NewPicturesService("images", logger, sqliteDB)
	authService := auth.NewAuthService(logger, sqliteDB, argon, a.jwtKey, a.apiKey)
	return picturesService, authService
}

// setupServer configures and returns the Gin server
func (a *Api) setupServer(logger *zap.Logger, picturesService *pictures.PicturesService, authService *auth.AuthService) *gin.Engine {
	// Create a new Gin router
	r := gin.Default()

	// Configure and apply CORS middleware
	r.Use(a.configureCORS())

	// Setup API routes
	a.setupRoutes(r, authService, picturesService)

	// Setup and run the metrics server in a separate goroutine
	a.setupMetricsServer(logger)

	return r
}

// configureCORS returns the CORS middleware configuration
func (a *Api) configureCORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Customize as needed
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	return cors.New(config)
}

// setupRoutes configures the API endpoints
func (a *Api) setupRoutes(r *gin.Engine, authService *auth.AuthService, picturesService *pictures.PicturesService) {
	api := r.Group("/api")

	// Authentication routes
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/register", authService.Register)
		authRoutes.POST("/login", authService.Login)
		authRoutes.GET("/refresh", authService.Refresh)
		authRoutes.POST("/keys", authService.AddKey)
		authRoutes.GET("/keys", authService.GetKeys)

	}

	// Protected routes
	api.Use(a.AuthMiddleware)
	{
		api.GET("/pictures", picturesService.GetPictures)
		api.GET("/picture", picturesService.GetPicture)
		api.POST("/pictures", picturesService.UploadPictures)
		api.GET("/pictures_total", picturesService.GetTotalPictures)
	}

}

// setupMetricsServer initializes and starts a separate server for metrics
func (a *Api) setupMetricsServer(logger *zap.Logger) {
	go func() {
		metricsRouter := gin.Default()
		metricsRouter.Use(a.PrometheusAuthMiddleware)

		metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))
		if err := metricsRouter.Run(":8081"); err != nil {
			logger.Error("Failed to start metrics server", zap.Error(err))
		}
	}()
}
