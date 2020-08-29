package main

import (
	log "log"
	"strconv"
	"time"

	"k8s.io/client-go/kubernetes"
)

// namespace which is observed
var namespace string = getEnv("OBSERVED_NAMESPACE", "default")

// observe period (sec)
var observePeriod = 10

// Teams endpoint
// !! Replace TEAMS_ENDPOINT like "https://outlook.office.com/webhook/XXXX" with your endpoint !!
var teamsEndpoint string = getEnv("TEAMS_ENDPOINT", "")

// (Optional) teamsHeartbeatEndpoint is a endpoint where this tool alert when all pod works successfully
var teamsHeartbeatEndpoint string = getEnv("TEAMS_HEARTBEAT_ENDPOINT", "")

func main() {
	if teamsEndpoint == "" {
		log.Printf("TEAMS_ENDPOINT in not set\n")
		log.Fatal("please set TEAMS_ENDPOINT\n")
		return
	}

	if teamsHeartbeatEndpoint != "" {
		// send heartbeat to teams at every observation period
		log.Printf("Heartbeat will be notified every %v seconds\n", observePeriod)
	}

	observePeriod, err := strconv.Atoi(getEnv("OBSERVE_PERIOD", string(10)))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("OBSERVE_PERIOD is ", observePeriod)
	// get kubeConfig from Home Directory
	config := getKubeConfig()

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// main loop
	for {
		listPod(clientset)
		time.Sleep(time.Duration(observePeriod) * time.Second)
	}
}
