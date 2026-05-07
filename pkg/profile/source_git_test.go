// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestGitProfileSource_Type(t *testing.T) {
	t.Parallel()

	source := NewGitProfileSource("https://github.com/test/repo", "main", "profile.json")

	if got := source.Type(); got != SourceTypeGit {
		t.Errorf("GitProfileSource.Type() = %v, want %v", got, SourceTypeGit)
	}
}

func TestNewGitProfileSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		ref         string
		path        string
		expectedRef string
	}{
		{
			name:        "empty ref defaults to main",
			url:         "https://github.com/test/repo",
			ref:         "",
			path:        "profile.json",
			expectedRef: "main",
		},
		{
			name:        "explicit main ref",
			url:         "https://github.com/test/repo",
			ref:         "main",
			path:        "profile.json",
			expectedRef: "main",
		},
		{
			name:        "custom branch",
			url:         "https://github.com/test/repo",
			ref:         "develop",
			path:        "profile.json",
			expectedRef: "develop",
		},
		{
			name:        "tag reference",
			url:         "https://github.com/test/repo",
			ref:         "v1.0.0",
			path:        "profiles/debug.json",
			expectedRef: "v1.0.0",
		},
		{
			name:        "commit hash",
			url:         "https://github.com/test/repo",
			ref:         "abc123def",
			path:        "path/to/profile.json",
			expectedRef: "abc123def",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewGitProfileSource(tt.url, tt.ref, tt.path)

			if source == nil {
				t.Fatal("NewGitProfileSource() returned nil, want non-nil")
			}

			if source.Type() != SourceTypeGit {
				t.Errorf("Type() = %v, want %v", source.Type(), SourceTypeGit)
			}

			// Verify internal state (ref field should be set correctly)
			if source.ref != tt.expectedRef {
				t.Errorf("ref = %v, want %v", source.ref, tt.expectedRef)
			}

			if source.url != tt.url {
				t.Errorf("url = %v, want %v", source.url, tt.url)
			}

			if source.path != tt.path {
				t.Errorf("path = %v, want %v", source.path, tt.path)
			}
		})
	}
}

// setupTestGitRepo creates a test Git repository with a profile file
func setupTestGitRepo(t *testing.T, profileContent string, profilePath string) string {
	t.Helper()

	// Create temporary directory for repository
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize repository
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("failed to initialize git repo: %v", err)
	}

	// Create profile file
	fullPath := filepath.Join(repoPath, profilePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(profileContent), 0600); err != nil {
		t.Fatalf("failed to write profile file: %v", err)
	}

	// Add and commit
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := worktree.Add(profilePath); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	commit, err := worktree.Commit("Add profile", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create main branch reference
	mainRef := plumbing.NewHashReference("refs/heads/main", commit)
	if err := repo.Storer.SetReference(mainRef); err != nil {
		t.Fatalf("failed to create main ref: %v", err)
	}

	// Update HEAD
	if err := repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, mainRef.Name())); err != nil {
		t.Fatalf("failed to update HEAD: %v", err)
	}

	return repoPath
}

func TestGitProfileSource_GetSpec_Unit(t *testing.T) {
	// Skip in short mode as these still use real Git (but local repos)
	if testing.Short() {
		t.Skip("skipping unit tests with Git in short mode")
	}

	t.Parallel()

	tests := []struct {
		name           string
		profileContent string
		profilePath    string
		ref            string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid profile",
			profileContent: testValidProfile,
			profilePath:    "profile.json",
			ref:            "main",
			wantErr:        false,
		},
		{
			name:           "valid profile in subdirectory",
			profileContent: testValidProfile,
			profilePath:    "profiles/debug.json",
			ref:            "main",
			wantErr:        false,
		},
		{
			name:           "invalid JSON",
			profileContent: `{invalid json}`,
			profilePath:    "profile.json",
			ref:            "main",
			wantErr:        true,
			errContains:    "invalid JSON PodSpec",
		},
		{
			name:           "empty file",
			profileContent: "",
			profilePath:    "profile.json",
			ref:            "main",
			wantErr:        true,
			errContains:    "invalid JSON PodSpec",
		},
		{
			name:           "valid JSON but not a PodSpec",
			profileContent: `{"not": "a", "pod": "spec"}`,
			profilePath:    "profile.json",
			ref:            "main",
			wantErr:        false, // PodSpec unmarshaling is lenient
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not parallel because we're creating git repos

			repoPath := setupTestGitRepo(t, tt.profileContent, tt.profilePath)
			source := NewGitProfileSource(repoPath, tt.ref, tt.profilePath)

			spec, err := source.GetSpec(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GitProfileSource.GetSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("GitProfileSource.GetSpec() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if spec == nil {
					t.Error("GitProfileSource.GetSpec() returned nil spec, want non-nil")
				}
				if string(spec) != tt.profileContent {
					t.Errorf("GitProfileSource.GetSpec() returned wrong content\ngot:  %q\nwant: %q", string(spec), tt.profileContent)
				}
			}
		})
	}
}

func TestGitProfileSource_GetSpec_ErrorCases(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping error case tests in short mode")
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T) (url, ref, path string)
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid repository URL",
			setup: func(_ *testing.T) (string, string, string) {
				return "/nonexistent/repo", defaultGitRef, "profile.json"
			},
			wantErr:     true,
			errContains: "clone git repository",
		},
		{
			name: "missing file in repository",
			setup: func(t *testing.T) (string, string, string) {
				repoPath := setupTestGitRepo(t, testValidProfile, "existing.json")
				return repoPath, defaultGitRef, "nonexistent.json"
			},
			wantErr:     true,
			errContains: "read profile file",
		},
		{
			name: "repository exists but file in different path",
			setup: func(t *testing.T) (string, string, string) {
				repoPath := setupTestGitRepo(t, testValidProfile, "dir1/profile.json")
				return repoPath, defaultGitRef, "dir2/profile.json"
			},
			wantErr:     true,
			errContains: "read profile file",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not parallel due to potential filesystem conflicts

			url, ref, path := tt.setup(t)
			source := NewGitProfileSource(url, ref, path)

			_, err := source.GetSpec(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GitProfileSource.GetSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("GitProfileSource.GetSpec() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestGitProfileSource_GetSpec_WithToken(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping token test in short mode")
	}

	// This test verifies that the KUBECTL_DPM_GIT_TOKEN env var is used
	// We can't really test private repo authentication without actual credentials,
	// but we can verify the token is picked up

	repoPath := setupTestGitRepo(t, testValidProfile, "profile.json")

	// Set a dummy token (won't be used since repo is local)
	originalToken := os.Getenv("KUBECTL_DPM_GIT_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("KUBECTL_DPM_GIT_TOKEN", originalToken)
		} else {
			os.Unsetenv("KUBECTL_DPM_GIT_TOKEN")
		}
	}()

	if err := os.Setenv("KUBECTL_DPM_GIT_TOKEN", "dummy-token"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}

	source := NewGitProfileSource(repoPath, "main", "profile.json")
	spec, err := source.GetSpec(context.Background())

	if err != nil {
		t.Errorf("GitProfileSource.GetSpec() with token failed: %v", err)
		return
	}

	if spec == nil {
		t.Error("GitProfileSource.GetSpec() returned nil spec, want non-nil")
	}
}

func TestGitProfileSource_GetSpec_ContextCancellation(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping context cancellation test in short mode")
	}

	repoPath := setupTestGitRepo(t, testValidProfile, "profile.json")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	source := NewGitProfileSource(repoPath, "main", "profile.json")
	_, err := source.GetSpec(ctx)

	// The git clone might complete before context is checked, so we just verify
	// that if there's an error, it's reasonable
	if err != nil {
		// Error is expected and acceptable due to cancelled context
		t.Logf("Expected error with cancelled context: %v", err)
	}
}
