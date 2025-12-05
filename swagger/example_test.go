package swagger_test

import (
	"log"
	"net/http"

	"github.com/reation-io/apikit/swagger"
)

func Example() {
	// Sample OpenAPI spec
	spec := []byte(`
openapi: "3.0.0"
info:
  title: Sample API
  version: "1.0.0"
paths: {}
`)

	// Create handler with spec
	handler, err := swagger.NewHandler(
		swagger.WithSpec("api", spec),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Mount handlers
	// The UI handler automatically redirects /swagger to /swagger/
	http.Handle("/swagger/", handler.Handler("/swagger"))
	http.Handle("/swagger", handler.Handler("/swagger")) // Handles redirect

	// API endpoints
	routes := handler.GetRoutes()
	http.HandleFunc("/openapi/resources", routes.Resources())
	http.HandleFunc("/openapi/specs", routes.Specs())

	log.Println("Server starting on :8080")
	log.Println("Swagger UI: http://localhost:8080/swagger")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_multipleSpecs() {
	v1Spec := []byte(`{"openapi":"3.0.0","info":{"title":"API v1","version":"1.0.0"}}`)
	v2Spec := []byte(`{"openapi":"3.0.0","info":{"title":"API v2","version":"2.0.0"}}`)

	handler, err := swagger.NewHandler(
		swagger.WithSpec("v1", v1Spec),
		swagger.WithSpec("v2", v2Spec),
	)
	if err != nil {
		log.Fatal(err)
	}

	// All specs available at:
	// - /openapi/specs?q=v1
	// - /openapi/specs?q=v2
	// - /openapi/resources (lists all available specs)

	http.Handle("/", handler.Handler(""))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_withRouter() {
	spec := []byte(`openapi: "3.0.0"`)

	handler, _ := swagger.NewHandler(
		swagger.WithSpec("api", spec),
		swagger.WithBasePath("/api/docs"),
		swagger.WithUIPath("/docs/"),
	)

	routes := handler.GetRoutes()

	// For chi router:
	// r := chi.NewRouter()
	// r.Get("/api/docs/resources", routes.Resources())
	// r.Get("/api/docs/specs", routes.Specs())
	// r.Mount("/docs", routes.UI("/docs"))

	// For gorilla mux:
	// r := mux.NewRouter()
	// r.HandleFunc("/api/docs/resources", routes.Resources())
	// r.HandleFunc("/api/docs/specs", routes.Specs())
	// r.PathPrefix("/docs").Handler(routes.UI("/docs"))

	_ = routes
}
