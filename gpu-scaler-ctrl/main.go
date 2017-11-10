package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var minReplicas = flag.Int("min-replicas", 20, "Minimum number of replicas for the given deployment")
var maxReplicas = flag.Int("max-replicas", 20, "Maximum number of replicas for the given deployment")
var scaleUpThreshold = flag.Int("scale-up-threshold", 20, "Threshold value before we scale up the deployment.")
var scaleDownThreshold = flag.Int("scale-down-threshold", 20, "Threshold value before we scale down the deployment.")
var listenAddr = flag.String("listen", "0.0.0.0:1110", "Listen address")
var deploymentName = flag.String("deploy-name", "", "Name of Kubernetes Deployment")
var deploymentNamespace = flag.String("deploy-namespace", "", "Name of Kubernetes Namespace for Deployment")

func main() {
	flag.Parse()
	// Lock and update Map (sync Map)
	// Each X seconds, check if the general moving avg is good or not
	// Scale up, scale down.
	// Log the decision.
	// Fetch from the Cluster Kubernetes the `replicas` for the given deployment.
	// Scale up or down / Patch for the given thing... kubernetes client-go

	http.HandleFunc("/v1/gpu_usage", func(w http.ResponseWriter, r *http.Request) {
		hostname := r.FormValue("hostname")
		gpuUsageStr := r.Formvalue("volatile_gpu_usage")
		gpuUsage, err := strconv.ParseFloat(gpuUsageStr, 64)
		if err != nil {
			http.Error(w, "invalid value for 'volatile_gpu_usage'", 500)
			return
		}

		setUsageForPod(hostname, gpuUsage)

		w.WriteHeader(200)
		w.Write([]byte("ok, thanks"))
	})

	k8sClient, err := newK8sClient()
	if err != nil {
		log.Fatalln("Couldn't get Kubernetes client:", err.Error())
	}

	go func() {
		for {
			time.Sleep(5 * time.Minute)

			var err error
			currentUsage := computeUsageAvg()
			scaleReplicasDelta := 0
			if currentUsage > *maxThreshold {
				scaleReplicasDelta = -1
			} else if currentUsage < *minThreshold {
				scaleReplicasDelta = 1
			}

			if scaleReplicasDelta != 0 {
				if err := scaleDeployment(k8sClient, scaleReplicasDelta); err != nil {
					log.Println("Couldn't scale deployment:", err)
					continue
				}
			}
		}
	}()

	if err = http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Println("Failed to listen:", err)
	}
}

func scaleDeployment(k8sClient *kubernetes.Clientset, replicasDelta int) error {
	deployIface := k8sClient.AppsV1beta2().Deployments(*deploymentNamespace)

	dep, err := deployIface.Get(*deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("couldn't retrieve deployment from k8s api server: %s", err)
	}

	if dep.Spec.Replicas == nil {
		return fmt.Errorf("replicas count not listed in the deployment %q", *deploymentName)
	}

	replicas := *dep.Spec.Replicas

	targetReplicas := replicas + replicasDelta

	if targetReplicas > *maxReplicas {
		return fmt.Errorf("requesting %d replicas, but limited by max replicas = %d", targetReplicas, *maxReplicas)
	}

	if targetReplicas < *minReplicas {
		return fmt.Errorf("requesting %d replicas, but limited by min replicas = %d", targetReplicas, *minReplicas)
	}

	patchRequest, _ := json.Marshal(map[string]interface{}{
		"op":    "replace",
		"path":  "spec.replicas",
		"value": fmt.Sprintf("%d", targetReplicas),
	})
	dep, err := deployIface.Patch(*deploymentName, types.JSONPatchType, patchRequest)

	return err
}
