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

// BirdSpec defines the desired state of Bird
type BirdSpec struct {
	EggNumbers *int32 `json:"eggNumbers,omitempty"`
}

// BirdStatus defines the observed state of Bird
type BirdStatus struct {
	EggNumbers int32 `json:"eggNumbers,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:JSONPath=".spec.eggNumbers",name=Desire_Eggs,type=integer
// +kubebuilder:printcolumn:JSONPath=".status.eggNumbers",name=Ready_Eggs,type=integer

// Bird is the Schema for the birds API
type Bird struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BirdSpec   `json:"spec,omitempty"`
	Status BirdStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BirdList contains a list of Bird
type BirdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bird `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bird{}, &BirdList{})
}
