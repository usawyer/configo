package validation

import (
	"testing"
)

func TestIsValidHostnameOrIP(t *testing.T) {
	tests := []struct {
		host         string
		fieldName    string
		isAllowEmpty bool
		want         bool
		wantErr      bool
	}{
		{"localhost", "hostname", false, true, false},
		{"256.256.256.256", "IP", false, false, true},
		{"192.168.0.1", "IP", false, true, false},
		{"example.com", "hostname", false, true, false},
		{"", "hostname", true, true, false},
		{"", "hostname", false, false, true},
		{"-invalid-.com", "hostname", false, false, true},
		{"verylonghostnamepartthatexceedssixtythreecharacterswhichisinvalid.localhost", "hostname", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got, err := IsValidHostnameOrIP(tt.host, tt.fieldName, tt.isAllowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidHostnameOrIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidHostnameOrIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		port        int
		fieldName   string
		isAllowZero bool
		want        bool
		wantErr     bool
	}{
		{80, "port", false, true, false},
		{0, "port", true, true, false},
		{0, "port", false, false, true},
		{70000, "port", false, false, true},
		{-1, "port", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			got, err := IsValidPort(tt.port, tt.fieldName, tt.isAllowZero)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidValueInList(t *testing.T) {
	tests := []struct {
		value           string
		fieldName       string
		allowed         []string
		isCaseSensitive bool
		want            bool
		wantErr         bool
	}{
		{"value1", "field", []string{"value1", "value2"}, false, true, false},
		{"Value1", "field", []string{"value1", "value2"}, false, true, false},
		{"value3", "field", []string{"value1", "value2"}, false, false, true},
		{"Value1", "field", []string{"value1", "value2"}, true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsValidValueInList(tt.value, tt.fieldName, tt.allowed, tt.isCaseSensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidValueInList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidValueInList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotEmpty(t *testing.T) {
	tests := []struct {
		value     string
		fieldName string
		want      bool
		wantErr   bool
	}{
		{"nonempty", "field", true, false},
		{"", "field", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsNotEmpty(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsNotEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidStringLength(t *testing.T) {
	tests := []struct {
		value        string
		fieldName    string
		minLen       int
		maxLen       int
		isAllowEmpty bool
		want         bool
		wantErr      bool
	}{
		{"valid", "field", 1, 10, false, true, false},
		{"toolong", "field", 1, 5, false, false, true},
		{"", "field", 1, 10, true, true, false},
		{"", "field", 1, 10, false, false, true},
		{"short", "field", 10, 20, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsValidStringLength(tt.value, tt.fieldName, tt.minLen, tt.maxLen, tt.isAllowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidStringLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidStringLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		value        string
		fieldName    string
		isAllowEmpty bool
		want         bool
		wantErr      bool
	}{
		{"Alphanumeric123", "field", false, true, false},
		{"has space", "field", false, false, true},
		{"", "field", true, true, false},
		{"", "field", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsAlphanumeric(tt.value, tt.fieldName, tt.isAllowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAlphanumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsAlphanumeric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		value        string
		fieldName    string
		isAllowEmpty bool
		want         bool
		wantErr      bool
	}{
		{"https://example.com", "URL", false, true, false},
		{"ftp://example.com", "URL", false, true, false},
		{"invalid-url", "URL", false, false, true},
		{"", "URL", true, true, false},
		{"", "URL", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsValidURL(tt.value, tt.fieldName, tt.isAllowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		value        string
		fieldName    string
		isAllowEmpty bool
		want         bool
		wantErr      bool
	}{
		{"example@example.com", "email", false, true, false},
		{"invalid-email", "email", false, false, true},
		{"", "email", true, true, false},
		{"", "email", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := IsValidEmail(tt.value, tt.fieldName, tt.isAllowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPositiveInt(t *testing.T) {
	tests := []struct {
		value     int
		fieldName string
		want      bool
		wantErr   bool
	}{
		{10, "number", true, false},
		{0, "number", false, true},
		{-1, "number", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			got, err := IsPositiveInt(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPositiveInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsPositiveInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsValidCronExpression тестирует функцию IsValidCronExpression.
func TestIsValidCronExpression(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		expression    string
		isWithSeconds bool
		expectedValid bool
		expectedErr   string
	}{
		{
			name:          "Valid with seconds",
			fieldName:     "DataTransfer",
			expression:    "*/20 * * * * *",
			isWithSeconds: true,
			expectedValid: true,
			expectedErr:   "",
		},
		{
			name:          "Valid without seconds",
			fieldName:     "ConfigUpdate",
			expression:    "0 * * * *",
			isWithSeconds: false,
			expectedValid: true,
			expectedErr:   "",
		},
		{
			name:          "Invalid with extra field",
			fieldName:     "DataBackup",
			expression:    "*/30 * * * * * *",
			isWithSeconds: true,
			expectedValid: false,
			expectedErr:   "DataBackup cron выражение '*/30 * * * * * *' недействительно:",
		},
		{
			name:          "Invalid without seconds",
			fieldName:     "ReportGeneration",
			expression:    "0 * * * * *",
			isWithSeconds: false,
			expectedValid: false,
			expectedErr:   "ReportGeneration cron выражение '0 * * * * *' недействительно:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, err := IsValidCronExpression(test.expression, test.fieldName, test.isWithSeconds)
			if valid != test.expectedValid {
				t.Errorf("expected validity %v, got %v", test.expectedValid, valid)
			}
			if err != nil && test.expectedErr == "" {
				t.Errorf("expected no error, got error %v", err)
			}
			if err == nil && test.expectedErr != "" {
				t.Errorf("expected error %s, got no error", test.expectedErr)
			}
			if err != nil && test.expectedErr != "" && !startsWith(err.Error(), test.expectedErr) {
				t.Errorf("expected error message to start with '%s', got '%s'", test.expectedErr, err.Error())
			}
		})
	}
}

// startsWith checks if a string starts with a given prefix.
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
