package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var consent bool
	var all bool
	var overwrite bool
	var kubeContext string
	var kubeConfig string

	flag.BoolVar(&consent, "yes", false, "run without confirmation")
	flag.BoolVar(&all, "all", false, "run on all nodes in context")
	flag.BoolVar(&overwrite, "overwrite", false, "overwrite conditions with default status")
	flag.StringVar(&kubeContext, "context", "", "select context/cluster")
	flag.StringVar(&kubeConfig, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "override kube config file name")

	flag.Parse()

	if flag.NArg() == 0 && !all {
		fmt.Printf("Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	var nodename string = flag.Arg(0)

	config, err := clientcmd.LoadFromFile(kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	if kubeContext == "" {
		kubeContext = config.CurrentContext
	}

	clientConfig := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{CurrentContext: kubeContext})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	var nodes []string
	if all {
		nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, node := range nodeList.Items {
			nodes = append(nodes, node.Name)
		}
	} else {
		nodes = append(nodes, nodename)
	}

	if !consent {
		if overwrite {
			fmt.Printf("Main conditions will be OVERWRITTEN with default values.\n\n")
		}
		fmt.Printf("Clear conditions for: \n\n")
		fmt.Println("Context:", kubeContext)
		fmt.Printf("Kubeconfig: %s\n", kubeConfig)
		fmt.Printf("Node(s):\n")
		for i, node := range nodes {
			fmt.Printf("%-50s", node)
			if (i+1)%3 == 0 {
				fmt.Println()
			}
		}
		if len(nodes)%3 != 0 {
			fmt.Println()
		}
		fmt.Println()
	}

	var response string
	if !consent {
		fmt.Print("Do you want to continue? [y/N] ")
		fmt.Scanln(&response)
	} else {
		response = "y"
	}

	if response != "y" && response != "Y" {
		fmt.Println("Canceled.")
		return
	}

	readyCondition := corev1.NodeCondition{
		Type:               corev1.NodeReady,
		Status:             corev1.ConditionTrue,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "KubeletReady",
		Message:            "kubelet is posting ready status",
	}
	memoryPressureCondition := corev1.NodeCondition{
		Type:               corev1.NodeMemoryPressure,
		Status:             corev1.ConditionFalse,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "KubeletHasSufficientMemory",
		Message:            "kubelet has sufficient memory available",
	}
	diskPressureCondition := corev1.NodeCondition{
		Type:               corev1.NodeDiskPressure,
		Status:             corev1.ConditionFalse,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "KubeletHasNoDiskPressure",
		Message:            "kubelet has no disk pressure",
	}
	pidPressureCondition := corev1.NodeCondition{
		Type:               corev1.NodePIDPressure,
		Status:             corev1.ConditionFalse,
		LastHeartbeatTime:  metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             "KubeletHasSufficientPID",
		Message:            "kubelet has sufficient PID available",
	}

	for _, nodename := range nodes {
		node, err := clientset.CoreV1().Nodes().Get(ctx, nodename, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		if !overwrite {
			updatedConditions := []corev1.NodeCondition{}
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady ||
					condition.Type == corev1.NodeMemoryPressure ||
					condition.Type == corev1.NodeDiskPressure ||
					condition.Type == corev1.NodePIDPressure {
					updatedConditions = append(updatedConditions, condition)
				}
			}
			node.Status.Conditions = []corev1.NodeCondition{}
			node.Status.Conditions = updatedConditions
		}

		if overwrite {
			node.Status.Conditions = []corev1.NodeCondition{}
			node.Status.Conditions = append(node.Status.Conditions, memoryPressureCondition, diskPressureCondition, pidPressureCondition, readyCondition)
		}

		_, err = clientset.CoreV1().Nodes().UpdateStatus(ctx, node, metav1.UpdateOptions{})
		if err != nil {
      fmt.Printf("Failed: %s\n", node.Name)
      continue
		}

		fmt.Printf("Cleared: %s\n", node.Name)
	}

	if !consent {
		fmt.Println("Done.")
	}
}
