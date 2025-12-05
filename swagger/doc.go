// Package swagger provides an HTTP handler for serving Swagger UI
// and OpenAPI specifications.
//
// Basic usage:
//
//	handler, err := swagger.NewHandler(
//		swagger.WithSpec("api", specBytes),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Option 1: Use as standalone handler
//	http.Handle("/swagger/", handler.Handler("/swagger"))
//	http.Handle("/openapi/", handler)
//
//	// Option 2: Use with any router (chi, gorilla, etc.)
//	routes := handler.GetRoutes()
//	router.Get("/openapi/resources", routes.Resources())
//	router.Get("/openapi/specs", routes.Specs())
//	router.Mount("/swagger", routes.UI("/swagger"))
//
// The handler automatically redirects /swagger to /swagger/ to ensure
// relative paths in Swagger UI work correctly.
//
// Multiple specs can be registered:
//
//	handler, _ := swagger.NewHandler(
//		swagger.WithSpec("v1", v1Spec),
//		swagger.WithSpec("v2", v2Spec),
//		swagger.WithSpec("legacy", legacySpec),
//	)
//
// Specs are accessible via query parameter: /openapi/specs?q=v1
package swagger
