// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileProfileSource_GetSpec(t *testing.T) {
	t.Parallel()

	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Invalid JSON
	invalidJSON := `{
		"volumeMounts": [
			{
				"mountPath": "/app/config"
		]
	}`

	// Not a PodSpec (valid JSON but wrong structure)
	wrongStructure := `{
		"foo": "bar",
		"baz": 123
	}`

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // Returns path to test file
		wantErr     bool
		errContains string
	}{
		{
			name: "valid profile file",
			setup: func(t *testing.T) string {
				t.Helper()
				path := filepath.Join(tmpDir, "valid.json")
				if err := os.WriteFile(path, []byte(testValidProfile), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr: false,
		},
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			wantErr:     true,
			errContains: "read profile file",
		},
		{
			name: "invalid JSON",
			setup: func(t *testing.T) string {
				t.Helper()
				path := filepath.Join(tmpDir, "invalid.json")
				if err := os.WriteFile(path, []byte(invalidJSON), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr:     true,
			errContains: "invalid JSON PodSpec",
		},
		{
			name: "wrong structure",
			setup: func(t *testing.T) string {
				t.Helper()
				path := filepath.Join(tmpDir, "wrong.json")
				if err := os.WriteFile(path, []byte(wrongStructure), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr: false, // PodSpec unmarshaling is lenient - extra fields are ignored
		},
		{
			name: "file with environment variable in path",
			setup: func(t *testing.T) string {
				t.Helper()
				// Set env var for test
				testDir := tmpDir
				if err := os.Setenv("TEST_PROFILE_DIR", testDir); err != nil {
					t.Fatalf("failed to set env var: %v", err)
				}
				path := filepath.Join(testDir, "env.json")
				if err := os.WriteFile(path, []byte(testValidProfile), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return "${TEST_PROFILE_DIR}/env.json"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setup(t)
			source := NewFileProfileSource(path)

			spec, err := source.GetSpec(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("FileProfileSource.GetSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("FileProfileSource.GetSpec() error = %v, want error containing %q", err, tt.errContains)
				}
			}

			if !tt.wantErr && spec == nil {
				t.Error("FileProfileSource.GetSpec() returned nil spec, want non-nil")
			}
		})
	}
}

func TestFileProfileSource_Type(t *testing.T) {
	t.Parallel()

	source := NewFileProfileSource("/path/to/profile.json")
	if got := source.Type(); got != SourceTypeFile {
		t.Errorf("FileProfileSource.Type() = %v, want %v", got, SourceTypeFile)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
