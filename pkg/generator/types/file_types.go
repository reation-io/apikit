package types

import "fmt"

func init() {
	// Register file upload types
	registerFileTypes()
}

// registerFileTypes registers multipart file types
func registerFileTypes() {
	// *multipart.FileHeader - single file upload
	DefaultRegistry.Register(&Extractor{
		TypeName: "*multipart.FileHeader",
		Import:   "mime/multipart",
		ParseFunc: func(varName, fieldName string, isPointer bool) string {
			// File fields are handled by FormExtractor
			// This is just for type registration
			return fmt.Sprintf("payload.%s = %s", fieldName, varName)
		},
		RequiresError: false,
	})

	// []*multipart.FileHeader - multiple file uploads
	DefaultRegistry.Register(&Extractor{
		TypeName: "[]*multipart.FileHeader",
		Import:   "mime/multipart",
		ParseFunc: func(varName, fieldName string, isPointer bool) string {
			// File fields are handled by FormExtractor
			// This is just for type registration
			return fmt.Sprintf("payload.%s = %s", fieldName, varName)
		},
		RequiresError: false,
	})
}
