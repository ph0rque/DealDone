package webhooks

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// WebhookSchemaValidator provides JSON schema validation for webhook payloads
type WebhookSchemaValidator struct {
	schemas      map[string]*JSONSchema
	apiVersion   APIVersion
	strictMode   bool
	cacheEnabled bool
}

// JSONSchema represents a JSON schema definition
type JSONSchema struct {
	SchemaName           string               `json:"$schema"`
	Type                 string               `json:"type"`
	Title                string               `json:"title"`
	Description          string               `json:"description"`
	Version              string               `json:"version"`
	Properties           map[string]*Property `json:"properties"`
	Required             []string             `json:"required"`
	AdditionalProperties bool                 `json:"additionalProperties"`
	Examples             []interface{}        `json:"examples,omitempty"`
	LastUpdated          int64                `json:"lastUpdated"`
}

// Property represents a property in a JSON schema
type Property struct {
	Type        string               `json:"type"`
	Format      string               `json:"format,omitempty"`
	Description string               `json:"description"`
	Required    bool                 `json:"required,omitempty"`
	Minimum     *float64             `json:"minimum,omitempty"`
	Maximum     *float64             `json:"maximum,omitempty"`
	MinLength   *int                 `json:"minLength,omitempty"`
	MaxLength   *int                 `json:"maxLength,omitempty"`
	Pattern     string               `json:"pattern,omitempty"`
	Enum        []interface{}        `json:"enum,omitempty"`
	Items       *Property            `json:"items,omitempty"`
	Properties  map[string]*Property `json:"properties,omitempty"`
	OneOf       []*Property          `json:"oneOf,omitempty"`
	AnyOf       []*Property          `json:"anyOf,omitempty"`
	Examples    []interface{}        `json:"examples,omitempty"`
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Path     string      `json:"path"`
	Field    string      `json:"field"`
	Value    interface{} `json:"value"`
	Expected string      `json:"expected"`
	Message  string      `json:"message"`
	Code     string      `json:"code"`
	Severity string      `json:"severity"` // "error", "warning", "info"
}

// SchemaValidationResult represents the result of schema validation
type SchemaValidationResult struct {
	Valid          bool              `json:"valid"`
	Errors         []ValidationError `json:"errors,omitempty"`
	Warnings       []ValidationError `json:"warnings,omitempty"`
	SchemaUsed     string            `json:"schemaUsed"`
	ValidationTime int64             `json:"validationTimeMs"`
	PayloadSize    int               `json:"payloadSize"`
	SchemaVersion  string            `json:"schemaVersion"`
}

// NewWebhookSchemaValidator creates a new schema validator
func NewWebhookSchemaValidator() *WebhookSchemaValidator {
	validator := &WebhookSchemaValidator{
		schemas: make(map[string]*JSONSchema),
		apiVersion: APIVersion{
			Major: 1,
			Minor: 1,
			Patch: 0,
			Label: "stable",
		},
		strictMode:   true,
		cacheEnabled: true,
	}

	// Initialize built-in schemas
	validator.initializeSchemas()

	return validator
}

// ValidatePayload validates a payload against its appropriate schema
func (v *WebhookSchemaValidator) ValidatePayload(payload interface{}, schemaName string) (*SchemaValidationResult, error) {
	start := time.Now()

	schema, exists := v.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaName)
	}

	// Convert payload to map for validation
	var payloadMap map[string]interface{}

	switch p := payload.(type) {
	case map[string]interface{}:
		payloadMap = p
	case []byte:
		if err := json.Unmarshal(p, &payloadMap); err != nil {
			return &SchemaValidationResult{
				Valid: false,
				Errors: []ValidationError{{
					Path:     "root",
					Field:    "payload",
					Message:  fmt.Sprintf("Invalid JSON: %v", err),
					Code:     "JSON_PARSE_ERROR",
					Severity: "error",
				}},
				ValidationTime: time.Since(start).Milliseconds(),
			}, err
		}
	case string:
		if err := json.Unmarshal([]byte(p), &payloadMap); err != nil {
			return &SchemaValidationResult{
				Valid: false,
				Errors: []ValidationError{{
					Path:     "root",
					Field:    "payload",
					Message:  fmt.Sprintf("Invalid JSON: %v", err),
					Code:     "JSON_PARSE_ERROR",
					Severity: "error",
				}},
				ValidationTime: time.Since(start).Milliseconds(),
			}, err
		}
	default:
		// Convert struct to map via JSON marshaling
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &payloadMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	// Perform validation
	result := &SchemaValidationResult{
		SchemaUsed:     schemaName,
		SchemaVersion:  schema.Version,
		ValidationTime: 0,
		PayloadSize:    len(fmt.Sprintf("%v", payload)),
		Valid:          true,
	}

	errors, warnings := v.validateObject(payloadMap, schema, "")
	result.Errors = errors
	result.Warnings = warnings
	result.Valid = len(errors) == 0
	result.ValidationTime = time.Since(start).Milliseconds()

	return result, nil
}

// validateObject validates an object against a schema
func (v *WebhookSchemaValidator) validateObject(obj map[string]interface{}, schema *JSONSchema, path string) ([]ValidationError, []ValidationError) {
	var errors []ValidationError
	var warnings []ValidationError

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, exists := obj[requiredField]; !exists {
			errors = append(errors, ValidationError{
				Path:     v.buildPath(path, requiredField),
				Field:    requiredField,
				Message:  fmt.Sprintf("Required field '%s' is missing", requiredField),
				Code:     "REQUIRED_FIELD_MISSING",
				Severity: "error",
			})
		}
	}

	// Validate each property
	for fieldName, value := range obj {
		fieldPath := v.buildPath(path, fieldName)

		// Check if field is defined in schema
		property, exists := schema.Properties[fieldName]
		if !exists {
			if !schema.AdditionalProperties {
				if v.strictMode {
					errors = append(errors, ValidationError{
						Path:     fieldPath,
						Field:    fieldName,
						Value:    value,
						Message:  fmt.Sprintf("Additional property '%s' is not allowed", fieldName),
						Code:     "ADDITIONAL_PROPERTY_NOT_ALLOWED",
						Severity: "error",
					})
				} else {
					warnings = append(warnings, ValidationError{
						Path:     fieldPath,
						Field:    fieldName,
						Value:    value,
						Message:  fmt.Sprintf("Unknown property '%s'", fieldName),
						Code:     "UNKNOWN_PROPERTY",
						Severity: "warning",
					})
				}
			}
			continue
		}

		// Validate property value
		propErrors, propWarnings := v.validateProperty(fieldName, value, property, fieldPath)
		errors = append(errors, propErrors...)
		warnings = append(warnings, propWarnings...)
	}

	return errors, warnings
}

// validateProperty validates a single property value
func (v *WebhookSchemaValidator) validateProperty(name string, value interface{}, property *Property, path string) ([]ValidationError, []ValidationError) {
	var errors []ValidationError
	var warnings []ValidationError

	// Check type
	actualType := v.getJSONType(value)
	expectedType := property.Type

	if expectedType != "" && actualType != expectedType {
		// Handle special cases
		if !(expectedType == "number" && actualType == "integer") {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: expectedType,
				Message:  fmt.Sprintf("Expected type '%s' but got '%s'", expectedType, actualType),
				Code:     "TYPE_MISMATCH",
				Severity: "error",
			})
			return errors, warnings
		}
	}

	// Type-specific validations
	switch actualType {
	case "string":
		strValue := value.(string)

		// Length validation
		if property.MinLength != nil && len(strValue) < *property.MinLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("minimum length %d", *property.MinLength),
				Message:  fmt.Sprintf("String too short: %d < %d", len(strValue), *property.MinLength),
				Code:     "STRING_TOO_SHORT",
				Severity: "error",
			})
		}

		if property.MaxLength != nil && len(strValue) > *property.MaxLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("maximum length %d", *property.MaxLength),
				Message:  fmt.Sprintf("String too long: %d > %d", len(strValue), *property.MaxLength),
				Code:     "STRING_TOO_LONG",
				Severity: "error",
			})
		}

		// Pattern validation
		if property.Pattern != "" {
			matched, err := regexp.MatchString(property.Pattern, strValue)
			if err != nil {
				warnings = append(warnings, ValidationError{
					Path:     path,
					Field:    name,
					Value:    value,
					Message:  fmt.Sprintf("Invalid regex pattern: %v", err),
					Code:     "INVALID_PATTERN",
					Severity: "warning",
				})
			} else if !matched {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    name,
					Value:    value,
					Expected: fmt.Sprintf("pattern: %s", property.Pattern),
					Message:  fmt.Sprintf("String does not match pattern: %s", property.Pattern),
					Code:     "PATTERN_MISMATCH",
					Severity: "error",
				})
			}
		}

		// Format validation
		if property.Format != "" {
			if !v.validateFormat(strValue, property.Format) {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    name,
					Value:    value,
					Expected: fmt.Sprintf("format: %s", property.Format),
					Message:  fmt.Sprintf("Invalid format '%s'", property.Format),
					Code:     "INVALID_FORMAT",
					Severity: "error",
				})
			}
		}

	case "number", "integer":
		var numValue float64
		switch v := value.(type) {
		case int:
			numValue = float64(v)
		case int64:
			numValue = float64(v)
		case float64:
			numValue = v
		case float32:
			numValue = float64(v)
		}

		// Range validation
		if property.Minimum != nil && numValue < *property.Minimum {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("minimum %f", *property.Minimum),
				Message:  fmt.Sprintf("Number too small: %f < %f", numValue, *property.Minimum),
				Code:     "NUMBER_TOO_SMALL",
				Severity: "error",
			})
		}

		if property.Maximum != nil && numValue > *property.Maximum {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("maximum %f", *property.Maximum),
				Message:  fmt.Sprintf("Number too large: %f > %f", numValue, *property.Maximum),
				Code:     "NUMBER_TOO_LARGE",
				Severity: "error",
			})
		}

	case "array":
		arrayValue := value.([]interface{})

		// Array length validation
		if property.MinLength != nil && len(arrayValue) < *property.MinLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("minimum length %d", *property.MinLength),
				Message:  fmt.Sprintf("Array too short: %d < %d", len(arrayValue), *property.MinLength),
				Code:     "ARRAY_TOO_SHORT",
				Severity: "error",
			})
		}

		if property.MaxLength != nil && len(arrayValue) > *property.MaxLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("maximum length %d", *property.MaxLength),
				Message:  fmt.Sprintf("Array too long: %d > %d", len(arrayValue), *property.MaxLength),
				Code:     "ARRAY_TOO_LONG",
				Severity: "error",
			})
		}

		// Validate array items
		if property.Items != nil {
			for i, item := range arrayValue {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				itemErrors, itemWarnings := v.validateProperty(fmt.Sprintf("%s[%d]", name, i), item, property.Items, itemPath)
				errors = append(errors, itemErrors...)
				warnings = append(warnings, itemWarnings...)
			}
		}

	case "object":
		if objValue, ok := value.(map[string]interface{}); ok && property.Properties != nil {
			// Create a mini-schema for object validation
			objSchema := &JSONSchema{
				Properties:           property.Properties,
				Required:             []string{}, // Required fields would be specified in the property definition
				AdditionalProperties: true,
			}

			objErrors, objWarnings := v.validateObject(objValue, objSchema, path)
			errors = append(errors, objErrors...)
			warnings = append(warnings, objWarnings...)
		}
	}

	// Enum validation
	if len(property.Enum) > 0 {
		found := false
		for _, enumValue := range property.Enum {
			if reflect.DeepEqual(value, enumValue) {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    name,
				Value:    value,
				Expected: fmt.Sprintf("one of: %v", property.Enum),
				Message:  fmt.Sprintf("Value not in enum: %v", property.Enum),
				Code:     "VALUE_NOT_IN_ENUM",
				Severity: "error",
			})
		}
	}

	return errors, warnings
}

// validateFormat validates string formats
func (v *WebhookSchemaValidator) validateFormat(value, format string) bool {
	switch format {
	case "email":
		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, _ := regexp.MatchString(emailPattern, value)
		return matched
	case "uri", "url":
		return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "ftp://")
	case "date-time":
		_, err := time.Parse(time.RFC3339, value)
		return err == nil
	case "date":
		_, err := time.Parse("2006-01-02", value)
		return err == nil
	case "time":
		_, err := time.Parse("15:04:05", value)
		return err == nil
	case "uuid":
		uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
		matched, _ := regexp.MatchString(uuidPattern, strings.ToLower(value))
		return matched
	default:
		return true // Unknown format, assume valid
	}
}

// getJSONType returns the JSON type of a value
func (v *WebhookSchemaValidator) getJSONType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// buildPath builds a JSON path string
func (v *WebhookSchemaValidator) buildPath(basePath, field string) string {
	if basePath == "" {
		return field
	}
	return fmt.Sprintf("%s.%s", basePath, field)
}

// GetSchema returns a schema by name
func (v *WebhookSchemaValidator) GetSchema(name string) (*JSONSchema, error) {
	if schema, exists := v.schemas[name]; exists {
		return schema, nil
	}
	return nil, fmt.Errorf("schema not found: %s", name)
}

// AddSchema adds a custom schema
func (v *WebhookSchemaValidator) AddSchema(name string, schema *JSONSchema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	schema.LastUpdated = time.Now().UnixMilli()
	v.schemas[name] = schema
	return nil
}

// ListSchemas returns all available schema names
func (v *WebhookSchemaValidator) ListSchemas() []string {
	var names []string
	for name := range v.schemas {
		names = append(names, name)
	}
	return names
}

// GetSchemaInfo returns information about a schema
func (v *WebhookSchemaValidator) GetSchemaInfo(name string) (*JSONSchemaInfo, error) {
	schema, exists := v.schemas[name]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", name)
	}

	return &JSONSchemaInfo{
		SchemaName:    name,
		SchemaVersion: schema.Version,
		Description:   schema.Description,
		LastUpdated:   schema.LastUpdated,
	}, nil
}

// SetStrictMode enables or disables strict validation mode
func (v *WebhookSchemaValidator) SetStrictMode(strict bool) {
	v.strictMode = strict
}

// GetAPIVersion returns the current API version
func (v *WebhookSchemaValidator) GetAPIVersion() APIVersion {
	return v.apiVersion
}

// GetCompatibilityInfo returns API compatibility information
func (v *WebhookSchemaValidator) GetCompatibilityInfo() CompatibilityInfo {
	return CompatibilityInfo{
		CurrentVersion: v.apiVersion,
		SupportedVersions: []APIVersion{
			{Major: 1, Minor: 0, Patch: 0, Label: "stable"},
			{Major: 1, Minor: 1, Patch: 0, Label: "stable"},
		},
		DeprecatedVersions: []APIVersion{
			{Major: 0, Minor: 9, Patch: 0, Label: "beta"},
		},
		MinimumVersion: APIVersion{Major: 1, Minor: 0, Patch: 0},
	}
}

// initializeSchemas initializes built-in schemas for all webhook payload types
func (v *WebhookSchemaValidator) initializeSchemas() {
	// Document Webhook Payload Schema
	v.schemas["document-webhook-payload"] = &JSONSchema{
		SchemaName:  "document-webhook-payload",
		Type:        "object",
		Title:       "Document Webhook Payload",
		Description: "Schema for document analysis webhook payloads",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"dealName": {
				Type:        "string",
				Description: "Name of the deal",
				MinLength:   &[]int{1}[0],
				MaxLength:   &[]int{255}[0],
			},
			"filePaths": {
				Type:        "array",
				Description: "Array of file paths to process",
				MinLength:   &[]int{1}[0],
				Items: &Property{
					Type: "string",
				},
			},
			"triggerType": {
				Type:        "string",
				Description: "Type of trigger that initiated the processing",
				Enum:        []interface{}{"file_change", "user_button", "analyze_all", "scheduled", "retry", "user_correction"},
			},
			"workflowType": {
				Type:        "string",
				Description: "Type of workflow to execute",
				Enum:        []interface{}{"document-analysis", "error-handling", "user-corrections", "cleanup", "batch-processing", "health-check"},
			},
			"jobId": {
				Type:        "string",
				Description: "Unique job identifier",
				Pattern:     "^job_[0-9]+_[a-zA-Z0-9_-]+$",
			},
			"priority": {
				Type:        "integer",
				Description: "Processing priority (1=high, 2=normal, 3=low)",
				Minimum:     &[]float64{1}[0],
				Maximum:     &[]float64{3}[0],
			},
			"timestamp": {
				Type:        "integer",
				Description: "Unix timestamp in milliseconds",
				Minimum:     &[]float64{0}[0],
			},
			"retryCount": {
				Type:        "integer",
				Description: "Number of retry attempts",
				Minimum:     &[]float64{0}[0],
			},
			"maxRetries": {
				Type:        "integer",
				Description: "Maximum number of retries allowed",
				Minimum:     &[]float64{0}[0],
				Maximum:     &[]float64{10}[0],
			},
			"timeoutSeconds": {
				Type:        "integer",
				Description: "Timeout in seconds",
				Minimum:     &[]float64{1}[0],
				Maximum:     &[]float64{3600}[0],
			},
		},
		Required:             []string{"dealName", "filePaths", "triggerType", "workflowType", "jobId", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}

	// Webhook Result Payload Schema
	v.schemas["webhook-result-payload"] = &JSONSchema{
		SchemaName:  "webhook-result-payload",
		Type:        "object",
		Title:       "Webhook Result Payload",
		Description: "Schema for webhook result payloads from n8n",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"jobId": {
				Type:        "string",
				Description: "Job identifier",
				Pattern:     "^job_[0-9]+_[a-zA-Z0-9_-]+$",
			},
			"dealName": {
				Type:        "string",
				Description: "Name of the deal",
				MinLength:   &[]int{1}[0],
			},
			"workflowType": {
				Type:        "string",
				Description: "Type of workflow executed",
				Enum:        []interface{}{"document-analysis", "error-handling", "user-corrections", "cleanup", "batch-processing", "health-check"},
			},
			"status": {
				Type:        "string",
				Description: "Processing status",
				Enum:        []interface{}{"completed", "failed", "partial_success", "in_progress"},
			},
			"processedDocuments": {
				Type:        "integer",
				Description: "Number of documents processed",
				Minimum:     &[]float64{0}[0],
			},
			"totalDocuments": {
				Type:        "integer",
				Description: "Total number of documents",
				Minimum:     &[]float64{0}[0],
			},
			"averageConfidence": {
				Type:        "number",
				Description: "Average confidence score",
				Minimum:     &[]float64{0}[0],
				Maximum:     &[]float64{1}[0],
			},
			"processingTimeMs": {
				Type:        "integer",
				Description: "Processing time in milliseconds",
				Minimum:     &[]float64{0}[0],
			},
			"startTime": {
				Type:        "integer",
				Description: "Start time timestamp",
				Minimum:     &[]float64{0}[0],
			},
			"timestamp": {
				Type:        "integer",
				Description: "Result timestamp",
				Minimum:     &[]float64{0}[0],
			},
		},
		Required:             []string{"jobId", "dealName", "workflowType", "status", "startTime", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}

	// Error Handling Payload Schema
	v.schemas["error-handling-payload"] = &JSONSchema{
		SchemaName:  "error-handling-payload",
		Type:        "object",
		Title:       "Error Handling Payload",
		Description: "Schema for error handling webhook payloads",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"originalJobId": {
				Type:        "string",
				Description: "Original job ID that failed",
				Pattern:     "^job_[0-9]+_[a-zA-Z0-9_-]+$",
			},
			"errorJobId": {
				Type:        "string",
				Description: "Error handling job ID",
				Pattern:     "^job_[0-9]+_[a-zA-Z0-9_-]+$",
			},
			"dealName": {
				Type:        "string",
				Description: "Deal name",
				MinLength:   &[]int{1}[0],
			},
			"errorType": {
				Type:        "string",
				Description: "Type of error",
				MinLength:   &[]int{1}[0],
			},
			"retryAttempt": {
				Type:        "integer",
				Description: "Current retry attempt",
				Minimum:     &[]float64{1}[0],
			},
			"maxRetries": {
				Type:        "integer",
				Description: "Maximum retries allowed",
				Minimum:     &[]float64{1}[0],
			},
			"retryStrategy": {
				Type:        "string",
				Description: "Retry strategy",
				Enum:        []interface{}{"immediate", "exponential", "scheduled"},
			},
			"recoveryAction": {
				Type:        "string",
				Description: "Recovery action to take",
				Enum:        []interface{}{"retry", "skip", "manual"},
			},
			"timestamp": {
				Type:        "integer",
				Description: "Timestamp",
				Minimum:     &[]float64{0}[0],
			},
		},
		Required:             []string{"originalJobId", "errorJobId", "dealName", "errorType", "retryAttempt", "maxRetries", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}

	// User Correction Payload Schema
	v.schemas["user-correction-payload"] = &JSONSchema{
		SchemaName:  "user-correction-payload",
		Type:        "object",
		Title:       "User Correction Payload",
		Description: "Schema for user correction webhook payloads",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"correctionId": {
				Type:        "string",
				Description: "Unique correction identifier",
				MinLength:   &[]int{1}[0],
			},
			"originalJobId": {
				Type:        "string",
				Description: "Original job ID",
				Pattern:     "^job_[0-9]+_[a-zA-Z0-9_-]+$",
			},
			"dealName": {
				Type:        "string",
				Description: "Deal name",
				MinLength:   &[]int{1}[0],
			},
			"templatePath": {
				Type:        "string",
				Description: "Path to template being corrected",
				MinLength:   &[]int{1}[0],
			},
			"corrections": {
				Type:        "array",
				Description: "Array of field corrections",
				MinLength:   &[]int{1}[0],
			},
			"correctionType": {
				Type:        "string",
				Description: "Type of correction",
				Enum:        []interface{}{"manual", "assisted", "bulk"},
			},
			"applyToSimilar": {
				Type:        "boolean",
				Description: "Whether to apply learning to similar documents",
			},
			"confidence": {
				Type:        "number",
				Description: "Overall confidence in corrections",
				Minimum:     &[]float64{0}[0],
				Maximum:     &[]float64{1}[0],
			},
			"timestamp": {
				Type:        "integer",
				Description: "Timestamp",
				Minimum:     &[]float64{0}[0],
			},
		},
		Required:             []string{"correctionId", "originalJobId", "dealName", "templatePath", "corrections", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}

	// Batch Processing Payload Schema
	v.schemas["batch-processing-payload"] = &JSONSchema{
		SchemaName:  "batch-processing-payload",
		Type:        "object",
		Title:       "Batch Processing Payload",
		Description: "Schema for batch processing webhook payloads",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"batchId": {
				Type:        "string",
				Description: "Unique batch identifier",
				MinLength:   &[]int{1}[0],
			},
			"dealName": {
				Type:        "string",
				Description: "Deal name",
				MinLength:   &[]int{1}[0],
			},
			"batchType": {
				Type:        "string",
				Description: "Type of batch processing",
				Enum:        []interface{}{"deal_analysis", "template_update", "bulk_correction", "cleanup"},
			},
			"items": {
				Type:        "array",
				Description: "Items to process in batch",
				MinLength:   &[]int{1}[0],
			},
			"priority": {
				Type:        "integer",
				Description: "Batch priority",
				Minimum:     &[]float64{1}[0],
				Maximum:     &[]float64{3}[0],
			},
			"timestamp": {
				Type:        "integer",
				Description: "Timestamp",
				Minimum:     &[]float64{0}[0],
			},
		},
		Required:             []string{"batchId", "dealName", "batchType", "items", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}

	// Health Check Payload Schema
	v.schemas["health-check-payload"] = &JSONSchema{
		SchemaName:  "health-check-payload",
		Type:        "object",
		Title:       "Health Check Payload",
		Description: "Schema for health check webhook payloads",
		Version:     "1.1.0",
		Properties: map[string]*Property{
			"checkId": {
				Type:        "string",
				Description: "Unique check identifier",
				MinLength:   &[]int{1}[0],
			},
			"checkType": {
				Type:        "string",
				Description: "Type of health check",
				Enum:        []interface{}{"system", "component", "workflow", "end_to_end"},
			},
			"components": {
				Type:        "array",
				Description: "Components to check",
				Items: &Property{
					Type: "string",
				},
			},
			"timestamp": {
				Type:        "integer",
				Description: "Timestamp",
				Minimum:     &[]float64{0}[0],
			},
		},
		Required:             []string{"checkId", "checkType", "timestamp"},
		AdditionalProperties: true,
		LastUpdated:          time.Now().UnixMilli(),
	}
}
