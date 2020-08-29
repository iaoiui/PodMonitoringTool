package main

import (
	"flag"
	log "log"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getKubeConfig() *rest.Config {
	// Inside Cluster Config
	if exists("/run/secrets/kubernetes.io/serviceaccount") {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		log.Printf("Run on inside cluster\n")
		return config
	} else {
		log.Printf("Run on outside cluster\n")
	}
	// Outside Cluster Config
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return config
}
func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
