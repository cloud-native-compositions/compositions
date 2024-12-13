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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ConditionType defines the type of ManagedConfigSync condition
type ConditionType string

// The valid conditions of Compositions
const (
	Ready ConditionType = "Ready"
	// Error implies the last reconcile attempt failed
	Error ConditionType = "Error"
	// Validation implies the validation failed
	ValidationFailed ConditionType = "ValidationFailed"
	// Waiting - Plan is waiting for values to progress
	Waiting ConditionType = "Waiting"
)

// Schema represents the attributes that define an instance of
// a resourcegroup.
type Schema struct {
	// The group of the resourcegroup. This is used to generate
	// and create the CRD for the resourcegroup.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="kind is immutable"
	Group string `json:"group,omitempty"`
	// The kind of the resourcegroup. This is used to generate
	// and create the CRD for the resourcegroup.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="kind is immutable"
	Kind string `json:"kind,omitempty"`
	// The APIVersion of the resourcegroup. This is used to generate
	// and create the CRD for the resourcegroup.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="apiVersion is immutable"
	APIVersion string `json:"apiVersion,omitempty"`
	// The spec of the resourcegroup. Typically, this is the spec of
	// the CRD that the resourcegroup is managing. This is adhering
	// to the SimpleSchema spec
	Spec runtime.RawExtension `json:"spec,omitempty"`
	// The status of the resourcegroup. This is the status of the CRD
	// that the resourcegroup is managing. This is adhering to the
	// SimpleSchema spec.
	Status runtime.RawExtension `json:"status,omitempty"`
	// Validation is a list of validation rules that are applied to the
	// resourcegroup.
	// Not implemented yet.
	Validation []string `json:"validation,omitempty"`
}

type Jinja2 struct {
	Template string `json:"template"`
}

// ConfigReference - For BYO Expanders, we can extend it
type ConfigReference struct {
	//+kubebuilder:validation:Required
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type ExpanderConfig struct {
	// Built in expanders
	Jinja2 *Jinja2 `json:"jinja2,omitempty"`
	// For BYO Expanders use generic template or ref for external config
	Template  string           `json:"template,omitempty"`
	ConfigRef *ConfigReference `json:"configref,omitempty"`
}

type Expander struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// Type indicates what expander to use
	//   jinja - jinja2 expander
	//   ...
	// +kubebuilder:default=jinja2
	Type string `json:"type"`
	// +kubebuilder:default=latest
	Version string `json:"version,omitempty"`

	// TODO (barney-s): Make ConfigReference the only way to specify and dont have any inline expander configs
	//  This would make the UX experience uniform.
	ExpanderConfig `json:""`
}

type NamespaceMode string

const (
	// NamespaceModeNone is when nothing is set, this is the same as Inherit
	NamespaceModeNone NamespaceMode = ""
	// NamespaceModeInherit implies all the objects namespace is replaced with the  input api object's namespace
	NamespaceModeInherit NamespaceMode = "inherit"
	// NamespaceModeExplicit implies the objects in the template must have its namespace set
	NamespaceModeExplicit NamespaceMode = "explicit"
)

// ReadyOn defines ready condition for a GVK
type ReadyOn struct {
	//+kubebuilder:validation:Required
	Group   string `json:"group"`
	Version string `json:"version,omitempty"`
	//+kubebuilder:validation:Required
	Kind      string `json:"kind"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	//+kubebuilder:validation:Required
	Ready string `json:"readyIf"`
}

// CompositionSpec defines the desired state of Composition
type CompositionSpec struct {
	// NOTE: Tighten the Composition API to include fields that are used in the controller
	//  As we add features we can uncomment these fields
	//Name           string     `json:"name"`
	//Namespace      string     `json:"namespace"`
	//InputName      string     `json:"inputName,omitempty"`
	//InputNamespace string     `json:"inputNamespace,omitempty"`
	//Sinc      Sinc       `json:"sinc,omitempty"`

	Description string `json:"description,omitempty"`

	// TODO (barney -s) rename to FacadeAPIGroup,facadeAPIGroup

	// Use existing KRM API
	InputAPIGroup string `json:"inputAPIGroup,omitempty"`

	// The schema of the resourcegroup, which includes the
	// apiVersion, kind, spec, status, types, and some validation
	// rules.
	//
	// +kubebuilder:validation:Required
	Schema *Schema `json:"schema,omitempty"`

	//+kubebuilder:validation:MinItems=1
	Expanders []Expander `json:"expanders"`
	// Namespace mode indicates how compositions set the namespace of the objects from expanders.
	// ""|inherit implies inherit the facade api's namespace. Only namespaced objects are allowed.
	// explicit     implies the objects in the template must have the namespace set.
	// +kubebuilder:validation:Enum=inherit;explicit
	NamespaceMode NamespaceMode `json:"namespaceMode,omitempty"`

	// Readiness
	Readiness []ReadyOn `json:"readiness,omitempty"`
}

type ValidationStatus string

const (
	// ValidationStatusUnkown is when it is not validated
	ValidationStatusUnknown ValidationStatus = "unknown"
	// ValidationStatusSuccess is when valdiation succeeds
	ValidationStatusSuccess ValidationStatus = "success"
	// ValidationStatusFailed is when valdiation fails
	ValidationStatusFailed ValidationStatus = "failed"
	// ValidationStatusError is when validation was not called
	ValidationStatusError ValidationStatus = "error"
)

// StageStatus captures the status of a stage
type StageValidationStatus struct {
	ValidationStatus ValidationStatus `json:"validationStatus,omitempty"`
	Reason           string           `json:"reason,omitempty"`
	Message          string           `json:"message,omitempty"`
}

// CompositionStatus defines the observed state of Composition
type CompositionStatus struct {
	Generation int64                            `json:"generation,omitempty"`
	Conditions []metav1.Condition               `json:"conditions,omitempty"`
	Stages     map[string]StageValidationStatus `json:"stages,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Composition is the Schema for the compositions API
type Composition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositionSpec   `json:"spec,omitempty"`
	Status CompositionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CompositionList contains a list of Composition
type CompositionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Composition `json:"items"`
}

// Status helpers
func (s *CompositionStatus) ClearCondition(condition ConditionType) {
	meta.RemoveStatusCondition(&s.Conditions, string(condition))
}

func init() {
	SchemeBuilder.Register(&Composition{}, &CompositionList{})
}
