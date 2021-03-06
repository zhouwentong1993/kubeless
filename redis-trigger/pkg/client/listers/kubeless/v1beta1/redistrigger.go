/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/kubeless/redis-trigger/pkg/apis/kubeless/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// RedisTriggerLister helps list RedisTriggers.
// All objects returned here must be treated as read-only.
type RedisTriggerLister interface {
	// List lists all RedisTriggers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.RedisTrigger, err error)
	// RedisTriggers returns an object that can list and get RedisTriggers.
	RedisTriggers(namespace string) RedisTriggerNamespaceLister
	RedisTriggerListerExpansion
}

// redisTriggerLister implements the RedisTriggerLister interface.
type redisTriggerLister struct {
	indexer cache.Indexer
}

// NewRedisTriggerLister returns a new RedisTriggerLister.
func NewRedisTriggerLister(indexer cache.Indexer) RedisTriggerLister {
	return &redisTriggerLister{indexer: indexer}
}

// List lists all RedisTriggers in the indexer.
func (s *redisTriggerLister) List(selector labels.Selector) (ret []*v1beta1.RedisTrigger, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.RedisTrigger))
	})
	return ret, err
}

// RedisTriggers returns an object that can list and get RedisTriggers.
func (s *redisTriggerLister) RedisTriggers(namespace string) RedisTriggerNamespaceLister {
	return redisTriggerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// RedisTriggerNamespaceLister helps list and get RedisTriggers.
// All objects returned here must be treated as read-only.
type RedisTriggerNamespaceLister interface {
	// List lists all RedisTriggers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.RedisTrigger, err error)
	// Get retrieves the RedisTrigger from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.RedisTrigger, error)
	RedisTriggerNamespaceListerExpansion
}

// redisTriggerNamespaceLister implements the RedisTriggerNamespaceLister
// interface.
type redisTriggerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all RedisTriggers in the indexer for a given namespace.
func (s redisTriggerNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.RedisTrigger, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.RedisTrigger))
	})
	return ret, err
}

// Get retrieves the RedisTrigger from the indexer for a given namespace and name.
func (s redisTriggerNamespaceLister) Get(name string) (*v1beta1.RedisTrigger, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("redistrigger"), name)
	}
	return obj.(*v1beta1.RedisTrigger), nil
}
