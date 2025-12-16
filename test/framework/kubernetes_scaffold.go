package framework

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesScaffold struct {
	kubectlOptions *k8s.KubectlOptions
	clientset      *kubernetes.Clientset
}

func NewKubernetesScaffold(opts KubectlOptions) (*KubernetesScaffold, error) {
	kubectlOptions := k8s.NewKubectlOptions(opts.ContextName, opts.ConfigPath, opts.Namespace)
	restConfig, err := buildRestConfig(opts.ContextName)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &KubernetesScaffold{
		kubectlOptions: kubectlOptions,
		clientset:      clientset,
	}, nil
}

type KubectlOptions struct {
	ContextName string
	ConfigPath  string
	Namespace   string
}

// buildRestConfig builds the rest.Config object from kubeconfig filepath and
// context, if kubeconfig is missing, building from in-cluster configuration.
func buildRestConfig(context string) (*rest.Config, error) {

	// Config loading rules:
	// 1. kubeconfig if it not empty string
	// 2. Config(s) in KUBECONFIG environment variable
	// 3. In cluster config if running in-cluster
	// 4. Use $HOME/.kube/config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  context,
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return clientConfig.ClientConfig()
}
