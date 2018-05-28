package crd_controller

import (
	//apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//"k8s.io/client-go/kubernetes/fake"
	"apiextensions-apiserver/test/integration/testserver"
	//"meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	 trust_schema "citcrd/crd_schema/cit_trust_schema"
	 "testing"
	)

/*

type testContext struct {
        tearDown           func()
        tlc                 *garbagecollector.GarbageCollector
        clientSet          clientset.Interface
        apiExtensionClient apiextensionsclientset.Interface
        clientPool         dynamic.ClientPool
        startGC            func(workers int)
        // syncPeriod is how often the GC started with startGC will be resynced.
        syncPeriod time.Duration
}

func newCRDInstance() *trust_schema.Trusttab{
        return &trust_schema.Trusttab{
		ObjectMeta: meta_v1.ObjectMeta{
                        Name:   "geo123",
                },
                Spec: trust_schema.Trusttabspec{
                        HostList: []trust_schema.HostList{
                        {
                                Hostname : "node1",
                                Trusted : "true",
                                TrustTagSignedReport : "asdfakjhwer123jkasdf43kaNF9U4T6ASadsf",
                                TrustTagExpiry : "28-10-2017T24:00:12.943",
                                },
                        },
                },
        }
}
*/


func TestTLCRDCreation(t *testing.T) {
	masterConfig, err := rest.InClusterConfig()
	fakeClient := &fake.Clientset{}
	err := NewcitTLCustomResourceDefinition(fakeClient)
        if err != nil {
                t.Fatalf("error creating Trust Label CRD: %v", err)
        }

        t.Logf("Testing cit TL controller success")
}



