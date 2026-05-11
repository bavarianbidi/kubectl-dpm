// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	corev1 "k8s.io/api/core/v1"
)

const (
	// defaultGitRef is the default branch reference when none is specified.
	defaultGitRef = "main"
)

// GitProfileSource represents a profile source that fetches from a Git repository.
type GitProfileSource struct {
	url  string
	ref  string
	path string
}

// NewGitProfileSource creates a new Git-based profile source.
// The ref parameter is optional and defaults to "main" if empty.
func NewGitProfileSource(url, ref, path string) *GitProfileSource {
	if ref == "" {
		ref = defaultGitRef
	}
	return &GitProfileSource{
		url:  url,
		ref:  ref,
		path: path,
	}
}

// GetSpec clones the Git repository and returns the JSON specification from the specified path.
func (g *GitProfileSource) GetSpec(ctx context.Context) ([]byte, error) {
	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "kubectl-dpm-git-*")
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Prepare clone options.
	// Note: go-git does not support fetching individual files without cloning the repository.
	// We use a shallow clone (Depth=1) to fetch only the latest commit and SingleBranch=true
	// to clone only the specified branch, minimizing bandwidth and storage impact.
	cloneOpts := &git.CloneOptions{
		URL:           g.url,
		Depth:         1,    // Shallow clone: only fetch the latest commit
		SingleBranch:  true, // Only clone the specified branch, not all branches
		ReferenceName: plumbing.NewBranchReferenceName(g.ref),
	}

	// Check for Git personal access token for private repositories
	if token := os.Getenv("KUBECTL_DPM_GIT_TOKEN"); token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "git", // This can be anything for token-based auth
			Password: token,
		}
	}

	// Clone the repository
	_, err = git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("clone git repository %s@%s: %w", g.url, g.ref, err)
	}

	// Read the profile file
	profilePath := filepath.Join(tmpDir, g.path)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("read profile file %q from git repo: %w", g.path, err)
	}

	// Validate it's valid JSON representing a PodSpec
	var podSpec corev1.PodSpec
	if err := json.Unmarshal(data, &podSpec); err != nil {
		return nil, fmt.Errorf("invalid JSON PodSpec in git repo %s@%s:%s: %w", g.url, g.ref, g.path, err)
	}

	return data, nil
}

// Type returns the source type identifier.
func (g *GitProfileSource) Type() string {
	return SourceTypeGit
}
