package helpers

// AppError represents application-level error
type AppError struct {
	Code    int
	Message string
	Details interface{}
}

// NewAppError creates a new app error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithDetails creates app error with details
func NewAppErrorWithDetails(code int, message string, details interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error codes
var (
	ErrInvalidRequest = NewAppError(400, "invalid request")
	ErrUnauthorized = NewAppError(401, "unauthorized")
	ErrForbidden = NewAppError(403, "forbidden")
	ErrNotFound = NewAppError(404, "not found")
	ErrConflict = NewAppError(409, "conflict")
	ErrValidation = NewAppError(422, "validation error")
	ErrInternal = NewAppError(500, "internal server error")
)
