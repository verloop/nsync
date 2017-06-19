package main

import (
	"flag"
	"github.com/verloop/nsync/controller"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var kubeconfig string
	var tickInterval uint
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.UintVar(&tickInterval, "tickinterval", 30, "time in seconds after which we should force a resync")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("clientset error:", err)
	}
	nc := controller.NewNamespaceController(clientset, tickInterval)

	err = nc.Start()
	if err != nil {
		log.Fatalln("Start error:", err)
	}
	handleSigTerm(nc)
	log.Println("Stopped server")
}

func handleSigTerm(namespaceController *controller.NamespaceController) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	log.Println("Received SIGTERM, shutting down")

	exitCode := 0
	if err := namespaceController.Stop(); err != nil {
		log.Printf("Error during shutdown %v", err)
		exitCode = 1
	}

	log.Printf("Exiting with %v", exitCode)
	os.Exit(exitCode)

}
