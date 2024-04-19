package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	//clientset就是基于rest-client的，按照group和version再次封装成客户端 例如: appsv1.interface  appsv1Beatv1等等
	//这些客户端又集成了deployment, statefulset等客户端对象（提供get，post等方法）,这些方法底层是调用了restclient的get，post方法
	//client-gen自动生成
	config, err := clientcmd.BuildConfigFromFlags("", ".kube/kubeconfig")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	list, err := clientset.AppsV1().Deployments("dev").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, deployment := range list.Items {
		fmt.Printf("namespace: %s, name: %s, historyLimit: %d\n", deployment.Namespace, deployment.Name, *deployment.Spec.RevisionHistoryLimit)
	}

}
