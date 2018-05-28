/*
Copyright 2017 
*/
package hack_schema

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	HackPlural   string = "hackcrds"
	HackSingular string = "hackcrd"
	HackKind     string = "Hackcrd"
	HackGroup    string = "hack.aricent.com"
	HackVersion  string = "v1beta1"
)

//HackClient returns CRD clientset required to apply watch on the CRD
func HackClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *hackclient {
	return &hackclient{cl: cl, ns: namespace, plural: HackPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type hackclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

type Hackcrd struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               Hackspec `json:"Spec"`
}

type Hackspec struct {
	HackLabel 	map[string]string `json:"HackLabel"`
}

type HackcrdList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Hackcrd `json:"items"`
}

// Create a  Rest client with the new CRD Schema
var SchemeGroupVersion = schema.GroupVersion{Group: HackGroup, Version: HackVersion}

//addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Hackcrd{},
		&HackcrdList{},
	)
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

//NewHackClient registers CRD schema and returns rest client for the CRD
func NewHackClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}
	return client, scheme, nil
}

// Create a new List watch for our HK CRD
func (f *hackclient) NewHKListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}

func (f *hackclient) Create(obj *Hackcrd) (*Hackcrd, error) {
        var result Hackcrd
        err := f.cl.Post().
                Namespace(f.ns).Resource(f.plural).
                Body(obj).Do().Into(&result)
        return &result, err
}

