// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/bavarianbidi/kubectl-dpm/pkg/config"
	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

var (
	MatchVersionKubeConfigFlags *cmdutil.MatchVersionFlags
	debugProfile                profile.Profile
)

func NewCmdDebugProfile(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		// DisableFlagsInUseLine: true,
		Short: "create an ephemeral debug container in a pod",
		Long:  "create an ephemeral debug container in a pod by using the kubectl debug implementation and a custom profile",
		// at most one argument is allowed, which is the target pod name. If no argument is provided, the plugin will try to find a target pod based on the profile's matchLabels and use the first container in that pod as target.
		Args: cobra.MaximumNArgs(1),

		RunE: func(c *cobra.Command, args []string) error {
			if err := config.GenerateConfig(); err != nil {
				return fmt.Errorf("generate config: %w", err)
			}

			// if no profile flag is set, start interactive mode to select a profile
			if !c.Flags().Changed(profileFlagName) {
				model, err := initTeaModel(c.Context())
				if err != nil {
					return fmt.Errorf("run debug command: %w", err)
				}
				p := tea.NewProgram(model)
				if _, err := p.Run(); err != nil {
					return fmt.Errorf("error running program: %w", err)
				}

				if flagProfileName == "" {
					return fmt.Errorf("no profile selected - exiting")
				}
			}

			if err := run(c.Context(), args, streams); err != nil {
				return fmt.Errorf("initialize interactive mode: %w", err)
			}

			return nil
		},
	}

	// define flags
	flags := cmd.PersistentFlags()

	// add kubeconfig flags
	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	MatchVersionKubeConfigFlags = cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	kubeConfigFlags.AddFlags(flags)

	// add custom flag
	cmd.Flags().StringVarP(&flagProfileName, profileFlagName, "p", "", "profile name")
	cmd.Flags().StringVarP(&flagImage, "image", "i", "", "image to use for the debug container")
	cmd.Flags().BoolVarP(&flagDebug, "debug", "d", false, "print debug information")

	return cmd
}

func run(ctx context.Context, args []string, streams genericiooptions.IOStreams) error {
	// validate kubectl path
	if err := profile.ValidateKubectlPath(); err != nil {
		return fmt.Errorf("run debug profile: %w", err)
	}

	// check kubectl version
	if err := profile.CheckKubectlVersion(); err != nil {
		return fmt.Errorf("validate kubectl path: %w", err)
	}

	// complete profile
	if err := profile.CompleteProfile(flagProfileName); err != nil {
		return fmt.Errorf("check kubectl version: %w", err)
	}

	// validate profile
	if err := profile.ValidateProfile(ctx, flagProfileName); err != nil {
		return fmt.Errorf("complete profile %q: %w", flagProfileName, err)
	}

	// get the index of the profile where the profile name matches
	idx, err := profile.GetProfileIdx(flagProfileName)
	if err != nil {
		return err
	}

	debugProfile = profile.Config.Profiles[idx]

	// For ConfigMap sources, we need to inject a Kubernetes client
	if debugProfile.ProfileSource.Type == profile.SourceTypeConfigMap && debugProfile.GetSource() == nil {
		restClient, err := MatchVersionKubeConfigFlags.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("get REST config: %w", err)
		}

		clientset, err := corev1client.NewForConfig(restClient)
		if err != nil {
			return fmt.Errorf("create k8s clientset: %w", err)
		}

		// Inject the client and validate the ConfigMap source
		if err := profile.InitializeConfigMapSource(ctx, &debugProfile, clientset); err != nil {
			return fmt.Errorf("initialize configmap source: %w", err)
		}
		profile.Config.Profiles[idx] = debugProfile
	}

	var targetContainer string
	namespace := getTargetNamespace()

	switch {
	case len(args) == 1:
		targetContainer = args[0]
	case len(debugProfile.MatchLabels) > 0:
		var err error
		targetContainer, err = getTargetPod(ctx, namespace)
		if err != nil {
			return fmt.Errorf("get target pod in namespace %q: %w", namespace, err)
		}
	default:
		return fmt.Errorf("no target container specified")
	}

	if flagDebug {
		fmt.Fprintf(streams.Out, "Using profile: %+v\n", debugProfile)
		fmt.Fprintf(streams.Out, "kubectl path: %s\n", os.ExpandEnv(profile.Config.KubectlPath))
		if debugProfile.GetSource() != nil {
			fmt.Fprintf(streams.Out, "profile source type: %s\n", debugProfile.GetSource().Type())
		} else {
			fmt.Fprintf(streams.Out, "profile path (legacy): %s\n", debugProfile.Profile)
			fmt.Fprintf(streams.Out, "profile path resolved (legacy): %s\n", os.ExpandEnv(debugProfile.Profile))
		}
	}

	var debugCommand *exec.Cmd

	// Use ProfileSource if available, otherwise fall back to legacy Profile field
	if debugProfile.GetSource() != nil {
		source := debugProfile.GetSource()

		if source.Type() == profile.SourceTypeBuiltIn {
			// Built-in profile - use --profile flag with the profile name
			builtInSource, ok := source.(*profile.BuiltInProfileSource)
			if !ok {
				return fmt.Errorf("internal error: built-in source type assertion failed")
			}

			// nolint:gosec
			debugCommand = exec.Command(
				os.ExpandEnv(profile.Config.KubectlPath),
				"debug",
				"--namespace", namespace,
				"--profile", builtInSource.ProfileName(),
				"--image", debugProfile.Image, targetContainer,
				"-it",
			)
		} else {
			// Custom profile - fetch spec and write to temp file
			specData, err := source.GetSpec(ctx)
			if err != nil {
				return fmt.Errorf("fetch profile spec from %s source: %w", source.Type(), err)
			}

			// Create temp file for the profile spec
			tmpFile, err := os.CreateTemp("", "kubectl-dpm-profile-*.json")
			if err != nil {
				return fmt.Errorf("create temp file for profile spec: %w", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(specData); err != nil {
				tmpFile.Close()
				return fmt.Errorf("write profile spec to temp file: %w", err)
			}
			tmpFile.Close()

			if flagDebug {
				fmt.Fprintf(streams.Out, "profile spec written to temp file: %s\n", tmpFile.Name())
			}

			// nolint:gosec
			debugCommand = exec.Command(
				os.ExpandEnv(profile.Config.KubectlPath),
				"debug",
				"--namespace", namespace,
				"--custom", tmpFile.Name(),
				"--image", debugProfile.Image, targetContainer,
				"-it",
			)
		}
	} else {
		// Legacy profile field
		switch {
		case debugProfile.IsBuiltInProfile():
			// nolint:gosec
			debugCommand = exec.Command(
				os.ExpandEnv(profile.Config.KubectlPath),
				"debug",
				"--namespace", namespace,
				"--profile", debugProfile.Profile,
				"--image", debugProfile.Image, targetContainer,
				"-it",
			)
		default:
			// nolint:gosec
			debugCommand = exec.Command(
				os.ExpandEnv(profile.Config.KubectlPath),
				"debug",
				"--namespace", namespace,
				"--custom", os.ExpandEnv(debugProfile.Profile),
				"--image", debugProfile.Image, targetContainer,
				"-it",
			)
		}
	}

	debugCommand.Env = os.Environ()
	// kubectl feature flag DebugCustomProfile got dropped in 1.34
	// explicitly set it to true to support kubectl versions < 1.34
	debugCommand.Env = append(debugCommand.Env, string("KUBECTL_DEBUG_CUSTOM_PROFILE=true"))

	debugCommand.Stdout = streams.Out
	debugCommand.Stderr = streams.ErrOut
	debugCommand.Stdin = streams.In

	if flagDebug {
		fmt.Fprintf(streams.Out, "Running command: %s\n", debugCommand.String())
	}

	if err := debugCommand.Run(); err != nil {
		return fmt.Errorf("validate profile %q: %w", flagProfileName, err)
	}

	return nil
}

func getTargetPod(ctx context.Context, namespace string) (string, error) {
	restClient, err := MatchVersionKubeConfigFlags.ToRESTConfig()
	if err != nil {
		return "", fmt.Errorf("get REST config: %w", err)
	}

	podClient := corev1client.NewForConfigOrDie(restClient)

	matchingPods, err := podClient.Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(
			&metav1.LabelSelector{
				MatchLabels: debugProfile.MatchLabels,
			}),
	})
	if err != nil {
		return "", fmt.Errorf("list pods in namespace %q: %w", namespace, err)
	}

	if len(matchingPods.Items) == 0 {
		return "", fmt.Errorf("no pods in namespace %s found with label selector %v", namespace, debugProfile.MatchLabels)
	}

	var podName string

	for _, pod := range matchingPods.Items {
		podName = pod.Name
	}

	return podName, nil
}

func getTargetNamespace() string {
	if debugProfile.Namespace != "" {
		return debugProfile.Namespace
	}

	kubectlNamespace, _, _ := MatchVersionKubeConfigFlags.ToRawKubeConfigLoader().Namespace()

	if kubectlNamespace != corev1.NamespaceDefault {
		return kubectlNamespace
	}

	return corev1.NamespaceDefault
}
