package extractors

import (
	"fmt"

	"github.com/reation-io/apikit/parser"
)

func init() {
	RegisterFramework(&FiberExtractor{})
}

// FiberExtractor implements FrameworkExtractor for gofiber/fiber
type FiberExtractor struct{}

func (e *FiberExtractor) Name() string {
	return "fiber"
}

func (e *FiberExtractor) Imports() []string {
	return []string{"github.com/gofiber/fiber/v2"}
}

func (e *FiberExtractor) HandlerSignature() string {
	return "fiber.Handler"
}

func (e *FiberExtractor) ParseFuncSignature() string {
	return "func parse%s(c *fiber.Ctx, payload *%s) error"
}

func (e *FiberExtractor) ExtractQuery(field *parser.Field, paramName, fieldName string) (string, []string) {
	varName := fmt.Sprintf(`c.Query("%s")`, paramName)
	return GenerateCodeByType(varName, fieldName, GetBaseType(field), field)
}

func (e *FiberExtractor) ExtractQuerySlice(field *parser.Field, paramName, fieldName string) (string, []string) {
	// Fiber doesn't have native slice query support, need to use QueryParser or manual
	code := fmt.Sprintf(`if vals := c.Context().QueryArgs().PeekMulti("%s"); len(vals) > 0 {
		payload.%s = make([]string, len(vals))
		for i, v := range vals {
			payload.%s[i] = string(v)
		}
	}`, paramName, fieldName, fieldName)
	return code, nil
}

func (e *FiberExtractor) ExtractPath(field *parser.Field, paramName, fieldName string) (string, []string) {
	varName := fmt.Sprintf(`c.Params("%s")`, paramName)
	return GenerateCodeByType(varName, fieldName, GetBaseType(field), field)
}

func (e *FiberExtractor) ExtractHeader(field *parser.Field, paramName, fieldName string) (string, []string) {
	varName := fmt.Sprintf(`c.Get("%s")`, paramName)
	return GenerateCodeByType(varName, fieldName, GetBaseType(field), field)
}

func (e *FiberExtractor) ExtractHeaderSlice(field *parser.Field, paramName, fieldName string) (string, []string) {
	// Fiber returns single header value, wrap in slice
	code := fmt.Sprintf(`if val := c.Get("%s"); val != "" {
		payload.%s = []string{val}
	}`, paramName, fieldName)
	return code, nil
}

func (e *FiberExtractor) ExtractCookie(field *parser.Field, paramName, fieldName string) (string, []string) {
	varName := fmt.Sprintf(`c.Cookies("%s")`, paramName)
	return GenerateCodeByType(varName, fieldName, GetBaseType(field), field)
}

func (e *FiberExtractor) ExtractForm(field *parser.Field, paramName, fieldName string) (string, []string) {
	varName := fmt.Sprintf(`c.FormValue("%s")`, paramName)
	return GenerateCodeByType(varName, fieldName, field.Type, field)
}

func (e *FiberExtractor) ExtractFormSlice(field *parser.Field, paramName, fieldName string) (string, []string) {
	code := fmt.Sprintf(`if form, err := c.MultipartForm(); err == nil && form != nil {
		if vals := form.Value["%s"]; len(vals) > 0 {
			payload.%s = vals
		}
	}`, paramName, fieldName)
	return code, nil
}

func (e *FiberExtractor) ExtractFormFile(field *parser.Field, paramName, fieldPath string) (string, []string) {
	code := fmt.Sprintf(`if header, err := c.FormFile("%s"); err == nil {
		payload.%s = header
	}`, paramName, fieldPath)
	return code, []string{"mime/multipart"}
}

func (e *FiberExtractor) ExtractFormFiles(field *parser.Field, paramName, fieldPath string) (string, []string) {
	code := fmt.Sprintf(`if form, err := c.MultipartForm(); err == nil && form != nil {
		if files := form.File["%s"]; len(files) > 0 {
			payload.%s = files
		}
	}`, paramName, fieldPath)
	return code, []string{"mime/multipart"}
}

func (e *FiberExtractor) ExtractRequest(field *parser.Field) (string, []string) {
	// Fiber doesn't have *http.Request, skip
	return "", nil
}

func (e *FiberExtractor) ExtractResponse(field *parser.Field) (string, []string) {
	// Fiber doesn't have http.ResponseWriter, skip
	return "", nil
}

func (e *FiberExtractor) ParseBody(structName string) string {
	return `if err := c.BodyParser(&payload); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}`
}

func (e *FiberExtractor) WriteJSON(varName string) string {
	return fmt.Sprintf(`if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(%s)`, varName)
}

func (e *FiberExtractor) WriteError(statusCode int, message string) string {
	return fmt.Sprintf(`return c.Status(%d).JSON(fiber.Map{"error": "%s"})`, statusCode, message)
}

