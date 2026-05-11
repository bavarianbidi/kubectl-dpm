// SPDX-License-Identifier: MIT

package table

import (
	"fmt"
	"strings"
	"testing"

	bubbletable "github.com/charmbracelet/bubbles/table"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func TestGenerateTable(t *testing.T) {
	tests := []struct {
		name     string
		profiles []profile.Profile
		wide     bool
		expected []bubbletable.Row
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
			expected: []bubbletable.Row{
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
			expected: []bubbletable.Row{
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

func TestConfigureInteractive(t *testing.T) {
	tbl := GenerateTable([]profile.Profile{{ProfileName: "profile1", Profile: "test_data/profile1.json"}}, false)

	ConfigureInteractive(&tbl)

	if !tbl.Focused() {
		t.Fatal("expected interactive table to be focused")
	}
}

func TestConfigureStaticShowsAllRows(t *testing.T) {
	profiles := make([]profile.Profile, 0, 25)
	for i := range 25 {
		profiles = append(profiles, profile.Profile{
			ProfileName: fmt.Sprintf("profile-%02d", i),
			Profile:     fmt.Sprintf("test_data/profile-%02d.json", i),
		})
	}

	tbl := GenerateTable(profiles, false)
	ConfigureStatic(&tbl)

	if tbl.Focused() {
		t.Fatal("expected static table to be blurred")
	}

	view := tbl.View()
	if !strings.Contains(view, "profile-24") {
		t.Fatalf("expected static table view to include last row, got %q", view)
	}
	if tbl.Height() < len(tbl.Rows()) {
		t.Fatalf("expected static table height %d to include all %d rows", tbl.Height(), len(tbl.Rows()))
	}
}
