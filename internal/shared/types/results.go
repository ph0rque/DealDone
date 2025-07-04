package types

// Result represents a generic result with value and error
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a new result
func NewResult[T any](value T, err error) Result[T] {
	return Result[T]{
		Value: value,
		Error: err,
	}
}

// IsSuccess returns true if the result has no error
func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

// ProcessingResult represents the result of a processing operation
type ProcessingResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Errors  []string               `json:"errors,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewProcessingResult creates a new processing result
func NewProcessingResult(success bool, message string) *ProcessingResult {
	return &ProcessingResult{
		Success: success,
		Message: message,
		Details: make(map[string]interface{}),
		Errors:  make([]string, 0),
	}
}

// WithData adds data to the result
func (r *ProcessingResult) WithData(data interface{}) *ProcessingResult {
	r.Data = data
	return r
}

// WithError adds an error to the result
func (r *ProcessingResult) WithError(err string) *ProcessingResult {
	r.Errors = append(r.Errors, err)
	r.Success = false
	return r
}

// WithDetail adds a detail to the result
func (r *ProcessingResult) WithDetail(key string, value interface{}) *ProcessingResult {
	if r.Details == nil {
		r.Details = make(map[string]interface{})
	}
	r.Details[key] = value
	return r
}

// OperationResult represents the result of a file or data operation
type OperationResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ItemCount int                    `json:"itemCount,omitempty"`
	Items     []interface{}          `json:"items,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewOperationResult creates a new operation result
func NewOperationResult(success bool, message string) *OperationResult {
	return &OperationResult{
		Success:  success,
		Message:  message,
		Metadata: make(map[string]interface{}),
	}
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid    bool                   `json:"valid"`
	Errors   []ValidationError      `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// NewValidationResult creates a new validation result
func NewValidationResult(valid bool) *ValidationResult {
	return &ValidationResult{
		Valid:    valid,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
		Details:  make(map[string]interface{}),
	}
}

// AddError adds a validation error
func (r *ValidationResult) AddError(field, message, code string) *ValidationResult {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	r.Valid = false
	return r
}

// AddWarning adds a warning
func (r *ValidationResult) AddWarning(warning string) *ValidationResult {
	r.Warnings = append(r.Warnings, warning)
	return r
}
