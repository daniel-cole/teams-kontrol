package k8s

import "k8s.io/client-go/kubernetes"
import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func GetPods(client kubernetes.Interface, namespace string) (interface{}, error) {
	return client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func GetPod(client kubernetes.Interface, namespace string, name string) (interface{}, error) {
	return client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func DeletePod(client kubernetes.Interface, namespace string, name string) error {
	return client.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
}
