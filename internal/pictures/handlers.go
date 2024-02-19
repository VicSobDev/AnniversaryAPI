package pictures

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetPictures retrieves a paginated list of pictures based on the request parameters.
func (svc *PicturesService) GetPictures(c *gin.Context) {
	// Log the invocation of the GetPictures function
	svc.logger.Info("GetPictures called")

	// Parse pagination parameters (limit and offset) from the request
	pagination := svc.parsePaginationParams(c)

	// Log the details of the pagination request
	svc.logger.Info("getting pictures", zap.Int("limit", pagination.limit), zap.Int("offset", pagination.offset))

	// Retrieve a paginated list of images from the database
	images, err := svc.SQLiteDB.GetImagesPaginated(pagination.limit, pagination.offset)
	if err != nil {
		// Log the error and respond with an internal server error if the database query fails
		svc.ErrorHandler(getPicturesRequests, err, zap.String("error", "failed to get pictures"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get pictures"})
		return
	}

	// Log the total number of pictures retrieved and increment the success metric
	svc.logger.Info("sending pictures", zap.Int("total", len(images)))
	getPicturesRequests.WithLabelValues("successful").Inc()

	// Respond with the retrieved images and a 200 OK status
	c.JSON(http.StatusOK, images)
}

// GetPicture serves a specific picture file to the client with access restrictions.
func (svc *PicturesService) GetPicture(c *gin.Context) {
	// Check if the current request is made on allowed dates
	if time.Now().Day() != anniversaryDay && (time.Now().Month() != valentineDay && time.Now().Month() != valentineMonth) {
		// Log a warning and deny access if the request is outside of the specified dates
		svc.logger.Warn("GetPicture called outside of anniversary and valentine day")
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Log the successful access to the function on allowed dates
	svc.logger.Info("GetPicture called")

	// Extract the picture name from the query parameters
	name := c.Query("name")

	// Construct the file path by joining the base path with the requested picture name
	filePath := filepath.Join(svc.basePath, name)

	// Ensure that the constructed file path does not escape the base directory
	if !strings.HasPrefix(filePath, svc.basePath) {
		// Log and respond with an error if an attempt to access outside the base path is detected
		svc.ErrorHandler(getPictureRequests, errors.New("attempt to access a file outside the base path"), zap.String("error", "invalid file path"), zap.String("name", name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}

	// Lock around the file access to prevent race conditions
	svc.mx.Lock()
	_, err := os.Stat(filePath)
	svc.mx.Unlock()
	if os.IsNotExist(err) {
		// Respond with a not found error if the file does not exist
		svc.ErrorHandler(getPictureRequests, err, zap.String("error", "picture not found"), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{"error": "picture not found"})
		return
	}

	// Increment the metric counter for successful picture requests
	getPictureRequests.WithLabelValues("successful").Inc()

	// Log the action of sending the picture and serve the file to the client
	svc.logger.Info("sending picture", zap.String("name", name))
	c.File(filePath)
}

// GetTotalPictures retrieves and sends the total number of pictures stored in the database.
func (svc *PicturesService) GetTotalPictures(c *gin.Context) {
	// Log the invocation of the GetTotalImages function
	svc.logger.Info("GetTotalImages called")

	// Retrieve the total number of pictures from the database
	total, err := svc.SQLiteDB.GetTotalPictures()
	if err != nil {
		// Log the error and respond with an internal server error if the query fails
		svc.ErrorHandler(getTotalPicturesRequests, err, zap.String("error", "failed to get total images"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get total images"})
		return
	}

	// Increment the metric counter for successful requests
	getTotalPicturesRequests.WithLabelValues("successful").Inc()

	// Log the total number of images retrieved and send this information back to the client
	svc.logger.Info("sending total images", zap.Int("total", total))
	c.JSON(http.StatusOK, gin.H{"total": total})
}

// UploadPictures handles the uploading of multiple picture files from a client.
func (svc *PicturesService) UploadPictures(c *gin.Context) {
	// Log the invocation of the UploadPictures endpoint
	svc.logger.Info("UploadPictures called")

	// Retrieve the multipart form from the request
	form, err := c.MultipartForm()
	if err != nil {
		// Handle errors related to multipart form processing
		svc.handleUploadPictureError(c, err)
		return
	}

	// Extract the files from the "pictures" key in the multipart form
	files := form.File["pictures"]
	// Check if no files were uploaded and handle the error
	if len(files) == 0 {
		svc.handleUploadPictureError(c, errors.New("no pictures uploaded"))
		return
	}

	var successfullyUploaded []string // Track successfully uploaded file names
	var failedUploads []string        // Track file names of failed uploads

	// Iterate over the uploaded files
	for _, file := range files {
		// Validate and save each file, capturing the safe file name or an error
		safeFileName, err := svc.validateAndSaveFile(c, file)
		if err != nil {
			// If an error occurs, log it and add the file to the list of failed uploads
			failedUploads = append(failedUploads, file.Filename)
			svc.logger.Error("Failed to upload file", zap.Error(err))
			continue
		}

		// If uploaded successfully, add the safe file name to the success list
		successfullyUploaded = append(successfullyUploaded, safeFileName)
	}

	// Extract user ID from context, added by an earlier middleware or handler
	userID := c.GetInt("user_id")

	// For each successfully uploaded file, create a record in the database
	for _, filename := range successfullyUploaded {
		err := svc.SQLiteDB.CreateImage(userID, filename, time.Now().Unix())
		if err != nil {
			// Log any errors that occur while saving to the database
			svc.logger.Error("failed to save image to database", zap.Error(err))
			continue
		}
	}

	// If there are any successful uploads, send a confirmation response
	if len(successfullyUploaded) > 0 {
		svc.logger.Info("Files uploaded successfully", zap.Strings("paths", successfullyUploaded))
		c.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "paths": successfullyUploaded})
	}

	// If there are any failed uploads, inform the client
	if len(failedUploads) > 0 {
		svc.logger.Warn("Some files failed to upload", zap.Strings("files", failedUploads))
		c.JSON(http.StatusBadRequest, gin.H{"error": "some files failed to upload", "failed_files": failedUploads})
	}

	// Increment the metric counter for successful picture uploads
	uploadPictureRequests.WithLabelValues("successful").Inc()

}
