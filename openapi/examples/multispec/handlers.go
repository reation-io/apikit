package multispec

// swagger:meta
// Title: Admin API
// Version: 1.0.0
// Spec: admin
// Description: Administrative API for managing the system
type AdminMeta struct{}

// swagger:meta
// Title: Mobile API
// Version: 1.0.0
// Spec: mobile
// Description: Mobile application API
type MobileMeta struct{}

// swagger:meta
// Title: Public API
// Version: 1.0.0
// Spec: public
// Description: Public API for external clients
type PublicMeta struct{}

// swagger:model
type User struct {
	// User ID
	// Example: 123
	ID int `json:"id"`

	// User name
	// Example: John Doe
	Name string `json:"name"`

	// User email
	// Example: john@example.com
	Email string `json:"email"`
}

// swagger:model
type ErrorResponse struct {
	// Error message
	// Example: Invalid request
	Message string `json:"message"`
}

// swagger:model
type HealthResponse struct {
	// Service status
	// Example: ok
	Status string `json:"status"`
}

// swagger:route GET /admin/users admin listAdminUsers
// Spec: admin
// Summary: List all users (admin only)
// Description: Returns a list of all users in the system. Requires admin privileges.
// Responses:
// - 200: User
// - 401: ErrorResponse
// Security:
// - bearer
type AdminListUsers struct{}

// swagger:route POST /admin/users admin createUser
// Spec: admin
// Summary: Create a new user (admin only)
// Description: Creates a new user in the system. Requires admin privileges.
// Responses:
// - 201: User
// - 400: ErrorResponse
// - 401: ErrorResponse
// Security:
// - bearer
type AdminCreateUser struct{}

// swagger:route GET /mobile/profile mobile getProfile
// Spec: mobile
// Summary: Get user profile
// Description: Returns the authenticated user's profile information
// Responses:
// - 200: User
// - 401: ErrorResponse
// Security:
// - bearer
type MobileGetProfile struct{}

// swagger:route PUT /mobile/profile mobile updateProfile
// Spec: mobile
// Summary: Update user profile
// Description: Updates the authenticated user's profile information
// Responses:
// - 200: User
// - 400: ErrorResponse
// - 401: ErrorResponse
// Security:
// - bearer
type MobileUpdateProfile struct{}

// swagger:route GET /users/{id} users getUserByID
// Spec: admin mobile public
// Summary: Get user by ID
// Description: Returns a single user by their ID. Available in all APIs.
// Responses:
// - 200: User
// - 404: ErrorResponse
type GetUserByID struct{}

// swagger:route GET /health health healthCheck
// Summary: Health check endpoint
// Description: Returns the health status of the service
// Responses:
// - 200: HealthResponse
type HealthCheck struct{}

// swagger:route GET /public/info public getPublicInfo
// Spec: public
// Summary: Get public information
// Description: Returns public information about the service
// Responses:
// - 200: HealthResponse
type GetPublicInfo struct{}

