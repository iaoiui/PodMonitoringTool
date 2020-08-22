package main

import (
	"context"
	"flag"
	"fmt"
	log "log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// namespace which is observed
var namespace string = getEnv("OBSERVED_NAMESPACE", "default")

// observe period (sec)
var observePeriod = 10

// Teams endpoint
// !! Replace TEAMS_ENDPOINT like "https://outlook.office.com/webhook/XXXX" with your endpoint !!
// TODO Erace Specific endpoint
var teamsEndpoint string = getEnv("TEAMS_ENDPOINT", "")

// (Optional) teamsHeartbeatEndpoint is a endpoint where this tool alert when all pod works successfully
var teamsHeartbeatEndpoint string = getEnv("TEAMS_HEARTBEAT_ENDPOINT", "")

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

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
	for {
		listPod(clientset)

		time.Sleep(time.Duration(observePeriod) * time.Second)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func listPod(clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	notReadyPods := getNotReadyPods(pods)

	// There is no NotReady Pods
	if len(notReadyPods) == 0 {
		log.Println("All Pod work succesfully")
		if teamsHeartbeatEndpoint != "" {
			// send heartbeat to teams at every observation period
			sendAlertToTeams("PodMonitoringTool Heartbeat", "All Pod work succesfully", teamsHeartbeatEndpoint)
		}

	} else {
		msg := generateAlertMsg(notReadyPods)
		sendAlertToTeams("Pod Defect Alert", msg, teamsEndpoint)
	}

}

func generateAlertMsg(pods []v1.Pod) string {
	msg := ""
	log.Printf("%v pods is not running \n", len(pods))
	msg += fmt.Sprintf("# **%v pods is not running** \n", len(pods)) + "\n"
	for i, p := range pods {
		log.Println(i + 1)
		log.Println("\t", p.Namespace, "\t", p.Name, "\t")
		msg += fmt.Sprintln("\t Namespace: \t", p.Namespace, ", Pod: \t", p.Name) + "\n"
	}
	return msg
}

// getNotReadyPods returns notReady Pods
func getNotReadyPods(pods *v1.PodList) []v1.Pod {
	notReadyPods := []v1.Pod{}
	for _, p := range pods.Items {
		if p.Status.Phase != "Running" {
			// Pod is Not Ready
			notReadyPods = append(notReadyPods, p)
		} else {
			// Container is Not Ready
			if hasNotReadyContainer(p) {
				notReadyPods = append(notReadyPods, p)
			}
		}
	}
	return notReadyPods
}

// Identify that all container inside given pod are Running
func hasNotReadyContainer(p v1.Pod) bool {
	for _, containerStatus := range p.Status.ContainerStatuses {
		if containerStatus.Ready == false {
			return true
		}
	}
	return false
}

func getKubeConfig() *rest.Config {
	// Inside Cluster Config
	if exists("/run/secrets/kubernetes.io/serviceaccount") {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		return config
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

func sendAlertToTeams(title, msg, endpoint string) {
	b := fmt.Sprintf(`{ "title": "%v", "text": "%v"}`, title, msg)
	body := strings.NewReader(b)
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
}
