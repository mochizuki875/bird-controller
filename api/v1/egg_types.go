/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EggSpec defines the desired state of Egg
type EggSpec struct {
	Parent string `json:"parent,omitempty"`
}

// EggStatus defines the observed state of Egg
type EggStatus struct {
	Parent string `json:"parent,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:JSONPath=".status.parent",name=Parent,type=string

// Egg is the Schema for the eggs API
type Egg struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EggSpec   `json:"spec,omitempty"`
	Status EggStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EggList contains a list of Egg
type EggList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Egg `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Egg{}, &EggList{})
}
