package crd_controller

import (
	"hack_custom_controller/util"
	"fmt"
	"time"
	//"github.com/golang/glog"
	hack_schema "hack_custom_controller/crd_schema/hack_schema"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "k8s.io/client-go/kubernetes"
)

/*
const (
	trustexpiry = "TrustTagExpiry"
	trustlabel = "trusted"
	trustsignreport = "TrustTagSignedReport"
	assetexpiry = "AssetTagExpiry"
	assetsignreport = "AssetTagSignedReport"
)
*/

type CrdDefinition struct {
	Plural   string
	Singular string
	Group    string
	Kind     string
}

//NewHackCustomResourceDefinition Creates new CIT CRD's
func NewHackCustomResourceDefinition(cs clientset.Interface, crdDef *CrdDefinition) error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: crdDef.Plural + "." + crdDef.Group},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   crdDef.Group,
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   crdDef.Plural,
				Singular: crdDef.Singular,
				Kind:     crdDef.Kind,
				//ShortNames:  []string{"tt"},
			},
			Scope: apiextensionsv1beta1.NamespaceScoped,
				
		},
	}
	_, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil && apierrors.IsAlreadyExists(err) {
		fmt.Println("Hack CRD allready exsists")
		return nil
	} else {
		if err := waitForEstablishedCRD(cs, crd.Name); err != nil {
			fmt.Println("Failed to establish CRD ")
			return err
		}
	}

	fmt.Printf("sucessfully created CRD : %v \n", crd.Name)
	return err
}

//waitForEstablishedCRD polls until the CRD gets created and ready for use
func waitForEstablishedCRD(client clientset.Interface, name string) error {
	return wait.PollImmediate(500*time.Millisecond, wait.ForeverTestTimeout, func() (bool, error) {
		crd, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, err
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					fmt.Printf("Name conflict: %v\n", cond.Reason)
				}
			}
		}
		return false, nil
	})
}

//TO-DO: get node name from hack object
//apply label from hackSpec
//getHackObjLabel creates lables and annotations map based on Hack CRD
/*
func getHackObjLabel(obj hack_schema.HostList) (util.Labels, util.Annotations) {
	var lbl = make(util.Labels, 2)
	var annotation = make(util.Annotations, 1)
	lbl[trustexpiry] = obj.TrustTagExpiry
	lbl[trustlabel] = obj.Trusted
	annotation[trustsignreport] = obj.TrustTagSignedReport

	return lbl, annotation
}
*/

//UpdateHackTabObj Handler for addition event of the TL CRD
func UpdateHackTabObj(obj interface{}, helper util.APIHelpers, cli *k8sclient.Clientset) {
	myobj := obj.(*hack_schema.Hackcrd)
	fmt.Println("cast event name ", myobj.Name)

	//nodeList, err := helper.ListNode(cli) 
	node, err := helper.GetNode(cli, myobj.Name)
	if err != nil {
		fmt.Println("failed to get node: %s", err.Error())
		return
	}

	lbl := myobj.Spec.HackLabel 
	ann := map[string]string{"Gnaesh_annotate" : "fortimebeing"}
	helper.AddLabelsAnnotations(node, lbl, ann)
	err = helper.UpdateNode(cli, node)
	if err != nil {
		fmt.Println("can't update node: %s", err.Error())
		return
	}
}
