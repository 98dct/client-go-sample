package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", ".kube/kubeconfig")
	if err != nil {
		panic("load config fail " + err.Error())
	}

	config.APIPath = "/api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	//加载客户端
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic("加载rest client失败" + err.Error())
	}

	pods := &corev1.PodList{}

	//查找多个pod /api/v1/namespace/{namespace}/pods
	if err = restClient.
		Get().
		Namespace("dev").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(pods); err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		fmt.Printf("namespace: %s, name: %s, status: %s, labels: %s\n", pod.Namespace, pod.Name, pod.Status.Phase, pod.GetLabels())
	}

}
