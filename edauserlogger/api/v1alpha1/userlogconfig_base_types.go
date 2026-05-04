/*
Copyright 2026.

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

// UserLogConfig is the Schema for the userlogconfigs API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=userlogconfigs,scope=Cluster
type UserLogConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserLogConfigSpec   `json:"spec,omitempty"`
	Status UserLogConfigStatus `json:"status,omitempty"`
}

// UserLogConfigList contains a list of UserLogConfig
// +kubebuilder:object:root=true
type UserLogConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserLogConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserLogConfig{}, &UserLogConfigList{})
}
