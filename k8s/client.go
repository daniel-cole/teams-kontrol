package k8s

import (
	"flag"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var Client *kubernetes.Clientset

func CreateClient() (err error) {

	var kubeconfig *string

	logrus.Info("Attempting to load kube config")
	// first attempt to load in cluster config - if that doesn't work, use kube config
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Info("Failed to load in-cluster kube config. Will attempt to try kubeconfig file...")
		if home := os.Getenv("HOME"); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return err
		}
	}

	Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	logrus.Info("Successfully loaded kube config")
	return nil

}
