package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/fields"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/cache"

	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
)

type Options struct {
	Debug        bool   `short:"d" long:"debug" description:"Log debug messages"`
	Port         uint16 `short:"p" long:"port" description:"Port to listen on for /health" default:"8080"`
  SourceSecret string `short:"s" long:"sourcesecret" description:"Source secret name to copy data from" default:"default-tls"`
  TargetSecret string `short:"t" long:"targetsecret" description:"Targed secret name to create" default:"default-tls"`
	Version      bool   `short:"v" long:"version" description:"Show version"`
}

var version = "No version set!"

var opts Options
var myNamespace string
var parser = flags.NewParser(&opts, flags.Default)
var log = logging.MustGetLogger("default")
var logFormatter = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{level:.5s} (%{shortfile})%{color:reset} %{message}",
)

func createSecret(cs *kubernetes.Clientset, ns *v1.Namespace) error {
	log.Debug("Retrieving source secret")
	sourceSecret, err := cs.Core().Secrets(myNamespace).Get(opts.SourceSecret)
	if err != nil {
		return err
	}

	log.Debugf("Creating new secret in ns %s", ns.ObjectMeta.Name)
	_, err = cs.Core().Secrets(ns.ObjectMeta.Name).Create(&v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: opts.TargetSecret,
		},
		Data: sourceSecret.Data,
	})
	log.Infof("Created new secret in ns %s", ns.ObjectMeta.Name)
	return err
}

func checkSecretExists(cs *kubernetes.Clientset, ns *v1.Namespace) bool {
	_, err := cs.Core().Secrets(ns.ObjectMeta.Name).Get(opts.TargetSecret)
	return (err == nil)
}

func namespaceCreated(cs *kubernetes.Clientset, obj interface{}) error {
	ns := obj.(*v1.Namespace)
	log.Debug("Namespace created - %s", ns)
	if !checkSecretExists(cs, ns) {
		return createSecret(cs, ns)
	}
	return nil
}

func watchNamespaces(cs *kubernetes.Clientset) (cache.Store, chan struct{}) {
	watchlist := cache.NewListWatchFromClient(
		*cs.CoreClient,
		"namespaces",
		v1.NamespaceAll,
		fields.Everything(),
	)

	resyncPeriod := 5 * time.Second

	store, controller := cache.NewInformer(
		watchlist,
		&v1.Namespace{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				namespaceCreated(cs, obj)
			},
		},
	)

	// Run till we stop
	stop := make(chan struct{})
	go controller.Run(stop)

	return store, stop
}

func getNamespace() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
	}
	return "", err
}

func main() {
	// Setup logging
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetFormatter(logFormatter)
	logBackendLev := logging.AddModuleLevel(logBackend)
	logBackendLev.SetLevel(logging.ERROR, "default")
	logging.SetBackend(logBackendLev)

	// Parse flags
	_, err := parser.Parse()
	if err != nil {
    if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
      os.Exit(0)
    } else {
      log.Error(err)
      os.Exit(1)
    }
	}

	// Enable debug logging if debug flag is on
	if opts.Debug {
		logging.SetLevel(logging.DEBUG, "default")
	}
	log.Debug("Debug logging enabled")

	// Print version and exit if version flag is on
	if opts.Version {
		println(version)
		return
	}

	log.Infof("Starting - version %s", version)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	myNamespace, err = getNamespace()
	log.Infof("Detected namespace - %s", myNamespace)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Info("Connected to Kubernetes API")

	nsStore, stop := watchNamespaces(clientset)
	defer close(stop)
	log.Info("Watching for namespaces")

	log.Info("Going through existing namespaces")
	for _, obj := range nsStore.List() {
		ns := obj.(*v1.Namespace)
		log.Info("Ns - %s", ns.ObjectMeta.Name)
		if !checkSecretExists(clientset, ns) {
			createSecret(clientset, ns)
		}
	}

	log.Info("Listening on port 8080")
	log.Error(http.ListenAndServe(":8080", nil))
}
