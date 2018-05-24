package controller

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	minTickInterval = 10
	maxTickInterval = 300
)

// NewNamespaceController - Creates a NamespaceController given a k8s clientset and a tick interval
func NewNamespaceController(clientset *kubernetes.Clientset, tickInterval uint) *NamespaceController {
	managed := make(map[ObjectType]map[string]bool)
	managed[NAMESPACE] = make(map[string]bool)
	managed[CONFIGMAP] = make(map[string]bool)
	managed[SECRET] = make(map[string]bool)
	if tickInterval < minTickInterval {
		tickInterval = minTickInterval
	}
	if tickInterval > maxTickInterval {
		tickInterval = maxTickInterval
	}

	// Check if namespace is given from outside
	namespace := os.Getenv("POD_NAMESPACE")
	// Check if we are inside a cluster and can find the file
	if namespace == "" {
		ns, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		namespace = string(ns)
	}
	if namespace == "" {
		namespace = "default"
	}

	return &NamespaceController{
		watchNamespace: namespace,
		clientSet:      clientset,
		managed:        managed,
		stopChan:       make(chan bool),
		TickInterval:   tickInterval,
	}
}

// NamespaceController - Periodically sync Secrets and ConfigMaps from current namespace to all namespaces.
type NamespaceController struct {
	watchNamespace string
	clientSet      *kubernetes.Clientset
	managed        map[ObjectType]map[string]bool
	stopChan       chan bool
	TickInterval   uint
	sync.Mutex
}

// Start - Start running the watcher
func (n *NamespaceController) Start() error {
	go n.ticker(n.stopChan, n.TickInterval)
	return nil
}

// Stop watching the watcher
func (n *NamespaceController) Stop() error {
	close(n.stopChan)
	return nil
}

func (n *NamespaceController) ticker(stop chan bool, tickInterval uint) {
	tick := time.NewTicker(time.Duration(tickInterval) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			logrus.Info("got a tick")
			namespaces, err := n.clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				logrus.WithError(err).Error("list namespaces failed")
				continue
			}

			secretList, secretListErr := n.clientSet.CoreV1().Secrets(n.watchNamespace).List(metav1.ListOptions{})
			if secretListErr != nil {
				logrus.WithError(secretListErr).Error("list secrets failed")
			}
			filteredSecrets := make([]v1.Secret, 0, len(secretList.Items))
			for _, secret := range secretList.Items {
				if shouldManage(&secret) {
					filteredSecrets = append(filteredSecrets, secret)
				}
			}

			configMapList, configMapListErr := n.clientSet.CoreV1().ConfigMaps(n.watchNamespace).List(metav1.ListOptions{})
			if configMapListErr != nil {
				logrus.WithError(configMapListErr).Error("list configmaps failed")
			}

			filteredConfigMaps := make([]v1.ConfigMap, 0, len(configMapList.Items))
			for _, configmap := range configMapList.Items {
				if shouldManage(&configmap) {
					filteredConfigMaps = append(filteredConfigMaps, configmap)
				}
			}

			for _, namespace := range namespaces.Items {
				if !shouldManage(&namespace) {
					continue
				}
				for _, secret := range filteredSecrets {
					apply(ENSURE, n.clientSet, namespace.GetName(), &secret)
				}

				for _, configmap := range filteredConfigMaps {
					apply(ENSURE, n.clientSet, namespace.GetName(), &configmap)
				}
			}
			logrus.Info("tick successful")
		case <-stop:
			logrus.Info("got stop signal")
			return
		}
	}
}
