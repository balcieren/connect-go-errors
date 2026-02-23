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
		name string
		code int
		want string
	}{
		{"canceled", 1, "connect.CodeCanceled"},
		{"unknown", 2, "connect.CodeUnknown"},
		{"invalid_argument", 3, "connect.CodeInvalidArgument"},
		{"deadline_exceeded", 4, "connect.CodeDeadlineExceeded"},
		{"not_found", 5, "connect.CodeNotFound"},
		{"already_exists", 6, "connect.CodeAlreadyExists"},
		{"permission_denied", 7, "connect.CodePermissionDenied"},
		{"resource_exhausted", 8, "connect.CodeResourceExhausted"},
		{"failed_precondition", 9, "connect.CodeFailedPrecondition"},
		{"aborted", 10, "connect.CodeAborted"},
		{"out_of_range", 11, "connect.CodeOutOfRange"},
		{"unimplemented", 12, "connect.CodeUnimplemented"},
		{"internal", 13, "connect.CodeInternal"},
		{"unavailable", 14, "connect.CodeUnavailable"},
		{"data_loss", 15, "connect.CodeDataLoss"},
		{"unauthenticated", 16, "connect.CodeUnauthenticated"},
		{"fallback_unknown", 99, "connect.CodeInternal"},
		{"fallback_zero", 0, "connect.CodeInternal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapConnectCode(tt.code)
			if got != tt.want {
				t.Errorf("mapConnectCode(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestParseErrorDef(t *testing.T) {
	// Manual protobuf wire format construction
	// ErrorDef {
	//   code (1): "ERR_TEST"
	//   message (2): "msg"
	//   connect_code (3): 5 (not_found)
	//   retryable (4): true
	// }
	data := []byte{
		0x0a, 0x08, 'E', 'R', 'R', '_', 'T', 'E', 'S', 'T', // tag 1 (string): "ERR_TEST"
		0x12, 0x03, 'm', 's', 'g', // tag 2 (string): "msg"
		0x18, 0x05, // tag 3 (varint): 5
		0x20, 0x01, // tag 4 (varint): 1
	}

	got, ok := parseErrorDef(data)
	if !ok {
		t.Fatal("parseErrorDef failed")
	}

	if got.Code != "ERR_TEST" {
		t.Errorf("Code = %q, want ERR_TEST", got.Code)
	}
	if got.Message != "msg" {
		t.Errorf("Message = %q, want msg", got.Message)
	}
	if got.ConnectCode != 5 {
		t.Errorf("ConnectCode = %d, want 5", got.ConnectCode)
	}
	if !got.Retryable {
		t.Error("Retryable = false, want true")
	}
}
