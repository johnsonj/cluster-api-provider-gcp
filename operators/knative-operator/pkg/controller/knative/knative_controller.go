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

package knative

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	//"sigs.k8s.io/controller-runtime/alpha/patterns/addon"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	api "sigs.k8s.io/cluster-api-provider-gcp/operators/knative-operator/pkg/apis/addons/v1alpha1"
	"sigs.k8s.io/controller-runtime/alpha/patterns/declarative"
	"sigs.k8s.io/controller-runtime/alpha/patterns/declarative/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

// Add creates a new Knative Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r := &ReconcileKnative{}

	r.Reconciler.Init(mgr, &api.Knative{}, "knative",
		declarative.WithPreserveNamespace(),
		declarative.WithObjectTransform(handleKnativeLifecycle(mgr)),
	)

	c, err := controller.New("knative-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Knative
	err = c.Watch(&source.Kind{Type: &api.Knative{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to deployed objects
	_, err = declarative.WatchAll(mgr.GetConfig(), c, r, declarative.SourceLabel)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKnative{}

// +kubebuilder:rbac:groups=addons.sigs.k8s.io,resources=corednss,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// ReconcileKnative reconciles a Knative object
type ReconcileKnative struct {
	declarative.Reconciler
	client.Client
	scheme *runtime.Scheme
}

const (
	operatorFinalizer = "operator.knative.sig.addons.k8s.io"
)

func handleKnativeLifecycle(mgr manager.Manager) declarative.ObjectTransform {
	return func(ctx context.Context, o declarative.DeclarativeObject, m *manifest.Objects) error {
		obj, ok := o.(*api.Knative)
		if !ok {
			return fmt.Errorf("expected resource to be Knative but was: %T", o)
		}
		c := mgr.GetClient()
		finalizers := obj.GetFinalizers()
		finalizerSet := sets.NewString(finalizers...)
		if o.GetDeletionTimestamp() != nil {
			log.Info("Running finalizer on Knative")
			if err := deleteCrossNamespace(ctx, c); err != nil {
				return err
			}

			log.Info("Removing Finalizer")
			finalizerSet.Delete(operatorFinalizer)
			obj.SetFinalizers(finalizerSet.UnsortedList())
			if err := c.Update(ctx, obj); err != nil {
				return fmt.Errorf("could not finalize Knative: %v", err)
			}
			m.Items = nil
			return nil
		}
		if !finalizerSet.Has(operatorFinalizer) {
			// Ensure Finalizer has been registered
			obj.SetFinalizers(append(finalizers, operatorFinalizer))
			if err := c.Update(ctx, obj); err != nil {
				return fmt.Errorf("could not add finalizer to Knative: %v", err)
			}
		}
		return nil
	}
}

func deleteCrossNamespace(ctx context.Context, c client.Client) error {
	// hardcoded for demo!
	objs := []struct {
		key client.ObjectKey
		obj runtime.Object
	}{
		{
			key: client.ObjectKey{Name: "default", Namespace: "knative-build"},
			obj: &api.KnativeBuild{},
		},
		{
			key: client.ObjectKey{Name: "default", Namespace: "istio-system"},
			obj: &api.KnativeIstio{},
		},
		{
			key: client.ObjectKey{Name: "default", Namespace: "knative-monitoring"},
			obj: &api.KnativeMonitoring{},
		},
		{
			key: client.ObjectKey{Name: "default", Namespace: "knative-serving"},
			obj: &api.KnativeServing{},
		},
	}

	for _, target := range objs {
		if err := c.Get(ctx, target.key, target.obj); err != nil {
			if errors.IsNotFound(err) {
				continue
			} else {
				return err
			}
		}
		if err := c.Delete(ctx, target.obj); err != nil {
			return err
		}
	}

	return nil
}
