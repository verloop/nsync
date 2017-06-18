package main

import (
	"flag"
	"github.com/verloop/nSync/controller"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var kubeconfig *string

	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("clientset error:", err)
	}
	nc := controller.NewNamespaceController(clientset, 255)

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
