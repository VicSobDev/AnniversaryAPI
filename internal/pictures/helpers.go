package pictures

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// generateSafeFileName generates a new filename to prevent directory traversal and overwriting sensitive files.
func (svc *PicturesService) generateSafeFileName(originalFileName string) string {
	// Extract the extension of the original file
	ext := filepath.Ext(originalFileName)

	// Use SHA-256 hashing over the original filename with a timestamp to ensure uniqueness
	hash := sha256.New()
	hash.Write([]byte(originalFileName + time.Now().String()))
	hashedFileName := fmt.Sprintf("%x", hash.Sum(nil))

	// Append the original extension to maintain the file type
	safeFileName := hashedFileName + ext

	return safeFileName
}

// handleUploadPictureError handles different types of errors by sending appropriate responses.
func (svc *PicturesService) handleUploadPictureError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *FileError:
		svc.ErrorHandler(uploadPictureRequests, e, zap.String("error", "failed to upload picture"))
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
	case *ValidationError:
		svc.ErrorHandler(uploadPictureRequests, e, zap.String("error", "invalid file type"))
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
	default:
		svc.ErrorHandler(uploadPictureRequests, err, zap.String("error", "internal server error"), zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// validateAndSaveFile validates the file type and saves the file if valid.
func (svc *PicturesService) validateAndSaveFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return "", &ValidationError{Message: "Invalid file type"}
	}

	filename := svc.generateSafeFileName(file.Filename)
	savePath := filepath.Join(svc.basePath, filename)

	if err := svc.saveUploadedFile(file, savePath); err != nil {
		return "", &FileError{Message: "Failed to save the file"}
	}

	return filename, nil
}

func (svc *PicturesService) saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	svc.mx.Lock()
	defer svc.mx.Unlock()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// parsePaginationParams extracts and validates pagination parameters from the request.
// Returns the validated limit and offset values.
func (svc *PicturesService) parsePaginationParams(c *gin.Context) pagination {
	// Default values
	limit := 3
	offset := 0

	// Parsing limit
	if queryLimit, ok := c.GetQuery("limit"); ok {
		if newLimit, err := strconv.Atoi(queryLimit); err == nil && newLimit > 0 && newLimit <= 100 {
			limit = newLimit
		} else {
			svc.logger.Warn("Invalid limit provided, using default", zap.String("limit", queryLimit))
		}
	}

	// Parsing offset
	if queryOffset, ok := c.GetQuery("offset"); ok {
		if newOffset, err := strconv.Atoi(queryOffset); err == nil && newOffset >= 0 {
			offset = newOffset
		} else {
			svc.logger.Warn("Invalid offset provided, using default", zap.String("offset", queryOffset))
		}
	}

	return pagination{limit: limit, offset: offset}
}

func (svc *PicturesService) ErrorHandler(cv *prometheus.CounterVec, err error, fields ...zapcore.Field) {
	cv.WithLabelValues("error").Inc()
	svc.logger.Error(err.Error(), fields...)
}
