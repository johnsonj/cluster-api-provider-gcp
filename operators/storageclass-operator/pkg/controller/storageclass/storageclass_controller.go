/*
Copyright 2018 The Kubernetes Authors.

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
package storageclass

// Automatically generate RBAC rules to allow the Controller to read and write StorageClass
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.sigs.k8s.io,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete

import (
	api "sigs.k8s.io/cluster-api-provider-gcp/operators/storageclass-operator/pkg/apis/addons/v1alpha1"
	"sigs.k8s.io/controller-runtime/alpha/patterns/declarative"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileStorageClass{}

// ReconcileStorageClass reconciles a Dashboard object
type ReconcileStorageClass struct {
	declarative.Reconciler
}

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *ReconcileStorageClass {
	r := &ReconcileStorageClass{}

	r.Reconciler.Init(mgr, &api.StorageClass{}, "storageclass")
	//declarative.WithGroupVersionKind(api.SchemeGroupVersion.WithKind("storageclass")),

	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileStorageClass) error {
	// Create a new controller
	c, err := controller.New("storageclass-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to StorageClass
	err = c.Watch(&source.Kind{Type: &api.StorageClass{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}
