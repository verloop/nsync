package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"log"
)

func watcher(clientset *kubernetes.Clientset, namespace string, objectType ObjectType) (watch watch.Interface) {
	var err error
	switch objectType {
	case SECRET:
		watch, err = clientset.CoreV1().Secrets(namespace).Watch(metav1.ListOptions{})
	case CONFIGMAP:
		watch, err = clientset.CoreV1().ConfigMaps(namespace).Watch(metav1.ListOptions{})
	}

	if err != nil {
		log.Fatalln(err)
	}
	return

}

func update(clientset *kubernetes.Clientset, namespace string, object metav1.Object) (err error) {
	switch object.(type) {
	case *v1.Secret:
		_, err = clientset.CoreV1().Secrets(namespace).Update(object.(*v1.Secret))
	case *v1.ConfigMap:
		_, err = clientset.CoreV1().ConfigMaps(namespace).Update(object.(*v1.ConfigMap))
	}
	return err
}

func create(clientset *kubernetes.Clientset, namespace string, object metav1.Object) (err error) {
	switch object.(type) {
	case *v1.Secret:
		_, err = clientset.CoreV1().Secrets(namespace).Create(object.(*v1.Secret))

	case *v1.ConfigMap:
		_, err = clientset.CoreV1().ConfigMaps(namespace).Create(object.(*v1.ConfigMap))
	}
	return err
}

func remove(clientset *kubernetes.Clientset, namespace string, object metav1.Object) (err error) {
	switch object.(type) {
	case *v1.Secret:
		err = clientset.CoreV1().Secrets(namespace).Delete(object.GetName(), &metav1.DeleteOptions{})
	case *v1.ConfigMap:
		err = clientset.CoreV1().ConfigMaps(namespace).Delete(object.GetName(), &metav1.DeleteOptions{})

	}
	return err
}
