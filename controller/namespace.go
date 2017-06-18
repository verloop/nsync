package controller

import (
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"sync"
	"time"
)

func NewNamespaceController(clientset *kubernetes.Clientset, tickInterval uint8) *NamespaceController {
	managed := make(map[ObjectType]map[string]bool)
	managed[NAMESPACE] = make(map[string]bool)
	managed[CONFIGMAP] = make(map[string]bool)
	managed[SECRET] = make(map[string]bool)
	if tickInterval < 20 {
		tickInterval = 20
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

type NamespaceController struct {
	watchNamespace string
	clientSet      *kubernetes.Clientset
	managed        map[ObjectType]map[string]bool
	stopChan       chan bool
	TickInterval   uint8
	sync.Mutex
}

func (n *NamespaceController) Start() error {
	go n.watchNs(n.stopChan)
	go n.watchResource(n.stopChan, SECRET)
	go n.watchResource(n.stopChan, CONFIGMAP)
	go n.ticker(n.stopChan, n.TickInterval)
	return nil
}

func (n *NamespaceController) Stop() error {
	close(n.stopChan)
	return nil
}

func (n *NamespaceController) ticker(stop chan bool, tickInterval uint8) {
	tick := time.NewTicker(time.Duration(tickInterval) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			namespaces, err := n.clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				log.Println("List Namespaces error", err)
				continue
			}

			secretList, secretListErr := n.clientSet.CoreV1().Secrets(n.watchNamespace).List(metav1.ListOptions{})
			if secretListErr != nil {
				log.Println("List secrets error", secretListErr)
			}
			filteredSecrets := make([]metav1.Object, 0, len(secretList.Items))
			for _, secret := range secretList.Items {
				if shouldManage(&secret) {
					filteredSecrets = append(filteredSecrets, &secret)
				}
			}

			configMapList, configMapListErr := n.clientSet.CoreV1().ConfigMaps(n.watchNamespace).List(metav1.ListOptions{})
			if configMapListErr != nil {
				log.Println("List configmaps error", configMapListErr)
			}

			filteredConfigMaps := make([]metav1.Object, 0, len(configMapList.Items))
			for _, configmap := range configMapList.Items {
				if shouldManage(&configmap) {
					filteredConfigMaps = append(filteredConfigMaps, &configmap)
				}
			}

			for _, namespace := range namespaces.Items {
				if !shouldManage(&namespace) {
					continue
				}
				for _, secret := range filteredSecrets {
					apply(ENSURE, n.clientSet, namespace.GetName(), secret)
				}

				for _, configmap := range filteredConfigMaps {
					apply(ENSURE, n.clientSet, namespace.GetName(), configmap)
				}
			}
		case <-stop:
			return
		}
	}
}

func (n *NamespaceController) watchResource(stop chan bool, objType ObjectType) {

	resourceWatcher := watcher(n.clientSet, n.watchNamespace, objType)
	log.Println("Start watching", ObjectName[objType])

	resultChan := resourceWatcher.ResultChan()
	for {
		select {
		case event := <-resultChan:
			resource, action := n.objHandler(objType, event)
			if action == SKIP || resource == nil {
				continue
			}
			namespaces, err := n.clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				log.Println("List Namespaces error", err)
				continue
			}

			for _, namespace := range namespaces.Items {
				if !shouldManage(&namespace) {
					continue
				}
				apply(action, n.clientSet, namespace.GetName(), resource)
			}
		case <-stop:
			return
		}
	}
}

func (n *NamespaceController) watchNs(stop chan bool) {
	namespaces, err := n.clientSet.CoreV1().Namespaces().Watch(metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Start watching namespaces")
	resultChan := namespaces.ResultChan()
	for {
		select {
		case event := <-resultChan:
			ns, action := n.objHandler(NAMESPACE, event)
			if action == SKIP {
				continue
			}
			secretList, err := n.clientSet.CoreV1().Secrets(n.watchNamespace).List(metav1.ListOptions{})
			if err == nil {
				for _, secret := range secretList.Items {
					apply(action, n.clientSet, ns.GetName(), &secret)
				}
			} else {
				log.Println("List secrets error", err)
			}

			configMapList, err := n.clientSet.CoreV1().ConfigMaps(n.watchNamespace).List(metav1.ListOptions{})
			if err == nil {
				for _, configmap := range configMapList.Items {
					apply(action, n.clientSet, ns.GetName(), &configmap)
				}
			} else {
				log.Println("List configmap error", err)
			}

		case <-stop:
			return
		}
	}
}

func (n *NamespaceController) objHandler(objType ObjectType, event watch.Event) (obj metav1.Object, action Action) {
	if event.Object == nil {
		return
	}
	n.Lock()
	defer n.Unlock()
	obj = event.Object.(metav1.Object)
	isManaged := n.managed[objType][obj.GetName()]

	shouldManage := shouldManage(obj) && event.Type != watch.Deleted

	if shouldManage && (!isManaged || (isManaged && event.Type == watch.Modified)) {
		n.managed[objType][obj.GetName()] = true
		action = ENSURE
	}

	if !shouldManage && isManaged {
		delete(n.managed[objType], obj.GetName())
		action = REMOVE
	}
	//log.Printf("Type: %T  Name: %s  Should manage: %t, Is Managed: %t, event Type: %s Action: %s\n",obj,obj.GetName(),shouldManage,isManaged,event.Type,ActionName[action])

	return
}
