// SPDX-License-Identifier: MIT

package table

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func TestGenerateTable(t *testing.T) {
	tests := []struct {
		name     string
		profiles []profile.Profile
		wide     bool
		expected []table.Row
	}{
		{
			name: "basic profiles",
			profiles: []profile.Profile{
				{
					ProfileName: "profile1",
					Profile:     "test_data/profile1.json",
					Image:       "busybox",
					Namespace:   "default",
					MatchLabels: map[string]string{"app": "test"},
				},
				{
					ProfileName: "profile2",
					Profile:     "test_data/profile2.json",
					Image:       "nginx",
					Namespace:   "kube-system",
					MatchLabels: map[string]string{"app": "nginx"},
				},
			},
			wide: false,
			expected: []table.Row{
				{"profile1", "test_data/profile1.json"},
				{"profile2", "test_data/profile2.json"},
			},
		},
		{
			name: "wide profiles",
			profiles: []profile.Profile{
				{
					ProfileName: "profile1",
					Profile:     "test_data/profile1.json",
					Image:       "busybox",
					Namespace:   "default",
					MatchLabels: map[string]string{"app": "test"},
				},
				{
					ProfileName: "profile2",
					Profile:     "test_data/profile2.json",
					Image:       "nginx",
					Namespace:   "kube-system",
					MatchLabels: map[string]string{"app": "nginx"},
				},
			},
			wide: true,
			expected: []table.Row{
				{"profile1", "test_data/profile1.json", "busybox", "default", `{"app":"test"}`},
				{"profile2", "test_data/profile2.json", "nginx", "kube-system", `{"app":"nginx"}`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateTable(tt.profiles, tt.wide)

			if len(got.Rows()) != len(tt.expected) {
				t.Errorf("expected %d rows, got %d", len(tt.expected), len(got.Rows()))
			}

			for i, row := range got.Rows() {
				for j, cell := range row {
					if cell != tt.expected[i][j] {
						t.Errorf("expected cell %d in row %d to be %v, got %v", j, i, tt.expected[i][j], cell)
					}
				}
			}
		})
	}
}
