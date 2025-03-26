package utils

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// EncryptStructPII uses reflection to encrypt fields tagged with pii:"true"
func EncryptStructPII(data interface{}, key string) error {
	// Get the value of the struct
	val := reflect.ValueOf(data)

	// Ensure we're working with a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to a struct")
	}

	// Get the actual struct
	val = val.Elem()

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Check if the field is tagged as PII
		if tag, ok := fieldType.Tag.Lookup("pii"); ok && tag == "true" {
			// Perform encryption on PII fields
			encryptedValue, err := Encrypt(field.String(), key)
			if err != nil {
				log.Printf("Failed to encrypt %s: %v", fieldType.Name, err)
				return err
			}
			field.SetString(encryptedValue)
		}
		// If it's a nested struct, recursively encrypt it
		if field.Kind() == reflect.Struct {
			err := EncryptStructPII(field.Addr().Interface(), key)
			if err != nil {

				return err
			}
		}

		// Handle slices or arrays of structs
		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			// Check if the slice contains structs
			if field.Type().Elem().Kind() == reflect.Struct {
				for j := 0; j < field.Len(); j++ {
					// Get the element as a struct pointer and encrypt it
					elem := field.Index(j).Addr().Interface()
					err := EncryptStructPII(elem, key)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// DecryptStructPII uses reflection to decrypt fields tagged with pii:"true"
func DecryptStructPII(data interface{}, key string) error {
	val := reflect.ValueOf(data).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Check if the field is tagged as PII
		if tag, ok := fieldType.Tag.Lookup("pii"); ok && tag == "true" {
			// Perform decryption on PII fields
			decryptedValue, err := Decrypt(field.String(), key)
			if err != nil {
				log.Printf("Failed to decrypt %s: %v", fieldType.Name, err)
				return err
			}
			field.SetString(decryptedValue)
		}

		// If it's a nested struct, recursively encrypt it
		if field.Kind() == reflect.Struct {
			err := DecryptStructPII(field.Addr().Interface(), key)
			if err != nil {

				return err
			}
		}

		// Handle slices or arrays of structs
		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			// Check if the slice contains structs
			if field.Type().Elem().Kind() == reflect.Struct {
				for j := 0; j < field.Len(); j++ {
					// Get the element as a struct pointer and encrypt it
					elem := field.Index(j).Addr().Interface()
					err := DecryptStructPII(elem, key)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// RedactPII redacts fields based on role
func RedactPII(v interface{}, role string) {
	val := reflect.ValueOf(v).Elem() // Get the value of the struct
	typ := val.Type()                // Get the type of the struct

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if the field has the `pii:"true"` tag
		if tag := fieldType.Tag.Get("pii"); tag == "true" {
			if shouldRedact(role) {
				// Redact the field (if it's a string)
				if field.Kind() == reflect.String {
					field.SetString("REDACTED")
				}
			}
		}

		// Handle nested structs (like Address)
		if field.Kind() == reflect.Struct {
			RedactPII(field.Addr().Interface(), role) // Recursively redact nested structs
		}
	}
}

// shouldRedact returns true if the given role should not have access to PII
func shouldRedact(role string) bool {
	// Define roles that can view PII (e.g., "admin" or "manager")
	allowedRoles := map[string]bool{
		"admin":   true,
		"manager": true,
	}

	// If the role is not in the allowed list, redact the field
	_, canViewPII := allowedRoles[role]
	return !canViewPII
}

// SanitizeCustomerData sanitizes PII based on role
func SanitizeCustomerData(v interface{}, role string) {
	val := reflect.ValueOf(v).Elem() // Get the value of the struct
	typ := val.Type()                // Get the type of the struct

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Check if the field has the `pii:"true"` tag
		if tag := fieldType.Tag.Get("pii"); tag == "true" {
			if shouldSanitize(role, fieldType.Name) {
				// Sanitize the field (if it's a string)
				if field.Kind() == reflect.String {
					field.SetString(maskPII(field.String(), fieldType.Name))
				}
			}
		}

		// Handle nested structs (like Address)
		if field.Kind() == reflect.Struct {
			SanitizeCustomerData(field.Addr().Interface(), role) // Recursively sanitize nested structs
		}
	}
}

// shouldSanitize checks if a field should be sanitized based on the role
func shouldSanitize(role, fieldName string) bool {
	// Define roles that have full access to specific fields
	if role == "admin" {
		return false // Admins can view all PII
	}

	if role == "manager" && (fieldName == "Email" || fieldName == "Phone") {
		return false // Manager can view email, phone but not others
	}

	// Default case: sanitize for viewers and others
	return true
}

// maskPII returns a sanitized version of the PII field
func maskPII(value string, fieldName string) string {
	switch fieldName {
	case "Email":
		// Mask email address
		return maskEmail(value)
	case "Phone":
		// Mask phone number
		if len(value) >= 10 {
			return "*******" + value[len(value)-3:]
		}
	case "DocNumber":
		// Mask Doc number
		if len(value) >= 4 {
			return "***-**-" + value[len(value)-4:]
		}
	case "DOB":
		// Mask DOB (show only the year)
		parts := strings.Split(value, "-")
		if len(parts) >= 1 {
			return "**-**-" + parts[len(parts)-1]
		}
		return "****-**-**"
	}
	return "REDACTED"
}

// maskEmail masks part of the email address.
func maskEmail(email string) string {
	if len(email) == 0 {
		return "REDACTED"
	}
	// Split the email into the local part (before @) and the domain part (after @).
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "REDACTED"
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Mask all but the first character of the local part.
	if len(localPart) > 1 {
		localPart = string(localPart[0]) + strings.Repeat("*", len(localPart)-1)
	} else {
		localPart = "*"
	}

	return localPart + "@" + domainPart
}
