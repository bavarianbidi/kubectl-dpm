// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubectl/pkg/cmd/attach"
	kubectldebug "k8s.io/kubectl/pkg/cmd/debug"
	"k8s.io/kubectl/pkg/cmd/exec"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/interrupt"
	"k8s.io/kubectl/pkg/util/term"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/internal/util"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

type CustomDebugOptions struct {
	*kubectldebug.DebugOptions
	podClient   corev1client.CoreV1Interface
	kubectlPath string
}

func NewCustomDebugOptions(streams genericiooptions.IOStreams) *CustomDebugOptions {
	debugOptions := kubectldebug.NewDebugOptions(streams)

	return &CustomDebugOptions{
		DebugOptions: debugOptions,
	}
}

var (
	MatchVersionKubeConfigFlags *cmdutil.MatchVersionFlags
	debugProfile                profile.Profile
)

func NewCmdDebugProfile(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewCustomDebugOptions(streams)

	// add kubeconfig flags
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	MatchVersionKubeConfigFlags = cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	cmd := &cobra.Command{
		Use: "run",
		// DisableFlagsInUseLine: true,
		Short: "create an ephemeral debug container in a pod",
		Long:  "create an ephemeral debug container in a pod by using the kubectl debug implementation and a custom profile",

		RunE: func(c *cobra.Command, args []string) error {
			// if no args are given, print help
			if c.Flags().NFlag() == 0 {
				c.Help()
				os.Exit(0)
			}

			if err := config.GenerateConfig(); err != nil {
				return err
			}

			if err := o.run(streams, args); err != nil {
				return err
			}

			return nil
		},
	}

	// define flags
	flags := cmd.PersistentFlags()

	// add kubeconfig flags
	kubeConfigFlags.AddFlags(flags)

	// add custom flag
	cmd.Flags().StringVar(&flagProfileName, "profile", "", "profile name")
	cmd.Flags().StringVar(&flagImage, "image", "", "image to use for the debug container")

	return cmd
}

func (o *CustomDebugOptions) run(streams genericiooptions.IOStreams, args []string) error {
	// validate kubectl path
	if err := profile.ValidateKubectlPath(); err != nil {
		return err
	}

	// set kubectl path
	o.kubectlPath = profile.Config.KubectlPath

	// validate and complete profile
	if err := profile.ValidateAndCompleteProfile(flagProfileName); err != nil {
		return err
	}

	// get the index of the profile where the profile name matches
	idx := slices.IndexFunc(profile.Config.Profiles,
		func(c profile.Profile) bool { return c.ProfileName == flagProfileName },
	)

	fmt.Printf("Profile: %+v\n", debugProfile)

	debugProfile = profile.Config.Profiles[idx]

	// this is not needed atm as we want support kubectl versions < 1.30
	//
	// to just wrap the kubectl-debug command, we have to change a few things in here
	// nothing special, but it has to be done :-)
	//
	// enable custom debug profile via env var KUBECTL_DEBUG_CUSTOM_PROFILE
	// os.Setenv(string(cmdutil.DebugCustomProfile), "true")

	if err := o.prepareEphemeralContainer(streams, args); err != nil {
		return err
	}

	return nil
}

func (o *CustomDebugOptions) prepareEphemeralContainer(streams genericiooptions.IOStreams, args []string) error {
	o.DebugOptions.CustomProfileFile = debugProfile.CustomProfileFile
	o.DebugOptions.Profile = kubectldebug.ProfileLegacy

	// use image from profile
	o.DebugOptions.Image = debugProfile.Image

	// the flag has the highest priority
	if flagImage != "" {
		o.DebugOptions.Image = flagImage
	}

	if o.DebugOptions.Image == "" {
		return fmt.Errorf("image is required")
	}

	// i'm not using the upstream complete func as it requires flags of the debug cmd which i do not have atm
	//
	// do the custom complete func
	if err := o.complete(MatchVersionKubeConfigFlags); err != nil {
		return err
	}

	// get the namespace from the profile or the current kube context
	// if no namespace is given, use the default namespace
	// the namespace flag has the biggest priority
	//
	// TODO: move this to the complete func
	//nolint:godox
	kubectlNamespace, _, _ := MatchVersionKubeConfigFlags.ToRawKubeConfigLoader().Namespace()

	switch {
	case kubectlNamespace != corev1.NamespaceDefault:
		o.Namespace = kubectlNamespace
	case debugProfile.Namespace != "":
		o.Namespace = debugProfile.Namespace
	default:
		o.Namespace = corev1.NamespaceDefault
	}

	// get the pod name from the profile label selector if no args are given
	if len(args) == 0 && len(debugProfile.MatchLabels) > 0 {
		matchingPods, err := o.podClient.Pods(o.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(
				&metav1.LabelSelector{
					MatchLabels: debugProfile.MatchLabels,
				}),
		})
		if err != nil {
			return err
		}

		if len(matchingPods.Items) == 0 {
			return fmt.Errorf("no pods in namespace %s found with label selector %v", o.Namespace, debugProfile.MatchLabels)
		}

		for _, pod := range matchingPods.Items {
			args = append(args, pod.Name)
		}
	}

	// set target names
	o.DebugOptions.TargetNames = args

	// Validate is the upstream package Validate func
	if err := o.DebugOptions.Validate(); err != nil {
		return err
	}

	if err := o.Run(MatchVersionKubeConfigFlags, streams); err != nil {
		return err
	}

	return nil
}

// Complete completes all the required options to make the debug command work
// it's primary a copy of the upstream complete func
func (o *CustomDebugOptions) complete(restClientGetter genericclioptions.RESTClientGetter) error {
	o.PullPolicy = corev1.PullIfNotPresent

	o.TTY = true
	o.Interactive = true

	//
	applier, err := kubectldebug.NewProfileApplier(o.Profile)
	if err != nil {
		return err
	}
	o.Applier = applier

	o.Attach = true

	o.WarningPrinter = printers.NewWarningPrinter(o.ErrOut, printers.WarningPrinterOptions{Color: term.AllowsColorOutput(o.ErrOut)})

	if o.CustomProfileFile != "" {
		customProfileBytes, err := os.ReadFile(o.CustomProfileFile)
		if err != nil {
			return fmt.Errorf("must pass a container spec json file for custom profile: %w", err)
		}

		err = json.Unmarshal(customProfileBytes, &o.CustomProfile)
		if err != nil {
			return fmt.Errorf("%s does not contain a valid container spec: %w", o.CustomProfileFile, err)
		}
	}

	config, err := restClientGetter.ToRESTConfig()
	if err != nil {
		return err
	}

	client, err := corev1client.NewForConfig(config)
	if err != nil {
		return err
	}

	o.podClient = client

	o.Builder = resource.NewBuilder(restClientGetter)

	return nil
}

func (o *CustomDebugOptions) Run(restClientGetter genericclioptions.RESTClientGetter, streams genericiooptions.IOStreams) error {
	ctx := context.Background()

	r := o.Builder.
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		FilenameParam(true, &o.FilenameOptions).
		NamespaceParam(o.Namespace).DefaultNamespace().ResourceNames("pods", o.TargetNames...).
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	err := r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		var (
			debugPod      *corev1.Pod
			containerName string
			visitErr      error
		)
		switch obj := info.Object.(type) {
		case *corev1.Pod:
			debugPod, containerName, visitErr = o.debugByEphemeralContainer(ctx, obj)
		default:
			visitErr = fmt.Errorf("%q not supported by debug", info.Mapping.GroupVersionKind)
		}
		if visitErr != nil {
			return visitErr
		}

		o.AttachFunc = func(ctx context.Context, restClientGetter genericclioptions.RESTClientGetter, cmdPath, ns, podName, containerName string) error {
			return o.attachPod(ctx, streams, restClientGetter, cmdPath, debugPod.Namespace, debugPod.Name, containerName)
		}

		if o.Attach && len(containerName) > 0 && o.AttachFunc != nil {
			if err := o.AttachFunc(ctx, restClientGetter, "cmdPath", debugPod.Namespace, debugPod.Name, containerName); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// func attachPod(ctx context.Context, streams genericiooptions.IOStreams, restClientGetter genericclioptions.RESTClientGetter, cmdPath string, ns, podName, containerName string) kubectldebug.DebugAttachFunc {
func (o *CustomDebugOptions) attachPod(ctx context.Context, streams genericiooptions.IOStreams, restClientGetter genericclioptions.RESTClientGetter, cmdPath string, ns, podName, containerName string) error {
	opts := &attach.AttachOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: streams,
			Stdin:     true,
			TTY:       o.TTY,
			Quiet:     o.Quiet,
		},
		CommandName: cmdPath + " attach",

		Attach: &attach.DefaultRemoteAttach{},
	}
	config, err := restClientGetter.ToRESTConfig()
	if err != nil {
		return err
	}
	opts.Config = config

	// discover pod again
	// and get namespace, podname and newly created debug container name
	pod, err := o.waitForContainer(ctx, ns, podName, containerName)
	if err != nil {
		return err
	}

	opts.Namespace = ns
	opts.Pod = pod
	opts.PodName = podName
	opts.ContainerName = containerName
	if opts.AttachFunc == nil {
		opts.AttachFunc = attach.DefaultAttachFunc
	}

	status := util.GetContainerStatusByName(pod, containerName)
	if status == nil {
		// impossible path
		return fmt.Errorf("error getting container status of container name %q: %+v", containerName, err)
	}
	if status.State.Terminated != nil {
		return fmt.Errorf("ephemeral container %q terminated", containerName)
	}

	if err := opts.Run(); err != nil {
		fmt.Fprintf(opts.ErrOut, "couldn't attach to pod/%s\n", podName)
	}
	return nil
}

// debugByEphemeralContainer runs an EphemeralContainer in the target Pod for use as a debug container
//
// taken from: https://github.com/kubernetes/kubernetes/blob/9791f0d1f39f3f1e0796add7833c1059325d5098/staging/src/k8s.io/kubectl/pkg/cmd/debug/debug.go#L456-L494
// and modified to make it usable in here
func (o *CustomDebugOptions) debugByEphemeralContainer(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, string, error) {
	podJS, err := json.Marshal(pod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON for pod: %v", err)
	}

	debugPod, debugContainer, err := o.generateDebugContainer(pod)
	if err != nil {
		return nil, "", err
	}

	debugJS, err := json.Marshal(debugPod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON for debug container: %v", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(podJS, debugJS, pod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating patch to add debug container: %v", err)
	}

	pods := o.podClient.Pods(pod.Namespace)
	result, err := pods.Patch(ctx, pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		// The apiserver will return a 404 when the EphemeralContainers feature is disabled because the `/ephemeralcontainers` subresource
		// is missing. Unlike the 404 returned by a missing pod, the status details will be empty.
		if serr, ok := err.(*errors.StatusError); ok && serr.Status().Reason == metav1.StatusReasonNotFound && serr.ErrStatus.Details.Name == "" {
			return nil, "", fmt.Errorf("ephemeral containers are disabled for this cluster (error from server: %q)", err)
		}

		return nil, "", err
	}

	return result, debugContainer.Name, nil
}

// generateDebugContainer returns a debugging pod and an EphemeralContainer suitable for use as a debug container
// in the given pod.
func (o *CustomDebugOptions) generateDebugContainer(pod *corev1.Pod) (*corev1.Pod, *corev1.EphemeralContainer, error) {
	name := o.computeDebugContainerName(pod)
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:                     name,
			Env:                      o.Env,
			Image:                    o.Image,
			ImagePullPolicy:          o.PullPolicy,
			Stdin:                    o.Interactive,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TTY:                      o.TTY,
		},
		TargetContainerName: o.TargetContainer,
	}

	if o.ArgsOnly {
		ec.Args = o.Args
	} else {
		ec.Command = o.Args
	}

	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)
	if err := o.Applier.Apply(copied, name, copied); err != nil {
		return nil, nil, err
	}

	if o.CustomProfile != nil {
		err := o.applyCustomProfileEphemeral(copied, ec.Name)
		if err != nil {
			return nil, nil, err
		}
	}

	ec = &copied.Spec.EphemeralContainers[len(copied.Spec.EphemeralContainers)-1]

	return copied, ec, nil
}

func (o *CustomDebugOptions) computeDebugContainerName(pod *corev1.Pod) string {
	if len(o.Container) > 0 {
		return o.Container
	}

	cn, containerByName := "", util.ContainerNameToRef(pod)
	for len(cn) == 0 || (containerByName[cn] != nil) {
		cn = fmt.Sprintf("debugger-%s", utilrand.String(5))
	}
	if !o.Quiet {
		fmt.Fprintf(o.Out, "Defaulting debug container name to %s.\n", cn)
	}
	return cn
}

// applyCustomProfileEphemeral applies given partial container json file on to the profile
// incorporated ephemeral container of the pod.
func (o *CustomDebugOptions) applyCustomProfileEphemeral(debugPod *corev1.Pod, containerName string) error {
	o.CustomProfile.Name = containerName
	customJS, err := json.Marshal(o.CustomProfile)
	if err != nil {
		return fmt.Errorf("unable to marshall custom profile: %w", err)
	}

	var index int
	found := false
	for i, val := range debugPod.Spec.EphemeralContainers {
		if val.Name == containerName {
			index = i
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unable to find the %s ephemeral container in the pod %s", containerName, debugPod.Name)
	}

	var debugContainerJS []byte
	debugContainerJS, err = json.Marshal(debugPod.Spec.EphemeralContainers[index])
	if err != nil {
		return fmt.Errorf("unable to marshall ephemeral container:%w", err)
	}

	patchedContainer, err := strategicpatch.StrategicMergePatch(debugContainerJS, customJS, corev1.Container{})
	if err != nil {
		return fmt.Errorf("error creating three way patch to add debug container: %w", err)
	}

	err = json.Unmarshal(patchedContainer, &debugPod.Spec.EphemeralContainers[index])
	if err != nil {
		return fmt.Errorf("unable to unmarshall patched container to ephemeral container: %w", err)
	}

	return nil
}

// waitForContainer watches the given pod until the container is running
func (o *CustomDebugOptions) waitForContainer(ctx context.Context, ns, podName, containerName string) (*corev1.Pod, error) {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(ctx, 0*time.Second)
	defer cancel()

	fieldSelector := fields.OneTermEqualSelector("metadata.name", podName).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return o.podClient.Pods(ns).List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return o.podClient.Pods(ns).Watch(ctx, options)
		},
	}

	intr := interrupt.New(nil, cancel)
	var result *corev1.Pod
	err := intr.Run(func() error {
		ev, err := watchtools.UntilWithSync(ctx, lw, &corev1.Pod{}, nil, func(ev watch.Event) (bool, error) {
			if ev.Type == watch.Deleted {
				return false, errors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
			}

			p, ok := ev.Object.(*corev1.Pod)
			if !ok {
				return false, fmt.Errorf("watch did not return a pod: %v", ev.Object)
			}

			s := util.GetContainerStatusByName(p, containerName)
			if s == nil {
				return false, nil
			}
			if s.State.Running != nil || s.State.Terminated != nil {
				return true, nil
			}
			if !o.Quiet && s.State.Waiting != nil && s.State.Waiting.Message != "" {
				o.WarningPrinter.Print(fmt.Sprintf("container %s: %s", containerName, s.State.Waiting.Message))
			}
			return false, nil
		})
		if ev != nil {
			result = ev.Object.(*corev1.Pod)
		}
		return err
	})

	return result, err
}
