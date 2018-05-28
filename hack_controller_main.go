/*
Copyright 2017 

*/
package main

import (
	"hack_custom_controller/crd_controller"
	hack_schema "hack_custom_controller/crd_schema/hack_schema"
	"flag"
	"fmt"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"time"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {

	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	hkCrdDef := crd_controller.CrdDefinition{
		Plural:   hack_schema.CITTLPlural,
		Singular: hack_schema.CITTLSingular,
		Group:    hack_schema.CITTLGroup,
		Kind:     hack_schema.CITTLKind,
	}

	//crd_controller.NewCitCustomResourceDefinition to create HK CRD
	err = crd_controller.NewCitCustomResourceDefinition(cs, &hkCrdDef)
	if err != nil {
		panic(err)
	}

	// Wait for the CRD to be created before we use it (only needed if its a new one)
	time.Sleep(2 * time.Second)

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Hackcontroller")

	hkindexer, hkinformer := crd_controller.NewHKIndexerInformer(config, queue)

	controller := crd_controller.NewHackController(queue, hkindexer, hkinformer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	fmt.Println("Waiting for ever")
	// Wait forever
	select {}
}
