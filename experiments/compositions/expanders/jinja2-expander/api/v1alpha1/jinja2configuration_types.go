/*
Copyright 2024.

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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Jinja2ConfigurationSpec defines the desired state of Jinja2Configuration
type Jinja2ConfigurationSpec struct {
	Template string `json:"template"`
}

// Jinja2ConfigurationStatus defines the observed state of Jinja2Configuration
type Jinja2ConfigurationStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Jinja2Configuration is the Schema for the jinja2 expander
type Jinja2Configuration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Jinja2ConfigurationSpec   `json:"spec,omitempty"`
	Status Jinja2ConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Jinja2ConfigurationList contains a list of Jinja2Configuration
type Jinja2ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Jinja2Configuration `json:"items"`
}

// Status helpers
func (g *Jinja2ConfigurationStatus) ClearCondition(condition ConditionType) {
	meta.RemoveStatusCondition(&g.Conditions, string(condition))
}

func init() {
	SchemeBuilder.Register(&Jinja2Configuration{}, &Jinja2ConfigurationList{})
}
