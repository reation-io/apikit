package swagger

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

// Handler provides HTTP handlers for serving Swagger UI and OpenAPI specs.
type Handler struct {
	config    *Config
	resources []byte
	fileServer http.Handler
}

// NewHandler creates a new Swagger UI handler with the given options.
func NewHandler(opts ...Option) (*Handler, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	h := &Handler{
		config: cfg,
	}

	// Build resources JSON
	if err := h.buildResources(); err != nil {
		return nil, err
	}

	// Setup file server
	if err := h.setupFileServer(); err != nil {
		return nil, err
	}

	return h, nil
}

// buildResources creates the JSON response for /openapi/resources endpoint.
func (h *Handler) buildResources() error {
	type resource struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	resources := make([]resource, 0, len(h.config.specs))
	for name := range h.config.specs {
		resources = append(resources, resource{
			Name: name,
			URL:  h.config.basePath + "/specs?q=" + name,
		})
	}

	data, err := json.Marshal(resources)
	if err != nil {
		return err
	}

	h.resources = data
	return nil
}

// setupFileServer initializes the file server for Swagger UI assets.
func (h *Handler) setupFileServer() error {
	var fileSystem fs.FS

	if h.config.uiFS != nil {
		fileSystem = h.config.uiFS
	} else {
		// Use embedded default UI
		reader := bytes.NewReader(defaultSwaggerUI)
		zipReader, err := zip.NewReader(reader, int64(len(defaultSwaggerUI)))
		if err != nil {
			return err
		}
		fileSystem = zipReader
	}

	h.fileServer = http.FileServer(http.FS(fileSystem))
	return nil
}

// ServeHTTP implements http.Handler interface.
// It routes requests to the appropriate handler based on the path.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle base path redirect: /swagger -> /swagger/
	if path == h.config.uiPath || path == strings.TrimSuffix(h.config.uiPath, "/") {
		if !strings.HasSuffix(path, "/") {
			http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
			return
		}
	}

	// Route to appropriate handler
	switch {
	case path == h.config.basePath+"/resources":
		h.handleResources(w, r)
	case path == h.config.basePath+"/specs":
		h.handleSpecs(w, r)
	case strings.HasPrefix(path, h.config.uiPath):
		h.handleUI(w, r)
	default:
		http.NotFound(w, r)
	}
}

// Handler returns an http.Handler that can be mounted on a router.
// The prefix is stripped from incoming requests.
func (h *Handler) Handler(prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle redirect for paths without trailing slash
		if r.URL.Path == prefix && !strings.HasSuffix(prefix, "/") {
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
			return
		}

		// Strip prefix and serve
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		h.serveInternal(w, r)
	})
}

// serveInternal handles requests after prefix stripping.
func (h *Handler) serveInternal(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/resources":
		h.handleResources(w, r)
	case path == "/specs":
		h.handleSpecs(w, r)
	default:
		// Serve UI files
		h.fileServer.ServeHTTP(w, r)
	}
}

// handleResources serves the list of available specs.
func (h *Handler) handleResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(h.resources)
}

// handleSpecs serves the OpenAPI specification.
func (h *Handler) handleSpecs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	spec, ok := h.config.specs[q]
	if !ok {
		// Return first spec if no query or not found
		for _, s := range h.config.specs {
			spec = s
			break
		}
	}

	if spec == nil {
		http.Error(w, "no spec available", http.StatusNotFound)
		return
	}

	// Detect content type based on spec content
	contentType := "application/json"
	if len(spec) > 0 && (spec[0] == '-' || spec[0] == '#' || !isJSON(spec)) {
		contentType = "application/x-yaml"
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(spec)
}

// handleUI serves the Swagger UI files.
func (h *Handler) handleUI(w http.ResponseWriter, r *http.Request) {
	// Strip UI path prefix
	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.config.uiPath)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	h.fileServer.ServeHTTP(w, r)
}

// isJSON checks if the byte slice starts with JSON content.
func isJSON(data []byte) bool {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return true
		default:
			return false
		}
	}
	return false
}

// Routes returns handlers for manual router registration.
// This is useful when you need more control over routing.
type Routes struct {
	handler *Handler
}

// GetRoutes returns a Routes helper for manual registration.
func (h *Handler) GetRoutes() *Routes {
	return &Routes{handler: h}
}

// Resources returns the handler for /resources endpoint.
func (r *Routes) Resources() http.HandlerFunc {
	return r.handler.handleResources
}

// Specs returns the handler for /specs endpoint.
func (r *Routes) Specs() http.HandlerFunc {
	return r.handler.handleSpecs
}

// UI returns the handler for serving Swagger UI files.
// The prefix should be the path where UI is mounted (e.g., "/swagger").
func (r *Routes) UI(prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Handle redirect for path without trailing slash
		if req.URL.Path == prefix {
			http.Redirect(w, req, prefix+"/", http.StatusMovedPermanently)
			return
		}

		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		r.handler.fileServer.ServeHTTP(w, req)
	})
}
