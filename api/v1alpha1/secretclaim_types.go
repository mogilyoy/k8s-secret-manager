/*
Copyright 2025.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretClaimSpec defines the desired state of SecretClaim
type SecretClaimSpec struct {
	Type string `json:"type"`

	Data map[string]string `json:"data,omitempty"`

	Generation *GenerationConfig `json:"generation,omitempty"`
}

type GenerationConfig struct {
	Length int `json:"length"` // длина пароля

	Encoding string `json:"encoding,omitempty"` //  "hex", "base64", "alphanumeric"

	ReconcileTrigger string `json:"reconcileTrigger,omitempty"`

	DataKeys []string `json:"dataKeys,omitempty"` // ключи для сгенерированных данных
}

// SecretClaimStatus defines the observed state of SecretClaim.
type SecretClaimStatus struct {
	Synced bool `json:"synced"`

	CreatedSecretName string `json:"createdSecretName,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	LastUpdate *metav1.Time `json:"lastUpdate,omitempty"`

	LastReconcileTrigger string `json:"lastReconcileTrigger,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SecretClaim is the Schema for the secretclaims API
type SecretClaim struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SecretClaimSpec `json:"spec"`

	Status SecretClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretClaimList contains a list of SecretClaim
type SecretClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretClaim{}, &SecretClaimList{})
}
