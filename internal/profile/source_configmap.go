// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ConfigMapProfileSource represents a profile source that reads from a Kubernetes ConfigMap.
type ConfigMapProfileSource struct {
	client    corev1client.CoreV1Interface
	namespace string
	name      string
}

// NewConfigMapProfileSource creates a new ConfigMap-based profile source.
func NewConfigMapProfileSource(client corev1client.CoreV1Interface, namespace, name string) *ConfigMapProfileSource {
	return &ConfigMapProfileSource{
		client:    client,
		namespace: namespace,
		name:      name,
	}
}

// GetSpec fetches and returns the JSON specification from the ConfigMap.
// It tries multiple conventional key names in order: profile.json, profile, spec.json, spec
func (c *ConfigMapProfileSource) GetSpec(ctx context.Context) ([]byte, error) {
	cm, err := c.client.ConfigMaps(c.namespace).Get(ctx, c.name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap %s/%s: %w", c.namespace, c.name, err)
	}

	// Try multiple conventional key names
	tryKeys := []string{"profile.json", "profile", "spec.json", "spec"}
	var data string
	var foundKey string

	for _, key := range tryKeys {
		if val, ok := cm.Data[key]; ok {
			data = val
			foundKey = key
			break
		}
	}

	if foundKey == "" {
		return nil, fmt.Errorf("ConfigMap %s/%s does not contain any of the expected keys: %v", c.namespace, c.name, tryKeys)
	}

	// Validate it's valid JSON representing a PodSpec
	var podSpec corev1.PodSpec
	if err := json.Unmarshal([]byte(data), &podSpec); err != nil {
		return nil, fmt.Errorf("invalid JSON PodSpec in ConfigMap %s/%s key %q: %w", c.namespace, c.name, foundKey, err)
	}

	return []byte(data), nil
}

// Type returns the source type identifier.
func (c *ConfigMapProfileSource) Type() string {
	return SourceTypeConfigMap
}
