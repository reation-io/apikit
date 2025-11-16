package fileupload

import (
	"context"
	"fmt"
	"mime/multipart"
)

// UploadRequest represents a file upload request with form data
type UploadRequest struct {
	// in:form title
	Title string `form:"title" validate:"required"`

	// in:form description
	Description string `form:"description"`

	// in:form file
	File *multipart.FileHeader `form:"file" validate:"required"`

	// in:form tags
	Tags []string `form:"tags"`
}

// UploadResponse represents the upload response
type UploadResponse struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// swagger:route POST /upload files uploadFile
// Consumes: multipart/form-data
// Summary: Upload a file with metadata
// Description: Uploads a file along with title, description, and tags
// Responses:
// - 200: UploadResponse
// - 400: ErrorResponse
// - 500: ErrorResponse
type UploadFileRoute struct{}

// apikit:handler
func UploadFile(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	// In a real implementation, you would:
	// 1. Open the file: file, err := req.File.Open()
	// 2. Save it to storage (S3, local disk, etc.)
	// 3. Return the file metadata

	return UploadResponse{
		ID:       "file-123",
		Filename: req.File.Filename,
		Size:     req.File.Size,
	}, nil
}

// MultiUploadRequest represents a multiple file upload request
type MultiUploadRequest struct {
	// in:form title
	Title string `form:"title" validate:"required"`

	// in:form files
	Files []*multipart.FileHeader `form:"files" validate:"required,min=1"`
}

// MultiUploadResponse represents the multiple upload response
type MultiUploadResponse struct {
	UploadedFiles []FileInfo `json:"uploaded_files"`
}

// FileInfo represents information about an uploaded file
type FileInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// swagger:route POST /upload/multiple files uploadMultipleFiles
// Consumes: multipart/form-data
// Summary: Upload multiple files
// Description: Uploads multiple files at once with a title
// Responses:
// - 200: MultiUploadResponse
// - 400: ErrorResponse
// - 500: ErrorResponse
type UploadMultipleFilesRoute struct{}

// apikit:handler
func UploadMultipleFiles(ctx context.Context, req MultiUploadRequest) (MultiUploadResponse, error) {
	files := make([]FileInfo, len(req.Files))
	for i, file := range req.Files {
		files[i] = FileInfo{
			ID:       fmt.Sprintf("file-%d", i+1),
			Filename: file.Filename,
			Size:     file.Size,
		}
	}

	return MultiUploadResponse{
		UploadedFiles: files,
	}, nil
}

// ErrorResponse represents an error response
// swagger:model
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
