package main

import (
	"fmt"
	"github.com/kubeless/kubeless/pkg/utils"
	kubelessutils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/kubeless/redis-trigger/pkg/client/clientset/versioned"
	"github.com/kubeless/redis-trigger/pkg/controller"
	"github.com/kubeless/redis-trigger/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{
	Use:   "redis-trigger-controller",
	Short: "Kubeless redis trigger controller",
	Long:  "Kubeless redis trigger controller",
	Run: func(cmd *cobra.Command, args []string) {

		kubelessClient, err := kubelessutils.GetFunctionClientInCluster()
		if err != nil {
			logrus.Fatalf("Cannot get kubeless CR API client: %v", err)
		}

		redisTriggerClient, err := GetTriggerClientInCluster()
		if err != nil {
			logrus.Fatalf("Cannot get Redis trigger CR API client: %v", err)
		}

		httpTriggerCfg := controller.RedisTriggerConfig{
			KubeCli:        GetClient(),
			TriggerClient:  redisTriggerClient,
			KubelessClient: kubelessClient,
		}

		redisTriggerController := controller.NewRedisTriggerController(httpTriggerCfg)

		stopCh := make(chan struct{})
		defer close(stopCh)

		go redisTriggerController.Run(stopCh)

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
	},
}

// GetTriggerClientInCluster returns function clientset to the request from inside of cluster
func GetTriggerClientInCluster() (versioned.Interface, error) {
	config, err := utils.GetInClusterConfig()
	if err != nil {
		return nil, err
	}
	redisTriggerClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return redisTriggerClient, nil
}

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := utils.GetInClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not create kubernetes client: %v", err)
	}

	return clientset
}

func main() {
	logrus.Infof("Running Kubeless Redis trigger controller version: %v", version.Version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
