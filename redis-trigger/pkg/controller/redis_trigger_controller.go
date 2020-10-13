package controller

import (
	"fmt"
	kubelessApi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelessversioned "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions/kubeless/v1beta1"
	kubelessutils "github.com/kubeless/kubeless/pkg/utils"
	"github.com/sirupsen/logrus"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	_ "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	redistriggerapi "github.com/kubeless/redis-trigger/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/redis-trigger/pkg/client/clientset/versioned"
	redistriggerinformers "github.com/kubeless/redis-trigger/pkg/client/informers/externalversions/kubeless/v1beta1"
	"time"
)

const (
	redisTriggerFinalizer = "kubeless.io/redistrigger"
)

// RedisTriggerConfig contains k8s client of a controller
type RedisTriggerConfig struct {
	KubeCli        kubernetes.Interface
	TriggerClient  versioned.Interface
	KubelessClient kubelessversioned.Interface
}

// RedisTriggerController object
type RedisTriggerController struct {
	logger               *logrus.Entry
	clientset            kubernetes.Interface
	httpclient           versioned.Interface
	queue                workqueue.RateLimitingInterface
	redisTriggerInformer cache.SharedIndexInformer
	functionInformer     cache.SharedIndexInformer
}

func NewRedisTriggerController(cfg RedisTriggerConfig) *RedisTriggerController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	config, err := kubelessutils.GetKubelessConfig(cfg.KubeCli, kubelessutils.GetAPIExtensionsClientInCluster())
	if err != nil {
		logrus.Fatalf("Unable to read the configmap: %s", err)
	}
	redisTriggerInformer := redistriggerinformers.NewRedisTriggerInformer(cfg.TriggerClient, config.Data["functions-namespace"], 0, cache.Indexers{})
	functionInformer := kubelessInformers.NewFunctionInformer(cfg.KubelessClient, config.Data["functions-namespace"], 0, cache.Indexers{})

	redisTriggerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newObj := new.(*redistriggerapi.RedisTrigger)
				oldObj := old.(*redistriggerapi.RedisTrigger)
				if redisTriggerObjChanged(oldObj, newObj) {
					queue.Add(key)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := RedisTriggerController{
		logger:               logrus.WithField("controller", "redis-trigger-controller"),
		clientset:            cfg.KubeCli,
		httpclient:           cfg.TriggerClient,
		redisTriggerInformer: redisTriggerInformer,
		functionInformer:     functionInformer,
		queue:                queue,
	}
	functionInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.functionAddedDeletedUpdated(obj, false)
		},
		DeleteFunc: func(obj interface{}) {
			controller.functionAddedDeletedUpdated(obj, true)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.functionAddedDeletedUpdated(new, false)
		},
	})

	return &controller
}

func redisTriggerObjChanged(oldObj, newObj *redistriggerapi.RedisTrigger) bool {
	// If the function object's deletion timestamp is set, then process
	if oldObj.DeletionTimestamp != newObj.DeletionTimestamp {
		return true
	}
	// If the new and old function object's resource version is same
	if oldObj.ResourceVersion != newObj.ResourceVersion {
		return true
	}
	newSpec := &newObj.Spec
	oldSpec := &oldObj.Spec

	if !apiequality.Semantic.DeepEqual(newSpec, oldSpec) {
		return true
	}
	return false
}

// FunctionAddedDeletedUpdated process the updates to Function objects
func (c *RedisTriggerController) functionAddedDeletedUpdated(obj interface{}, deleted bool) error {
	functionObj, ok := obj.(*kubelessApi.Function)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			err := fmt.Errorf("Couldn't get object from tombstone %#v", obj)
			c.logger.Errorf(err.Error())
			return err
		}
		functionObj, ok = tombstone.Obj.(*kubelessApi.Function)

		if !ok {
			err := fmt.Errorf("Tombstone contained object that is not a Pod %#v", obj)
			c.logger.Errorf(err.Error())
			return err
		}
	}

	if deleted {
		c.logger.Infof("Function %s deleted. Removing associated redis trigger", functionObj.Name)
		httptList, err := c.httpclient.KubelessV1beta1().RedisTriggers(functionObj.Namespace).List(nil, metav1.ListOptions{})

		if err != nil {
			return err
		}
		for _, redisTrigger := range httptList.Items {
			// todo 待确定规则
			if redisTrigger.Spec.ListKey == functionObj.Name {
				err = c.httpclient.KubelessV1beta1().RedisTriggers(functionObj.Namespace).Delete(nil, redisTrigger.Name, metav1.DeleteOptions{})
				if err != nil && !k8sErrors.IsNotFound(err) {
					c.logger.Errorf("Failed to delete httptrigger created for the function %s in namespace %s, Error: %s", functionObj.ObjectMeta.Name, functionObj.ObjectMeta.Namespace, err)
					return err
				}
			}
		}
	}

	c.logger.Infof("Successfully processed update to function object %s Namespace: %s", functionObj.Name, functionObj.Namespace)
	return nil
}

// Run starts the Trigger controller
func (c *RedisTriggerController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting HTTP Trigger controller")

	go c.redisTriggerInformer.Run(stopCh)
	go c.functionInformer.Run(stopCh)

	if !c.waitForCacheSync(stopCh) {
		return
	}

	c.logger.Info("HTTP Trigger controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *RedisTriggerController) waitForCacheSync(stopCh <-chan struct{}) bool {
	if !cache.WaitForCacheSync(stopCh, c.redisTriggerInformer.HasSynced, c.functionInformer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches required for HTTP triggers controller to sync;"))
		return false
	}
	c.logger.Info("HTTP Trigger controller caches are synced and ready")
	return true
}

func (c *RedisTriggerController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *RedisTriggerController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncRedisTrigger(key.(string))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < 5 {
		c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

// Redis trigger 核心流程
func (c *RedisTriggerController) syncRedisTrigger(key string) error {
	c.logger.Infof("Processing update to RedisTrigger: %s", key)

	_, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obj, exists, err := c.redisTriggerInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v ", key, err)
	}

	// this is an update when Redis trigger API object is actually deleted, we dont need to process anything here
	if !exists {
		c.logger.Infof("Redis Trigger %s not found in the cache, ignoring the deletion update", key)
		return nil
	}

	redisTriggerObj := obj.(*redistriggerapi.RedisTrigger)

	// Redis trigger API object is marked for deletion (DeletionTimestamp != nil), so lets process the delete update
	//if redisTriggerObj.ObjectMeta.DeletionTimestamp != nil {
	//
	//	// If finalizer is removed, then we already processed the delete update, so just return
	//	if !c.httpTriggerObjHasFinalizer(redisTriggerObj) {
	//		return nil
	//	}
	//
	//	// remove ingress resource if any. Ignore any error, as ingress resource will be GC'ed
	//	_ = httptriggerutils.DeleteIngress(c.clientset, redisTriggerObj.Name, redisTriggerObj.Namespace)
	//
	//	err = c.httpTriggerObjRemoveFinalizer(redisTriggerObj)
	//	if err != nil {
	//		c.logger.Errorf("Failed to remove HTTP trigger controller as finalizer to http trigger Obj: %s due to: %v: ", key, err)
	//		return err
	//	}
	//	c.logger.Infof("HTTP trigger object %s has been successfully processed and marked for deleteion", key)
	//	return nil
	//}
	//
	//if !c.httpTriggerObjHasFinalizer(redisTriggerObj) {
	//	err = c.httpTriggerObjAddFinalizer(redisTriggerObj)
	//	if err != nil {
	//		c.logger.Errorf("Error adding HTTP trigger controller as finalizer to  HTTPTrigger Obj: %s CRD object due to: %v: ", key, err)
	//		return err
	//	}
	//	return nil
	//}
	//
	//// create ingress resource if required
	//c.logger.Infof("Adding ingress resource for http trigger Obj: %s ", key)
	//or, err := kubelessutils.GetOwnerReference(httpTriggerKind, httpTriggerAPIVersion, redisTriggerObj.Name, redisTriggerObj.UID)
	//if err != nil {
	//	return err
	//}
	//err = httptriggerutils.CreateIngress(c.clientset, redisTriggerObj, or)
	//if err != nil && !k8sErrors.IsAlreadyExists(err) {
	//	c.logger.Errorf("Failed to create ingress rule %s corresponding to http trigger Obj: %s due to: %v: ", redisTriggerObj.Name, key, err)
	//}

	// delete ingress resource if not required
	c.logger.Infof("Processed update to HTTPTrigger: %s", key)
	return nil
}

func (c *RedisTriggerController) httpTriggerObjHasFinalizer(triggerObj *redistriggerapi.RedisTrigger) bool {
	currentFinalizers := triggerObj.ObjectMeta.Finalizers
	for _, f := range currentFinalizers {
		if f == redisTriggerFinalizer {
			return true
		}
	}
	return false
}