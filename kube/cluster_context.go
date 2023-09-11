package kube

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ConfigParameters struct {
	Path    string
	Context string
	Save    bool
}

func LoadConfig(configParam ConfigParameters) (*rest.Config, error) {
	configAccess := clientcmd.NewDefaultPathOptions()

	loadingRules := *configAccess.LoadingRules
	loadingRules.Precedence = configAccess.GetLoadingPrecedence()
	if configParam.Path != "" {
		loadingRules.ExplicitPath = configParam.Path
	}

	configOverrides := clientcmd.ConfigOverrides{}
	if configParam.Context != "" {
		configOverrides.CurrentContext = configParam.Context
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&loadingRules, &configOverrides)

	if configParam.Save {
		rawConf, err := clientConfig.RawConfig()
		if err != nil {
			fmt.Printf("Error %s, getting kubeconfig", err.Error())

		} else {
			err = clientcmd.ModifyConfig(configAccess, rawConf, true)
			if err != nil {
				fmt.Printf("Error modifying kubeconfig, context will only be set for this command")
			}
		}
	}

	return clientConfig.ClientConfig()
}

func AvailableContexts(path string) []string {
	if path == "" {
		path = clientcmd.RecommendedHomeFile
	}
	config := clientcmd.GetConfigFromFileOrDie(path)

	keys := make([]string, len(config.Contexts))

	i := 0
	for k := range config.Contexts {
		keys[i] = k
		i++
	}
	return keys
}
