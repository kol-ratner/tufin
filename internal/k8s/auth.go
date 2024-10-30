package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// You should pass an empty string for kubeconfigPath if you expect to load the kubeconfig stored at the default path: ~/.kube/config.
// If, however, you want to load a kubeconfig from a different path, pass the path to the kubeconfig file as the kubeconfigPath parameter.
func GetKubeConfigFromHost(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// use the current context in kubeconfig
	config, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, err
	}

	return config, nil
}
