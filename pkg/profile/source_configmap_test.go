// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapProfileSource_GetSpec(t *testing.T) {
	t.Parallel()

	invalidJSON := `{invalid json}`

	tests := []struct {
		name        string
		configMap   *corev1.ConfigMap
		namespace   string
		cmName      string
		wantErr     bool
		errContains string
	}{
		{
			name:      "configmap with profile.json key",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"profile.json": testValidProfile,
				},
			},
			wantErr: false,
		},
		{
			name:      "configmap with profile key",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"profile": testValidProfile,
				},
			},
			wantErr: false,
		},
		{
			name:      "configmap with spec.json key",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"spec.json": testValidProfile,
				},
			},
			wantErr: false,
		},
		{
			name:      "configmap with spec key",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"spec": testValidProfile,
				},
			},
			wantErr: false,
		},
		{
			name:      "configmap with multiple keys - uses first match",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"profile.json": testValidProfile,
					"profile":      `{"other": "data"}`,
				},
			},
			wantErr: false,
		},
		{
			name:        "configmap not found",
			namespace:   "default",
			cmName:      "nonexistent",
			configMap:   nil,
			wantErr:     true,
			errContains: "failed to get ConfigMap",
		},
		{
			name:      "configmap without expected keys",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"other-key": testValidProfile,
				},
			},
			wantErr:     true,
			errContains: "does not contain any of the expected keys",
		},
		{
			name:      "configmap with invalid JSON",
			namespace: "default",
			cmName:    "test-cm",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
				Data: map[string]string{
					"profile.json": invalidJSON,
				},
			},
			wantErr:     true,
			errContains: "invalid JSON PodSpec",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fake clientset
			clientset := fake.NewSimpleClientset()

			// Create ConfigMap if provided
			if tt.configMap != nil {
				_, err := clientset.CoreV1().ConfigMaps(tt.namespace).Create(
					context.Background(),
					tt.configMap,
					metav1.CreateOptions{},
				)
				if err != nil {
					t.Fatalf("failed to create test ConfigMap: %v", err)
				}
			}

			source := NewConfigMapProfileSource(clientset.CoreV1(), tt.namespace, tt.cmName)
			spec, err := source.GetSpec(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigMapProfileSource.GetSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ConfigMapProfileSource.GetSpec() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr && spec == nil {
				t.Error("ConfigMapProfileSource.GetSpec() returned nil spec, want non-nil")
			}
		})
	}
}

func TestConfigMapProfileSource_Type(t *testing.T) {
	t.Parallel()

	clientset := fake.NewSimpleClientset()
	source := NewConfigMapProfileSource(clientset.CoreV1(), "default", "test-cm")

	if got := source.Type(); got != SourceTypeConfigMap {
		t.Errorf("ConfigMapProfileSource.Type() = %v, want %v", got, SourceTypeConfigMap)
	}
}

func TestConfigMapProfileSource_KeyPriority(t *testing.T) {
	t.Parallel()

	// This test verifies that keys are tried in the correct order:
	// profile.json > profile > spec.json > spec

	profiles := map[string]string{
		"profile.json": `{"volumeMounts": [{"name": "vol1", "mountPath": "/path1"}]}`,
		"profile":      `{"volumeMounts": [{"name": "vol2", "mountPath": "/path2"}]}`,
		"spec.json":    `{"volumeMounts": [{"name": "vol3", "mountPath": "/path3"}]}`,
		"spec":         `{"volumeMounts": [{"name": "vol4", "mountPath": "/path4"}]}`,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
		Data: profiles,
	}

	clientset := fake.NewSimpleClientset(configMap)
	source := NewConfigMapProfileSource(clientset.CoreV1(), "default", "test-cm")

	spec, err := source.GetSpec(context.Background())
	if err != nil {
		t.Fatalf("GetSpec() failed: %v", err)
	}

	// Should use profile.json (first priority)
	expected := profiles["profile.json"]
	if string(spec) != expected {
		t.Errorf("GetSpec() used wrong key, got %q, want %q", string(spec), expected)
	}
}
