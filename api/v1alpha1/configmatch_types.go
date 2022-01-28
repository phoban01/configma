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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	FieldManagerName   = "configmatch-controller"
	LabelMatcher       = "config.matcher.io/group"
	LabelSelector      = "config.matcher.io/auto-update-for-group"
	UpdateAnnotation   = "config.matcher.io/version"
	MatchLabelIndexKey = ".metadata.matchLabel"
)

// ConfigMatchSpec defines the desired state of ConfigMatch
type ConfigMatchSpec struct {
	//+kubebuilder:validation:Required
	SourceRef Source `json:"sourceRef"`

	//+kubebuilder:validation:Required
	Target Target `json:"target"`
}

// Source defines the source of the config to match
type Source struct {
	//+kubebuilder:validation:Required
	Kind string `json:"kind"`

	//+kubebuilder:validation:Required
	Pattern string `json:"pattern"`

	Namespace string `json:"namespace,omitempty"`

	//+kubebuilder:validation:Optional
	MatchGroup string `json:"matchGroup,omitempty"`
}

// Target defines the source of the config to match
type Target struct {
	//+kubebuilder:validation:Required
	Kind string `json:"kind"`

	//+kubebuilder:validation:Required
	Name string `json:"name"`

	Namespace string `json:"namespace"`
}

// ConfigMatchStatus defines the observed state of ConfigMatch
type ConfigMatchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigMatch is the Schema for the configmatches API
type ConfigMatch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigMatchSpec   `json:"spec,omitempty"`
	Status ConfigMatchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigMatchList contains a list of ConfigMatch
type ConfigMatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMatch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigMatch{}, &ConfigMatchList{})
}
