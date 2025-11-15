package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/reation-io/apikit/openapi/builder"
)

// swagger:meta
// Title: User Management API
// Version: 1.0.0
// Description: A comprehensive API for managing users
//   This API provides endpoints for creating, reading, updating, and deleting users.
//   It also includes authentication and authorization features.
type API struct{}

// swagger:route POST /users user createUser
// Summary: Create a new user
// Description: Creates a new user in the system
// Tags: users, admin
type CreateUserRequest struct{}

// swagger:route GET /users/{id} user getUser
// Summary: Get user by ID
// Tags: users
type GetUserRequest struct{}

// swagger:route PUT /users/{id} user updateUser
// Summary: Update user
// Tags: users
type UpdateUserRequest struct{}

// swagger:route DELETE /users/{id} user deleteUser
// Summary: Delete user
// Tags: users, admin
// Deprecated: true
type DeleteUserRequest struct{}

// swagger:model
type User struct {
	// Example: 123
	// Minimum: 1
	ID int `json:"id"`

	// Example: john.doe@example.com
	// Format: email
	// MinLength: 5
	// MaxLength: 100
	Email string `json:"email"`

	// Example: John Doe
	// MinLength: 3
	// MaxLength: 50
	Name string `json:"name"`

	// Example: 25
	// Minimum: 0
	// Maximum: 150
	Age int `json:"age"`

	// Example: true
	Active bool `json:"active"`
}

// swagger:model
type CreateUserPayload struct {
	// Example: john.doe@example.com
	// Format: email
	Email string `json:"email"`

	// Example: John Doe
	Name string `json:"name"`

	// Example: 25
	Age int `json:"age"`
}

func main() {
	// Build the OpenAPI spec from the current directory
	b := builder.NewBuilder("*.go")
	spec, err := b.Build()
	if err != nil {
		log.Fatalf("Failed to build OpenAPI spec: %v", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Print the spec
	fmt.Println(string(data))
}

