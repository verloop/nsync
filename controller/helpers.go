package controller

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strconv"
)

func apply(action Action, clientset *kubernetes.Clientset, namespace string, object metav1.Object) {
	if action == SKIP {
		return
	}
	log_prefix := fmt.Sprintf("[%s] %T->%s: %s:", namespace, object, object.GetName(), ActionName[action])

	var err error
	switch action {
	case ENSURE:
		object = prepareObject(object)
		if object == nil {
			log.Println(log_prefix, "Skipping")
			return
		}

		err = update(clientset, namespace, object)
		if err != nil {
			if statusErr, ok := err.(*apierrors.StatusError); ok && statusErr.Status().Code == 404 {
				err = create(clientset, namespace, object)
			}
			if err != nil {
				log.Println(log_prefix, "Error:", err)
			}
		}
	case REMOVE:
		err = remove(clientset, namespace, object)
		if err != nil {
			log.Println(log_prefix, "Error:", err)
		}
	}
	if err == nil {
		log.Println(log_prefix, "Successful")
	}
}

func shouldManage(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	managedAnnotationValue, foundAnnotation := obj.GetAnnotations()[VERLOOP_MANAGED_KEY]
	sm, err := strconv.ParseBool(managedAnnotationValue)
	if err != nil && foundAnnotation {
		log.Printf("Warning: %s has bad value of managed. Expected bool, found %s\nError from ParseBool:%s\n", obj.GetSelfLink(), managedAnnotationValue, err)
		// Should manage is false, logic below will handle the cases well.
	}
	return sm
}

func prepareObject(object metav1.Object) metav1.Object {

	if shouldManage(object) {
		object.SetNamespace("")
		object.SetSelfLink("")
		object.SetCreationTimestamp(metav1.Time{time.Now()})
		annotations := object.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
		annotations[VERLOOP_MANAGED_KEY] = "true"
		object.SetAnnotations(annotations)
		object.SetUID("")
		object.SetResourceVersion("")
		return object
	}
	return nil

}
