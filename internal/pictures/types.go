package pictures

import (
	"mime/multipart"
	"sync"

	"github.com/VicSobDev/anniversaryAPI/pkg/db"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type PicturesService struct {
	basePath string
	mx       sync.Mutex
	logger   *zap.Logger
	SQLiteDB *db.SQLiteDB
}

type FileError struct {
	Message string
}

type pagination struct {
	limit  int
	offset int
}

func (e *FileError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type FileSaver interface {
	SaveUploadedFile(*multipart.FileHeader, string) error
}

func NewPicturesService(basePath string, logger *zap.Logger, sqliteDB *db.SQLiteDB) *PicturesService {
	// Register metrics with Prometheus's default registry
	prometheus.MustRegister(getPicturesRequests)
	prometheus.MustRegister(getPictureRequests)
	prometheus.MustRegister(uploadPictureRequests)

	return &PicturesService{basePath: basePath, logger: logger, SQLiteDB: sqliteDB}
}
