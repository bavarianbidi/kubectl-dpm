<!-- SPDX-License-Identifier: MIT -->

# `dpm` the `kubectl debug-profile-manager`

The `dpm` is a `kubectl` plugin build to share custom debug profiles with others.

*Pull requests, bug reports, and all other forms of contribution are welcomed and highly encouraged!* :heart:

## `kubectl` custom debug profiles

With `kubectl v1.30` an alpha feature has landed to make it much more easier to create an `ephemeral` debug container by defining a custom profile which gets applied to the debug container.

It's now possible to define the same `volumeMounts` from a running application container in your debugging session.

```bash
KUBECTL_DEBUG_CUSTOM_PROFILE=true kubectl debug -it <POD_NAMe> --image=<DEBUG_CONTAINER_IMAGE> --target=<TARGET_CONTAINER> --custom="<PATH_TO_CUSTOM_DEBUG_PROFILE>"
```

With the below `pod.Spec` in json format as debug profile name (`PATH_TO_CUSTOM_DEBUG_PROFILE`), an ephemeral container will get attached to the existing pod with a custom container image (e.g. `alpine:latest`) and the same `volumeMounts` the application container might have.

```json
{
    "volumeMounts": [
        {
            "mountPath": "/opt/app",
            "name": "<VOLUME_NAME>",
            "readOnly": true
        }
    ]
}
```

## why `dpm`

Initially `kubectl-dpm` was build to get more familiar with the custom debug profile feature. Till version `v0.0.4` the `kubectl-dpm` plugin was adopting many
internal `func` from the `kubectl` implementation to make it more easier to create a custom debug profile.

Starting with version `v0.1.0` the `kubectl-dpm` is now focused on managing multiple custom debug profiles. `kubectl-dpm` discovers the path to `kubectl` and uses it to start a debugging container.

### Lower the barrier for having minimal container images for application workload

Minimal container images (`scratch`, `distroless`, ...) are awesome and from a security point of view something, you should definitely have.

But when it came to a rollout, operational aspects became more and more relevant, especially when something doesn't work than you expect.

How to check in a `distroless` image:
* if the application created a TCP-Socket
* `helm` (or any other templating engine) templated the app configuration from a `secret` or `configMap` into the `Deployment` you wanted to

To address this and other limitations with minimal container images and some limitations of the _current_ custom debug profile implementation, I've added a small configuration layer in front of `kubectl debug` to make it much more easier to start a debugging session. With that it's also possible to share a tested debug configuration within your team.

## configuration

The `dpm` needs a configuration file where re-usable profiles are stored.

### minimal configuration

As a minimal configuration, the following fields are needed:

```yaml
profiles:
  - name: <PROFILE_NAME>
    profile: <PATH_TO_PROFILE>
```

The `profile` field is a path to a json file which contains the `pod.Spec` of the debug container.

```json
{
    "volumeMounts": [
        {
            "mountPath": "/opt/app",
            "name": "<VOLUME_NAME>",
            "readOnly": true
        }
    ]
}
```

With the above configuration, the following command can be executed:

```bash
kubectl dpm run --profile=<PROFILE_NAME> --config=<DPM_CONFIG_FILE> --image=alpine/k8s:1.29.0 --namespace=<NAMESPACE> <POD_NAME>
```

### full configuration

The full configuration file looks like this:

```yaml
profiles:
  - name: <PROFILE_NAME>
    profile: <PATH_TO_PROFILE>
    image: <DEBUG_CONTAINER_IMAGE>
    namespace: <NAMESPACE>
    matchLabels:
      <LABEL_KEY>: <LABEL_VALUE>
```

With the above configuration, the following command can be executed:

```bash
kubectl dpm run --profile=<PROFILE_NAME> --config=<DPM_CONFIG_FILE>
```

`dpm` will use the defined `namespace` and `image` to generate the ephemeral debug container.
As target container, the first running container with the matching `matchLabels` will get selected.


### `kubectlPath`

`dpm` needs to know where the `kubectl` binary is located. By default,
`dpm` is using the value of the `_` environment variable to determine the path to the `kubectl` binary.
This only works, if the `dpm` is run as a `kubectl` plugin.

As standalone binary, the `kubectlPath` value must be defined.

## flags

The `dpm` has the following flags:

* `--profile` - the name of the profile to use
* `--config` - the path to the configuration file
* `--image` - the image of the debug container

As we also register the generic `kubectl` flags, the following _relevant_  flags (IMHO) are also available:

* `--namespace` - the namespace of the pod
* `--context` - the context of the pod
* `--kubeconfig` - the path to the kubeconfig file
