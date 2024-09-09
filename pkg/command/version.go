// SPDX-License-Identifier: MIT

package command

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information
var (
	// AppVersion is the version of the application
	appVersion = "v0.0.0-dev"

	// BuildTime is the time the application was built
	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format

	// GitCommit is the commit hash of the build
	gitCommit string
)

func Version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print current version",

		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version:\t%s\n", appVersion)
			fmt.Printf("BuildTime:\t%s\n", buildDate)
			fmt.Printf("GitCommit:\t%s\n", gitCommit)
			fmt.Printf("GoVersion:\t%s\n", runtime.Version())
			fmt.Printf("Compiler:\t%s\n", runtime.Compiler)
			fmt.Printf("OS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
