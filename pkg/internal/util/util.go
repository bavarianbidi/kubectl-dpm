// SPDX-License-Identifier: MIT

package util

import (
	corev1 "k8s.io/api/core/v1"
)

func ContainerNameToRef(pod *corev1.Pod) map[string]*corev1.Container {
	names := map[string]*corev1.Container{}
	for i := range pod.Spec.Containers {
		ref := &pod.Spec.Containers[i]
		names[ref.Name] = ref
	}
	for i := range pod.Spec.InitContainers {
		ref := &pod.Spec.InitContainers[i]
		names[ref.Name] = ref
	}
	for i := range pod.Spec.EphemeralContainers {
		ref := (*corev1.Container)(&pod.Spec.EphemeralContainers[i].EphemeralContainerCommon)
		names[ref.Name] = ref
	}
	return names
}

func GetContainerStatusByName(pod *corev1.Pod, containerName string) *corev1.ContainerStatus {
	allContainerStatus := [][]corev1.ContainerStatus{pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses, pod.Status.EphemeralContainerStatuses}
	for _, statusSlice := range allContainerStatus {
		for i := range statusSlice {
			if statusSlice[i].Name == containerName {
				return &statusSlice[i]
			}
		}
	}
	return nil
}
