// SPDX-License-Identifier: MIT

package profile

// testValidProfile is a valid PodSpec JSON used across multiple tests
const testValidProfile = `{
	"volumeMounts": [
		{
			"mountPath": "/app/config",
			"name": "app-config",
			"readOnly": true
		}
	]
}`
