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

// KnativeBuildSpec defines the desired state of KnativeBuild
type KnativeBuildSpec struct {
	addon.CommonSpec
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// KnativeBuildStatus defines the observed state of KnativeBuild
type KnativeBuildStatus struct {
	addon.CommonStatus
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

var _ addon.CommonObject = &KnativeBuild{}

func (o *KnativeBuild) ComponentName() string {
	return "knativebuild"
}

func (o *KnativeBuild) CommonSpec() addon.CommonSpec {
	return o.Spec.CommonSpec
}

func (o *KnativeBuild) GetCommonStatus() addon.CommonStatus {
	return o.Status.CommonStatus
}

func (o *KnativeBuild) SetCommonStatus(s addon.CommonStatus) {
	o.Status.CommonStatus = s
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KnativeBuild is the Schema for the knativebuilds API
// +k8s:openapi-gen=true
type KnativeBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KnativeBuildSpec   `json:"spec,omitempty"`
	Status KnativeBuildStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KnativeBuildList contains a list of KnativeBuild
type KnativeBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KnativeBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KnativeBuild{}, &KnativeBuildList{})
}
