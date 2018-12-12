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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addon "sigs.k8s.io/controller-runtime/alpha/patterns/addon/pkg/apis/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KnativeSpec defines the desired state of Knative
type KnativeSpec struct {
	addon.CommonSpec
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// KnativeStatus defines the observed state of Knative
type KnativeStatus struct {
	addon.CommonStatus
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

var _ addon.CommonObject = &Knative{}

func (o *Knative) ComponentName() string {
	return "knative"
}

func (o *Knative) CommonSpec() addon.CommonSpec {
	return o.Spec.CommonSpec
}

func (o *Knative) GetCommonStatus() addon.CommonStatus {
	return o.Status.CommonStatus
}

func (o *Knative) SetCommonStatus(s addon.CommonStatus) {
	o.Status.CommonStatus = s
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Knative is the Schema for the knatives API
// +k8s:openapi-gen=true
type Knative struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnativeSpec   `json:"spec,omitempty"`
	Status KnativeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KnativeList contains a list of Knative
type KnativeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Knative `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Knative{}, &KnativeList{})
}
