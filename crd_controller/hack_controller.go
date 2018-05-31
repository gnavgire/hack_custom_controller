package crd_controller

import (
         "fmt"
         "time"
	"hack_custom_controller/util"
        "github.com/golang/glog"
	"hack_custom_controller/crd_schema/hack_schema"
	"k8s.io/client-go/rest"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/apimachinery/pkg/util/runtime"
        "k8s.io/client-go/util/workqueue"
        "k8s.io/client-go/tools/cache"
)

type HackController struct {
        indexer  cache.Indexer
        informer cache.Controller
        queue workqueue.RateLimitingInterface
}

func NewHackController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *HackController {
        return &HackController{
                informer: informer,
                indexer:  indexer,
                queue:    queue,
        }
}

func (c *HackController) processNextItem() bool {
        // Wait until there is a new item in the working queue
        key, quit := c.queue.Get()
        if quit {
                return false
        }
        // Tell the queue that we are done with processing this key. This unblocks the key for other workers
        // This allows safe parallel processing because two CRD with the same key are never processed in
        // parallel.
        defer c.queue.Done(key)

        // Invoke the method containing the business logic
        err := c.syncFromQueue(key.(string))
	if err == nil {
                c.queue.Forget(key)
                return true
        }
        // Handle the error if something went wrong during the execution of the business logic
        c.handleErr(err, key)
        return true
}

//processHackQueue 
func (c *HackController) processHackQueue(key string) error {
	fmt.Printf("processHackQueue for Key %#v \n", key)
	return nil
}

// syncFromQueue is the business logic of the controller. In this controller it simply prints
// information about the CRD to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *HackController) syncFromQueue(key string) error {
        obj, exists, err := c.indexer.GetByKey(key)
        if err != nil {
                glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
                return err
        }

        if !exists {
                // Below we will warm up our cache with a CDR, so that we will see a delete for one CRD
                fmt.Printf("TL CRD object %s does not exist anymore\n", key)
        } else {
                // Note that you also have to check the uid if you have a local controlled resource, which
                // is dependent on the actual instance, to detect that a CRD object was recreated with the same name
                fmt.Printf("Sync/Add/Update for CRD %#v \n", obj)
		c.processHackQueue(key)
        }
        return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *HackController) handleErr(err error, key interface{}) {
        if err == nil {
                // Forget about the #AddRateLimited history of the key on every successful synchronization.
                // This ensures that future processing of updates for this key is not delayed because of
                // an outdated error history.
                c.queue.Forget(key)
                return
        }

        // This controller retries 5 times if something goes wrong. After that, it stops trying.
        if c.queue.NumRequeues(key) < 5 {
                glog.Infof("Error syncing CRD %v: %v", key, err)

                // Re-enqueue the key rate limited. Based on the rate limiter on the
                // queue and the re-enqueue history, the key will be processed later again.
                c.queue.AddRateLimited(key)
                return
        }

        c.queue.Forget(key)
        // Report to an external entity that, even after several retries, we could not successfully process this key
        runtime.HandleError(err)
        glog.Infof("Dropping CRD %q out of the queue: %v", key, err)
}

func (c *HackController) Run(threadiness int, stopCh chan struct{}) {
        defer runtime.HandleCrash()

        // Let the workers stop when we are done
        defer c.queue.ShutDown()
        glog.Info("Starting Trust Tab controller")

        go c.informer.Run(stopCh)

        // Wait for all involved caches to be synced, before processing items from the queue is started
        if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
                runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
                return
        }

        for i := 0; i < threadiness; i++ {
                go wait.Until(c.runWorker, time.Second, stopCh)
        }

        <-stopCh
        glog.Info("Stopping Hack Tab controller")
}

func (c *HackController) runWorker() {
        for c.processNextItem() {
        }
}

//NewHKIndexerInformer returns informer for HK CRD object
func NewHKIndexerInformer(config *rest.Config, queue workqueue.RateLimitingInterface) ( cache.Indexer, cache.Controller ) {
	// Create a new clientset which include our CRD schema
        crdcs, scheme, err := hack_schema.NewHackClient(config)
        if err != nil {
                panic(err)
        }

        // Create a CRD client interface
        hkcrdclient := hack_schema.HackClient(crdcs, scheme, "default")

	//Create a HK CRD Helper object
	h_inf, cli := util.Getk8sClientHelper(config)

	return cache.NewIndexerInformer(hkcrdclient.NewHKListWatch(), &hack_schema.Hackcrd{}, 0, cache.ResourceEventHandlerFuncs{
                AddFunc: func(obj interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(obj)
                        fmt.Println("Received Add event for ", key)
                        //myobj := obj.(*trust_schema.Trustcrd)
                        //fmt.Println("cast event name ", myobj.Name)
                        //fmt.Println("cast event spec ", myobj.Spec)
                        if err == nil {
                                queue.Add(key)
                        }
                	AddHackTabObj(obj, h_inf, cli) 
                },
                UpdateFunc: func(old interface{}, new interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(new)
                        fmt.Println("Received Update event for %#v", key)
                        if err == nil {
                                queue.Add(key)
                        }
                	UpdateHackTabObj(new, h_inf, cli) 
                },
                DeleteFunc: func(obj interface{}) {
                        // IndexerInformer uses a delta queue, therefore for deletes we have to use this
                        // key function.
                        key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
                        fmt.Println("Received delete event for %#v", key)
                        if err == nil {
                                queue.Add(key)
                        }
                },
        }, cache.Indexers{})
}
