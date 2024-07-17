// SPDX-License-Identifier: MIT

package command

import (
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func TestGetTargetNamespace(t *testing.T) {
	tests := []struct {
		name         string
		debugProfile profile.Profile
		kubectlNs    string
		want         string
	}{
		{
			name: "debug profile has namespace",
			debugProfile: profile.Profile{
				Namespace: "custom-namespace",
			},
			kubectlNs: "default",
			want:      "custom-namespace",
		},
		{
			name:         "debug profile does not have namespace, kubectl namespace is not default",
			debugProfile: profile.Profile{},
			kubectlNs:    "custom-namespace",
			want:         "custom-namespace",
		},
		{
			name:         "debug profile does not have namespace, kubectl namespace is default",
			debugProfile: profile.Profile{},
			kubectlNs:    "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debugProfile = tt.debugProfile

			kubeConfigFlags := genericclioptions.NewConfigFlags(true)
			// nolint:gosec
			kubeConfigFlags.Namespace = &tt.kubectlNs
			MatchVersionKubeConfigFlags = cmdutil.NewMatchVersionFlags(kubeConfigFlags)

			got := getTargetNamespace()

			if got != tt.want {
				t.Errorf("getTargetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
