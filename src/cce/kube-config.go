package cce

import (
	"fmt"
	. "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"otc-auth/src/common"
	"otc-auth/src/config"
	"strings"
)

func getKubeConfig(kubeConfigParams KubeConfigParams) string {
	println("Getting kube config...")

	clusterId, err := getClusterId(kubeConfigParams.ClusterName, kubeConfigParams.ProjectName)
	if err != nil {
		common.OutputErrorToConsoleAndExit(err, "fatal: error receiving cluster id: %s")
	}

	response := getClusterCertFromServiceProvider(kubeConfigParams.ProjectName, clusterId, kubeConfigParams.DaysValid)

	return string(common.GetBodyBytesFromResponse(response))
}

func mergeKubeConfig(projectName string, clusterName string, kubeConfigData string) {
	kubeConfigContextData := addContextInformationToKubeConfig(projectName, clusterName, kubeConfigData)
	currentConfig, err := NewDefaultClientConfigLoadingRules().GetStartingConfig()
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}

	clientConfig, err := NewClientConfigFromBytes([]byte(kubeConfigContextData))
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}
	kubeConfig, err := clientConfig.RawConfig()
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}

	filenameNewFile := "kubeConfig_new"
	filenameCurrentFile := "kubeConfig_current"

	err = WriteToFile(kubeConfig, filenameNewFile)
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}
	err = WriteToFile(*currentConfig, filenameCurrentFile)
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}

	loadingRules := ClientConfigLoadingRules{
		Precedence: []string{filenameNewFile, filenameCurrentFile},
	}

	mergedConfig, err := loadingRules.Load()
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}
	err = WriteToFile(*mergedConfig, homedir.HomeDir()+"/.kube/config")
	if err != nil {
		common.OutputErrorToConsoleAndExit(err)
	}

	os.RemoveAll(filenameNewFile)
	os.RemoveAll(filenameCurrentFile)
}

func addContextInformationToKubeConfig(projectName string, clusterName string, kubeConfigData string) string {
	cloud := config.GetActiveCloudConfig()

	kubeConfigData = strings.ReplaceAll(kubeConfigData, "internalCluster", fmt.Sprintf("%s/%s-intranet", projectName, clusterName))
	kubeConfigData = strings.ReplaceAll(kubeConfigData, "externalCluster", fmt.Sprintf("%s/%s", projectName, clusterName))
	kubeConfigData = strings.ReplaceAll(kubeConfigData, "internal", fmt.Sprintf("%s/%s-intranet", projectName, clusterName))
	kubeConfigData = strings.ReplaceAll(kubeConfigData, "external", fmt.Sprintf("%s/%s", projectName, clusterName))
	kubeConfigData = strings.ReplaceAll(kubeConfigData, ":\"user\"", fmt.Sprintf(":\"%s\"", cloud.Username))

	return kubeConfigData
}
