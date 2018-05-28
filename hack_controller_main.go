/*
Copyright 2017 

*/
package main

import (
	"hack_custom_controller/crd_controller"
	hack_schema "hack_custom_controller/crd_schema/hack_schema"
	"hack_custom_controller/util"
	"flag"
	"fmt"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"time"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Plural:   hack_schema.HackPlural,
		Singular: hack_schema.HackSingular,
		Group:    hack_schema.HackGroup,
		Kind:     hack_schema.HackKind,
	}

	//crd_controller.NewHackCustomResourceDefinition to create HK CRD
	err = crd_controller.NewHackCustomResourceDefinition(cs, &hkCrdDef)
	if err != nil {
		panic(err)
	}

	// Wait for the CRD to be created before we use it (only needed if its a new one)
	time.Sleep(2 * time.Second)

	// Create a queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Hackcontroller")

	CreatedHackTabObj(config)

	hkindexer, hkinformer := crd_controller.NewHKIndexerInformer(config, queue)

	controller := crd_controller.NewHackController(queue, hkindexer, hkinformer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	fmt.Println("Waiting for ever")
	// Wait forever
	select {}
}

func createHackObjSpec(NodeName string) *hack_schema.Hackcrd {

        objele := &hack_schema.Hackcrd {
		//TypeMeta: meta_v1.TypeMeta{
		//	Kind:"HackCrd", 
		//	},
                ObjectMeta: meta_v1.ObjectMeta{
                        Name: NodeName + "-hack",
                        },
                Spec : hack_schema.Hackspec{
                        HackLabel: map[string]string{ "HackKey" : "HackValue" },
                        },
                }

        return objele
}

func CreatedHackTabObj(config *rest.Config) {
	// Create a new clientset which include our CRD schema
        crdcs, scheme, err := hack_schema.NewHackClient(config)
        if err != nil {
                panic(err)
        }

        // Create a CRD client interface
        hkcrdclient := hack_schema.HackClient(crdcs, scheme, "default")

	//Create a HK CRD Helper object
        h_inf, cli := util.Getk8sClientHelper(config)

	nodeList, _ := h_inf.ListNode(cli)

        for _, ele := range nodeList.Items {
                fmt.Println("Ganesh node name is ", ele.Name)
                objSpec := createHackObjSpec(ele.Name)
		result, err := hkcrdclient.Create(objSpec)
        	if err == nil {
                	fmt.Printf("CREATED: %#v\n", result)
        	} else {
                	fmt.Printf("FATAL ERROR : %#v\n", result)
                	panic(err)
        	}
        }
}
