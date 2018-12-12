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
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	//"sigs.k8s.io/controller-runtime/alpha/patterns/addon"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	api "sigs.k8s.io/cluster-api-provider-gcp/operators/knative-operator/pkg/apis/addons/v1alpha1"
	addons "sigs.k8s.io/controller-runtime/alpha/patterns/addon/pkg/apis/v1alpha1"
	"sigs.k8s.io/controller-runtime/alpha/patterns/declarative"
	"sigs.k8s.io/controller-runtime/alpha/patterns/declarative/pkg/manifest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		declarative.WithStatus(&knativeStatus{mgr: mgr}),
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

	// TODO: It'd be easier to watch for an istio CRD
	// Watch for changes to istio-system/istio-sidecar-injector so we Reconcile
	// when it appears after failing preflight.

	isPilot := func(m metav1.Object) bool {
		return m.GetNamespace() == "istio-system" && m.GetName() == "istio-sidecar-injector"
	}

	filter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isPilot(e.Meta)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isPilot(e.Meta)
		},
		UpdateFunc:  func(event.UpdateEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	mapToSingleton := &handler.EnqueueRequestsFromMapFunc{ToRequests: &mapper{}}

	dep := &appsv1.Deployment{}
	err = c.Watch(&source.Kind{Type: dep}, &handler.EnqueueRequestForObject{}, filter)
	if err != nil {
		return err
	}

	// Watch for changes to other Knative CRDs
	err = c.Watch(&source.Kind{Type: &api.KnativeBuild{}}, mapToSingleton)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &api.KnativeServing{}}, mapToSingleton)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &api.KnativeMonitoring{}}, mapToSingleton)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &api.KnativeIstio{}}, mapToSingleton)
	if err != nil {
		return err
	}
	return nil
}

type mapper struct {
}

// We can't use the typical owner ref mapping because we're going cross namepsace.
// This implementation always maps a request to the root Knative object.
func (mapper) Map(handler.MapObject) []reconcile.Request {
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "kube-system", Name: "default"}}
	return []reconcile.Request{req}
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

// hardcoded for demo!
var objs = []struct {
	key client.ObjectKey
	obj addons.CommonObject
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

func deleteCrossNamespace(ctx context.Context, c client.Client) error {
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

type knativeStatus struct {
	mgr manager.Manager
}

func (ks *knativeStatus) Reconciled(ctx context.Context, o declarative.DeclarativeObject, m *manifest.Objects) error {
	log.Info("on reconcile knative status!")

	c := ks.mgr.GetClient()

	obj, ok := o.(*api.Knative)
	if !ok {
		return fmt.Errorf("expected resource to be Knative but was: %T", o)
	}

	var errs []string
	for _, target := range objs {
		if err := c.Get(ctx, target.key, target.obj); err != nil {
			errs = append(errs, fmt.Sprintf("can not find %s", target.key))
			continue
		}

		if !target.obj.GetCommonStatus().Healthy {
			errs = append(errs, fmt.Sprintf("%s is not healthy", target.key))
		}
	}

	cs := addons.CommonStatus{Errors: errs}
	if len(errs) == 0 {
		cs.Healthy = true
	}

	if !reflect.DeepEqual(cs, obj.GetCommonStatus()) {
		log.Info("updating Knative status")
		obj.SetCommonStatus(cs)
		if err := c.Update(ctx, obj); err != nil {
			log.Error(err, "updating preflight status")
			return err
		}
	}

	return nil
}
func (ks *knativeStatus) Preflight(ctx context.Context, o declarative.DeclarativeObject) error {
	c := ks.mgr.GetClient()
	obj, ok := o.(*api.Knative)
	if !ok {
		return fmt.Errorf("expected resource to be Knative but was: %T", o)
	}

	// Don't check for Preflight on Delete
	if obj.GetDeletionTimestamp() != nil {
		return nil
	}

	// Check for istio installation
	key := client.ObjectKey{Namespace: "istio-system", Name: "istio-sidecar-injector"}
	dep := &appsv1.Deployment{}

	cs := addons.CommonStatus{}
	if err := c.Get(ctx, key, dep); err != nil {
		if errors.IsNotFound(err) {
			cs.Errors = []string{"deployment istio-system/istio-sidecar-injector not found"}
		} else {
			cs.Errors = []string{fmt.Sprintf("fetching istio-sidecar-injector: %v", err)}
		}
	} else {
		depHealthy := false
		for _, cond := range dep.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				depHealthy = true
			}
		}

		if !depHealthy {
			cs.Errors = []string{fmt.Sprintf("deployment (%s) does not meet condition: %s", key, appsv1.DeploymentAvailable)}
		}
	}

	if len(cs.Errors) != 0 {
		cs.Healthy = false
		if !reflect.DeepEqual(cs, obj.GetCommonStatus()) {
			obj.SetCommonStatus(cs)
			if err := c.Update(ctx, obj); err != nil {
				log.Error(err, "updating preflight status")
				return err
			}
		}

		return fmt.Errorf("%v", cs)
	}

	return nil
}
