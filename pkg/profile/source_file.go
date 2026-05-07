// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
)

// FileProfileSource represents a profile source that reads from a local JSON file.
type FileProfileSource struct {
	path string
}

// NewFileProfileSource creates a new file-based profile source.
func NewFileProfileSource(path string) *FileProfileSource {
	return &FileProfileSource{path: path}
}

// GetSpec reads and returns the JSON specification from the file.
func (f *FileProfileSource) GetSpec(_ context.Context) ([]byte, error) {
	expandedPath := os.ExpandEnv(f.path)
	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("read profile file %q: %w", f.path, err)
	}

	// Validate it's valid JSON representing a PodSpec
	var podSpec corev1.PodSpec
	if err := json.Unmarshal(data, &podSpec); err != nil {
		return nil, fmt.Errorf("invalid JSON PodSpec in %q: %w", f.path, err)
	}

	return data, nil
}

// Type returns the source type identifier.
func (f *FileProfileSource) Type() string {
	return SourceTypeFile
}
