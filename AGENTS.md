<!-- SPDX-License-Identifier: MIT -->

# kubectl-dpm Developer Notes

`kubectl-dpm` is a kubectl plugin for managing Kubernetes debug profiles. Written in Go 1.26.3.

## Commands

**Development workflow** (order matters):
```bash
make fmt      # format code
make vet      # vet code
make lint     # golangci-lint (use GOLANGCI_LINT_EXTRA_ARGS for options)
make test     # run tests
make build    # builds bin/kubectl-dpm
```

**Run locally**:
```bash
make run      # runs without building binary
```

**Security scans** (required for CI):
```bash
make verify              # runs all verify-* targets
make verify-license      # license header check via hack/verify-license.sh
make verify-security     # govulncheck + nancy scan (requires OSS_TOKEN, OSS_USERNAME)
```

**Release**:
```bash
make sbom-generate       # creates tmp/kubectl-dpm.bom.spdx (prerequisite for release)
make release             # goreleaser with --clean
```

## Tools

Go tools are managed via `go.mod` `tool` directive, not globally installed:
- `go tool golangci-lint` (linter)
- `go tool goreleaser` (release automation)
- `go tool govulncheck` (vulnerability scanner)
- `go tool nancy` (dependency scanner)
- `go tool bom` (SBOM generator)

**Never** suggest installing these with `go install`. They're already available.

## Architecture

```
cmd/kubectl-dpm.go       # main entrypoint
pkg/
  command/               # cobra CLI commands and run logic
  config/                # koanf-based config loading
  profile/               # debug profile operations
  table/                 # table rendering (charmbracelet)
```

- **No `internal/`** – everything under `pkg/` is implicitly internal to this plugin
- **Main packages**: Cobra (CLI), k8s.io/client-go (Kubernetes), charmbracelet/* (TUI)

## Linting

`.golangci.yaml` enforces specific k8s import aliases:
```go
corev1          // k8s.io/api/core/v1
metav1          // k8s.io/apimachinery/pkg/apis/meta/v1
corev1client    // k8s.io/client-go/kubernetes/typed/core/v1
watchtools      // k8s.io/client-go/tools/watch
kubectldebug    // k8s.io/kubectl/pkg/cmd/debug
cmdutil         // k8s.io/kubectl/pkg/cmd/util
```

**Formatters enabled**: gci, gofmt, gofumpt, goimports
- Local prefix: `github.com/bavarianbidi/kubectl-dpm`
- Import sections: standard, default, local

## Testing

- No coverage file output (disabled due to golang/go#75031)
- Tests must pass `make test` which runs `fmt` and `vet` first

## Build/Release

**goreleaser** (`v2`) config:
- Builds for: linux/darwin × amd64/arm64
- Version info injected via ldflags:
  - `pkg/command.appVersion`
  - `pkg/command.buildDate`
  - `pkg/command.gitCommit`
- Archive format: `tar.gz`
- Extra release file: `tmp/kubectl-dpm.bom.spdx` (SBOM)

**Krew plugin**: `.krew.yaml` present (kubectl plugin manager)

## Known Issues

From `GOLANG_ANALYSIS.md`:
- Global state in `pkg/command/run.go` (debugProfile, flagProfileName, flagImage, flagDebug)
- `context.TODO()` used in pod listing (should use proper context)
- Inconsistent error wrapping (mix of bare errors, `errors.Wrap`, `fmt.Errorf`)
- Bare error returns lose context – should use `fmt.Errorf("context: %w", err)`

## CI

`.github/workflows/build.yaml`:
- Runs: `make verify`, `make lint` (1h timeout), `make test`
- Codecov upload (requires `CODECOV_TOKEN`)
- Requires: `OSS_TOKEN`, `OSS_USERNAME` for nancy scan

## License

All files require `SPDX-License-Identifier: MIT` header. Verified by `hack/verify-license.sh`.
