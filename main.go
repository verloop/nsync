package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/verloop/nsync/controller"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	var tickInterval uint
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.UintVar(&tickInterval, "tickinterval", 30, "time in seconds after which we should force a resync")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.WithError(err).Fatalln("failed to build config from flag")
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.WithError(err).Fatalln("clientset error")
	}
	nc := controller.NewNamespaceController(clientset, tickInterval)
	logrus.Info("Starting server")
	err = nc.Start()
	if err != nil {
		logrus.WithError(err).Fatalln("start error")
	}
	handleSigTerm(nc)
	logrus.Info("Stopped server")
}

func handleSigTerm(namespaceController *controller.NamespaceController) {
	deathRay := make(chan os.Signal, 2)
	signal.Notify(deathRay, syscall.SIGINT, syscall.SIGTERM)
	<-deathRay
	logrus.Info("shutting down")

	exitCode := 0
	if err := namespaceController.Stop(); err != nil {
		logrus.WithError(err).Info("error during shutdown")
		exitCode = 1
	}

	logrus.WithField("exit_code", exitCode).Info("exiting")
	os.Exit(exitCode)

}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Output to stdout instead of the default stderr
	logrus.SetOutput(os.Stdout)

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
