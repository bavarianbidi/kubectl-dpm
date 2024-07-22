// SPDX-License-Identifier: MIT

package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"

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

func NewCmdDebugProfile(_ genericiooptions.IOStreams) *cobra.Command {
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

			if err := run(args); err != nil {
				return err
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
	cmd.Flags().StringVarP(&flagProfileName, "profile", "p", "", "profile name")
	cmd.Flags().StringVarP(&flagImage, "image", "i", "", "image to use for the debug container")
	cmd.Flags().BoolVarP(&flagDebug, "debug", "d", false, "print debug information")

	return cmd
}

func run(args []string) error {
	// validate kubectl path
	if err := profile.ValidateKubectlPath(); err != nil {
		return err
	}

	// check kubectl version
	if err := profile.CheckKubectlVersion(); err != nil {
		return err
	}

	// validate and complete profile
	if err := profile.ValidateAndCompleteProfile(flagProfileName); err != nil {
		return err
	}

	// get the index of the profile where the profile name matches
	idx := slices.IndexFunc(profile.Config.Profiles,
		func(c profile.Profile) bool { return c.ProfileName == flagProfileName },
	)

	debugProfile = profile.Config.Profiles[idx]

	var targetContainer string
	namespace := getTargetNamespace()

	switch {
	case len(args) > 1:
		return fmt.Errorf("too many arguments")
	case len(args) == 1:
		targetContainer = args[0]
	case len(debugProfile.MatchLabels) > 0:
		var err error
		targetContainer, err = getTargetPod(namespace)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("no target container specified")
	}

	if flagDebug {
		fmt.Printf("Using profile: %+v\n", debugProfile)
		fmt.Printf("kubectl path: %s\n", os.ExpandEnv(profile.Config.KubectlPath))
		fmt.Printf("profile path: %s\n", debugProfile.CustomProfileFile)
		fmt.Printf("profile path resolved: %s\n", os.ExpandEnv(debugProfile.CustomProfileFile))
	}

	// nolint:gosec
	debugCommand := exec.Command(
		os.ExpandEnv(profile.Config.KubectlPath),
		"debug",
		"--namespace", namespace,
		"--custom", os.ExpandEnv(debugProfile.CustomProfileFile),
		"--image", debugProfile.Image, targetContainer,
		"-it",
	)
	debugCommand.Env = os.Environ()
	debugCommand.Env = append(debugCommand.Env, string(cmdutil.DebugCustomProfile)+"=true")

	debugCommand.Stdout = os.Stdout
	debugCommand.Stderr = os.Stderr
	debugCommand.Stdin = os.Stdin

	if flagDebug {
		fmt.Printf("Running command: %s\n", debugCommand.String())
	}

	if err := debugCommand.Run(); err != nil {
		return err
	}

	return nil
}

func getTargetPod(namespace string) (string, error) {
	restClient, err := MatchVersionKubeConfigFlags.ToRESTConfig()
	if err != nil {
		return "", err
	}

	podClient := corev1client.NewForConfigOrDie(restClient)

	matchingPods, err := podClient.Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(
			&metav1.LabelSelector{
				MatchLabels: debugProfile.MatchLabels,
			}),
	})
	if err != nil {
		return "", err
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
