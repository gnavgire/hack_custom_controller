#Hack Custom Controller

This is a custom K8S Controller, that provides api extenstions,
so that the worker nodes can be lableled accordinglly

To build the controller binary:
go build

To execute the binary in a K8S cluster:
./hack_custom_controller --kubeconf=/root/.kube/config

To check the succesfull creation of CRD inside the cluster:
# kubectl get crd
NAME                        AGE
hackcrds.hack.aricent.com   24m

