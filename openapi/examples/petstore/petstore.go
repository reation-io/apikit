package petstore

// swagger:meta
// title: Swagger Petstore - OpenAPI 3.0
// version: 1.0.12
// description:
//
//	This is a sample Pet Store Server based on the OpenAPI 3.0 specification.  You can find out more about
//	Swagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!
//	You can now help us improve the API whether it's by making changes to the definition itself or to the code.
//	That way, with time, we can improve the API in general, and expose some of the new features in OAS3.
//
//	Some useful links:
//	- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)
//	- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)
type PetstoreMeta struct{}

// ============================================================================
// MODELS
// ============================================================================

// Category represents a pet category
// swagger:model
type Category struct {
	// Category ID
	// example: 1
	ID int64 `json:"id"`
	// Category name
	// example: Dogs
	Name string `json:"name"`
}

// Tag represents a pet tag
// swagger:model
type Tag struct {
	// Tag ID
	ID int64 `json:"id"`
	// Tag name
	Name string `json:"name"`
}

// Pet represents a pet in the store
// swagger:model
type Pet struct {
	// Pet ID
	// example: 10
	ID int64 `json:"id"`
	// Pet name
	// required: true
	// example: doggie
	Name string `json:"name"`
	// Pet category
	Category *Category `json:"category,omitempty"`
	// Photo URLs
	// required: true
	PhotoUrls []string `json:"photoUrls"`
	// Pet tags
	Tags []Tag `json:"tags,omitempty"`
	// Pet status in the store
	// enum: available,pending,sold
	Status string `json:"status,omitempty"`
}

// Order represents a store order
// swagger:model
type Order struct {
	// Order ID
	// example: 10
	ID int64 `json:"id"`
	// Pet ID
	// example: 198772
	PetID int64 `json:"petId"`
	// Quantity
	// example: 7
	Quantity int32 `json:"quantity"`
	// Ship date
	ShipDate string `json:"shipDate,omitempty"`
	// Order status
	// example: approved
	// enum: placed,approved,delivered
	Status string `json:"status,omitempty"`
	// Is complete
	Complete bool `json:"complete,omitempty"`
}

// User represents a user
// swagger:model
type User struct {
	// User ID
	// example: 10
	ID int64 `json:"id"`
	// Username
	// example: theUser
	Username string `json:"username"`
	// First name
	// example: John
	FirstName string `json:"firstName"`
	// Last name
	// example: James
	LastName string `json:"lastName"`
	// Email
	// example: john@email.com
	Email string `json:"email"`
	// Password
	// example: 12345
	Password string `json:"password"`
	// Phone
	// example: 12345
	Phone string `json:"phone"`
	// User status
	// example: 1
	UserStatus int32 `json:"userStatus"`
}

// ApiResponse represents an API response
// swagger:model
type ApiResponse struct {
	// Response code
	Code int32 `json:"code"`
	// Response type
	Type string `json:"type"`
	// Response message
	Message string `json:"message"`
}

// Error represents an error response
// swagger:model
type Error struct {
	// Error code
	// required: true
	Code string `json:"code"`
	// Error message
	// required: true
	Message string `json:"message"`
}

// ============================================================================
// PET ROUTES
// ============================================================================

// swagger:route PUT /pet pet updatePet
//
// Update an existing pet by Id.
//
// summary: Update an existing pet.
// description: Update an existing pet by Id.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: Pet
// - 400: Error
// - 404: Error
// - 422: Error
// - default: Error
type UpdatePetRequest struct {
	// in: body
	Body Pet
}

// swagger:route POST /pet pet addPet
//
// Add a new pet to the store.
//
// summary: Add a new pet to the store.
// description: Add a new pet to the store.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: Pet
// - 400: Error
// - 422: Error
// - default: Error
type AddPetRequest struct {
	// in: body
	Body Pet
}

// swagger:route GET /pet/findByStatus pet findPetsByStatus
//
// Finds Pets by status.
//
// summary: Finds Pets by status.
// description: Multiple status values can be provided with comma separated strings.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: PetListResponse
// - 400: Error
// - default: Error
type FindPetsByStatusRequest struct {
	// Status values that need to be considered for filter
	// in: query
	// enum: available,pending,sold
	// default: available
	Status string `json:"status"`
}

// PetListResponse is a list of pets
// swagger:model
type PetListResponse struct {
	// in: body
	Body []Pet
}

// swagger:route GET /pet/findByTags pet findPetsByTags
//
// Finds Pets by tags.
//
// summary: Finds Pets by tags.
// description: Multiple tags can be provided with comma separated strings. Use tag1, tag2, tag3 for testing.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: PetListResponse
// - 400: Error
// - default: Error
type FindPetsByTagsRequest struct {
	// Tags to filter by
	// in: query
	Tags []string `json:"tags"`
}

// swagger:route GET /pet/{petId} pet getPetById
//
// Returns a single pet.
//
// summary: Find pet by ID.
// description: Returns a single pet.
//
// Security:
// - api_key
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: Pet
// - 400: Error
// - 404: Error
// - default: Error
type GetPetByIDRequest struct {
	// ID of pet to return
	// in: path
	// required: true
	PetID int64 `json:"petId"`
}

// swagger:route POST /pet/{petId} pet updatePetWithForm
//
// Updates a pet in the store with form data.
//
// summary: Updates a pet in the store with form data.
// description: Updates a pet resource based on the form data.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: Pet
// - 400: Error
// - default: Error
type UpdatePetWithFormRequest struct {
	// ID of pet that needs to be updated
	// in: path
	// required: true
	PetID int64 `json:"petId"`
	// Name of pet that needs to be updated
	// in: query
	Name string `json:"name"`
	// Status of pet that needs to be updated
	// in: query
	Status string `json:"status"`
}

// swagger:route DELETE /pet/{petId} pet deletePet
//
// Delete a pet.
//
// summary: Deletes a pet.
// description: Delete a pet.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: SuccessResponse
// - 400: Error
// - default: Error
type DeletePetRequest struct {
	// API key
	// in: header
	APIKey string `json:"api_key"`
	// Pet id to delete
	// in: path
	// required: true
	PetID int64 `json:"petId"`
}

// SuccessResponse represents a successful operation
// swagger:model
type SuccessResponse struct {
	// Success message
	Message string `json:"message"`
}

// swagger:route POST /pet/{petId}/uploadImage pet uploadFile
//
// Upload image of the pet.
//
// summary: Uploads an image.
// description: Upload image of the pet.
//
// Security:
// - petstore_auth:
//   - write:pets
//   - read:pets
//
// Responses:
// - 200: ApiResponse
// - 400: Error
// - 404: Error
// - default: Error
type UploadFileRequest struct {
	// ID of pet to update
	// in: path
	// required: true
	PetID int64 `json:"petId"`
	// Additional Metadata
	// in: query
	AdditionalMetadata string `json:"additionalMetadata"`
}

// ============================================================================
// STORE ROUTES
// ============================================================================

// swagger:route GET /store/inventory store getInventory
//
// Returns a map of status codes to quantities.
//
// summary: Returns pet inventories by status.
// description: Returns a map of status codes to quantities.
//
// Security:
// - api_key
//
// Responses:
// - 200: InventoryResponse
// - default: Error
type GetInventoryRequest struct{}

// InventoryResponse represents inventory counts
// swagger:model
type InventoryResponse struct {
	// in: body
	Body map[string]int32
}

// swagger:route POST /store/order store placeOrder
//
// Place a new order in the store.
//
// summary: Place an order for a pet.
// description: Place a new order in the store.
//
// Responses:
// - 200: Order
// - 400: Error
// - 422: Error
// - default: Error
type PlaceOrderRequest struct {
	// in: body
	Body Order
}

// swagger:route GET /store/order/{orderId} store getOrderById
//
// For valid response try integer IDs with value <= 5 or > 10. Other values will generate exceptions.
//
// summary: Find purchase order by ID.
// description: For valid response try integer IDs with value <= 5 or > 10. Other values will generate exceptions.
//
// Responses:
// - 200: Order
// - 400: Error
// - 404: Error
// - default: Error
type GetOrderByIDRequest struct {
	// ID of order that needs to be fetched
	// in: path
	// required: true
	OrderID int64 `json:"orderId"`
}

// swagger:route DELETE /store/order/{orderId} store deleteOrder
//
// For valid response try integer IDs with value < 1000. Anything above 1000 or nonintegers will generate API errors.
//
// summary: Delete purchase order by identifier.
// description: For valid response try integer IDs with value < 1000. Anything above 1000 or nonintegers will generate API errors.
//
// Responses:
// - 200: SuccessResponse
// - 400: Error
// - 404: Error
// - default: Error
type DeleteOrderRequest struct {
	// ID of the order that needs to be deleted
	// in: path
	// required: true
	OrderID int64 `json:"orderId"`
}
