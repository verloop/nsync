package controller

import (
	"time"

	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func apply(action Action, clientset *kubernetes.Clientset, namespace string, object metav1.Object) {
	if action == SKIP {
		return
	}
	entry := logrus.WithFields(logrus.Fields{
		"namespace":     namespace,
		"name":          object.GetName(),
		"action":        ActionName[action],
		"source_object": object.GetSelfLink(),
	})

	var err error
	switch action {
	case ENSURE:
		object = prepareObject(object)
		if object == nil {
			entry.Info("skipping")
			return
		}

		err = update(clientset, namespace, object)
		if err != nil {
			if statusErr, ok := err.(*apierrors.StatusError); ok && statusErr.Status().Code == 404 {
				err = create(clientset, namespace, object)
			}
			if err != nil {
				entry.WithError(err).Error("couldn't ensure")
			}
		}
	case REMOVE:
		err = remove(clientset, namespace, object)
		if err != nil {
			entry.WithError(err).Error("couldn't remove")
		}
	}
	if err == nil {
		entry.Info("successful")
	}
}

func shouldManage(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	managedAnnotationValue, foundAnnotation := obj.GetAnnotations()[VerloopManagedKey]
	sm, err := strconv.ParseBool(managedAnnotationValue)
	if err != nil && foundAnnotation {
		logrus.WithError(err).WithFields(logrus.Fields{
			"self_link": obj.GetSelfLink(),
			"expected":  "bool",
			"got":       managedAnnotationValue,
		}).Warn("bad value of `managed`")
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
		annotations[VerloopManagedKey] = "true"
		object.SetAnnotations(annotations)
		object.SetUID("")
		object.SetResourceVersion("")
		return object
	}
	return nil

}
