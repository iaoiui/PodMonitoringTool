package main

import (
	"context"
	"fmt"
	log "log"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

	// https://github.com/iaoiui/PodMonitoringTool/issues/1
	observeReplicaNumbers(clientset)

}

// observe replica number of deployment and statefulsets
func observeReplicaNumbers(clientset *kubernetes.Clientset) {
	// get not ready deployment
	notReadyDeployments := getNotReadyDeployments(clientset)
	if len(notReadyDeployments) != 0 {
		msg := generateAlertMsgForDeployment(notReadyDeployments)
		sendAlertToTeams("Deployment Defect Alert", msg, teamsEndpoint)
	} else {
		// heartbeat
	}

	// get not ready statefulsets
	notReadyStatefulsets := getNotReadyStatefulsets(clientset)
	if len(notReadyStatefulsets) != 0 {
		msg := generateAlertMsgForStatefulset(notReadyStatefulsets)
		sendAlertToTeams("Statefulsets Defect Alert", msg, teamsEndpoint)
	} else {
		// heartbeat
	}
}

func getNotReadyStatefulsets(clientset *kubernetes.Clientset) []appsv1.StatefulSet {
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	notReadyStatefulsets := []appsv1.StatefulSet{}

	for _, s := range statefulsets.Items {
		desiredReplicas := s.Status.Replicas
		availableReplicas := s.Status.ReadyReplicas
		if desiredReplicas != availableReplicas {
			notReadyStatefulsets = append(notReadyStatefulsets, s)
		}

	}

	return notReadyStatefulsets
}

func getNotReadyDeployments(clientset *kubernetes.Clientset) []appsv1.Deployment {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	notReadyDeployments := []appsv1.Deployment{}

	for _, d := range deployments.Items {
		desiredReplicas := d.Status.Replicas
		availableReplicas := d.Status.AvailableReplicas
		if desiredReplicas != availableReplicas {
			notReadyDeployments = append(notReadyDeployments, d)
		}

	}

	return notReadyDeployments
}

func generateAlertMsgForStatefulset(statefulsets []appsv1.StatefulSet) string {
	msg := ""
	log.Printf("%v statefulsets is not ready \n", len(statefulsets))
	msg += fmt.Sprintf("# **%v statefulset is not ready** \n", len(statefulsets)) + "\n"
	for i, sts := range statefulsets {
		log.Println(i + 1)
		log.Println("\t", sts.Name, "\t")
		msg += fmt.Sprintln("\t Namespace: \t", sts.Namespace, "StatefulSet: \t", sts.Name, ", availableReplicas: \t", sts.Status.ReadyReplicas, ", desiredReplicas: \t", sts.Status.Replicas) + "\n"
	}
	return msg
}

func generateAlertMsgForDeployment(deployments []appsv1.Deployment) string {
	msg := ""
	log.Printf("%v deployments is not ready \n", len(deployments))
	msg += fmt.Sprintf("# **%v deployment is not ready** \n", len(deployments)) + "\n"
	for i, deploy := range deployments {
		log.Println(i + 1)
		log.Println("\t", deploy.Name, "\t")
		msg += fmt.Sprintln("\t Namespace: \t", deploy.Namespace, "Deployment: \t", deploy.Name, ", availableReplicas: \t", deploy.Status.AvailableReplicas, ", desiredReplicas: \t", deploy.Status.Replicas) + "\n"
	}
	return msg
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
