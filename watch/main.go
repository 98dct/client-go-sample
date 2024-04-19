package main

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", ".kube/kubeconfig")
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	watcher, err := clientset.AppsV1().
		StatefulSets("dev").
		Watch(
			context.TODO(),
			metav1.ListOptions{
				// 标签选择器
				LabelSelector: "sts=nginx",
			},
		)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				fmt.Println("channel closed")
				break
			}
			fmt.Println("Event Type:", event.Type)
			sts, ok := event.Object.(*appsv1.StatefulSet)
			if !ok {
				fmt.Println("not sts")
				continue
			}
			fmt.Printf("namespace: %s, name: %s\n", sts.Namespace, sts.Name)
		}
	}
}
