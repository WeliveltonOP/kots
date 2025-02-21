/*
Copyright 2019 Replicated, Inc..

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

package v1beta1

import (
	"github.com/replicatedhq/kots/kotskinds/multitype"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IdentitySpec struct {
	IdentityIssuerURL           string                 `json:"identityIssuerURL" yaml:"identityIssuerURL"`
	OIDCRedirectURIs            []string               `json:"oidcRedirectUris" yaml:"oidcRedirectUris"`
	OAUTH2AlwaysShowLoginScreen bool                   `json:"oauth2AlwaysShowLoginScreen,omitempty" yaml:"oauth2AlwaysShowLoginScreen,omitempty"`
	SigningKeysExpiration       string                 `json:"signingKeysExpiration,omitempty" yaml:"signingKeysExpiration,omitempty"`
	IDTokensExpiration          string                 `json:"idTokensExpiration,omitempty" yaml:"idTokensExpiration,omitempty"`
	SupportedProviders          []string               `json:"supportedProviders,omitempty" yaml:"supportedProviders,omitempty"`
	RequireIdentityProvider     multitype.BoolOrString `json:"requireIdentityProvider" yaml:"requireIdentityProvider"`
	Roles                       []IdentityRole         `json:"roles,omitempty" yaml:"roles,omitempty"`
	WebConfig                   *IdentityWebConfig     `json:"webConfig,omitempty" yaml:"webConfig,omitempty"`
}

type IdentityRole struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type IdentityWebConfig struct {
	Title string                  `json:"title,omitempty" yaml:"title,omitempty"`
	Theme *IdentityWebConfigTheme `json:"theme,omitempty" yaml:"theme,omitempty"`
}

type IdentityWebConfigTheme struct {
	StyleCSSBase64 string `json:"styleCssBase64,omitempty" yaml:"styleCssBase64,omitempty"`
	LogoURL        string `json:"logoUrl,omitempty" yaml:"logoUrl,omitempty"`
	LogoBase64     string `json:"logoBase64,omitempty" yaml:"logoBase64,omitempty"`
	FaviconBase64  string `json:"faviconBase64,omitempty" yaml:"faviconBase64,omitempty"`
}

// IdentityStatus defines the observed state of Identity
type IdentityStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Identity is the Schema for the identity document
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Identity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentitySpec   `json:"spec,omitempty"`
	Status IdentityStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IdentityList contains a list of Identities
type IdentityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Identity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Identity{}, &IdentityList{})
}
