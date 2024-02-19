package auth

import (
	"github.com/VicSobDev/anniversaryAPI/pkg/crypto"
	"github.com/VicSobDev/anniversaryAPI/pkg/db"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type AuthService struct {
	logger zap.Logger
	db     *db.SQLiteDB
	argon  *crypto.Argon2
	apiKey string
	jwtKey []byte
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

type AddKeyRequest struct {
	Key string `json:"key"`
}

func NewAuthService(logger *zap.Logger, db *db.SQLiteDB, argon *crypto.Argon2, jwtKey []byte, apiKey string) *AuthService {
	prometheus.MustRegister(loginAttempts)
	prometheus.MustRegister(registerAttempts)
	return &AuthService{logger: *logger, db: db, argon: argon, jwtKey: jwtKey, apiKey: apiKey}
}
