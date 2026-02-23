package main

import (
	"reflect"
	"testing"
)

func TestExtractTemplateFields(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    []string
	}{
		{"no placeholders", "Internal server error", nil},
		{"single field", "User '{{id}}' not found", []string{"id"}},
		{"multiple fields", "User '{{id}}' not found in '{{org}}'", []string{"id", "org"}},
		{"duplicate fields", "{{id}} and {{id}} again", []string{"id"}},
		{"snake_case field", "Product {{product_id}} unavailable", []string{"product_id"}},
		{"empty message", "", nil},
		{"adjacent placeholders", "{{a}}{{b}}", []string{"a", "b"}},
		{"unclosed placeholder", "Hello {{name", nil},
		{"three fields", "{{amount}} exceeds {{limit}} for {{account}}", []string{"amount", "limit", "account"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTemplateFields(tt.message)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTemplateFields(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestFieldToExportedName(t *testing.T) {
	tests := []struct {
		field string
		want  string
	}{
		{"id", "Id"},
		{"email", "Email"},
		{"product_id", "ProductId"},
		{"order_item_id", "OrderItemId"},
		{"reason", "Reason"},
		{"unlock_at", "UnlockAt"},
		{"last4", "Last4"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := fieldToExportedName(tt.field)
			if got != tt.want {
				t.Errorf("fieldToExportedName(%q) = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestErrorCodeToConstant(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"ERROR_NOT_FOUND", "NotFound"},
		{"ERROR_INVALID_ARGUMENT", "InvalidArgument"},
		{"ERROR_USER_NOT_FOUND", "UserNotFound"},
		{"ERROR_INTERNAL", "Internal"},
		{"ERROR_OUT_OF_STOCK", "OutOfStock"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := errorCodeToConstant(tt.code)
			if got != tt.want {
				t.Errorf("errorCodeToConstant(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestMapConnectCode(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"not_found", "connect.CodeNotFound"},
		{"internal", "connect.CodeInternal"},
		{"invalid_argument", "connect.CodeInvalidArgument"},
		{"unauthenticated", "connect.CodeUnauthenticated"},
		{"unknown_code", "connect.CodeInternal"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := mapConnectCode(tt.code)
			if got != tt.want {
				t.Errorf("mapConnectCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
