// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package usagedb

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// DefaultAuthorizationType is the Authorization type used when none is set.
const DefaultAuthorizationType = "Bearer"

// Auth configures how the usage db client authenticates to the usage database.
// BasicAuth and Authorization are mutually exclusive.
//
// A credential is never written inline. On a SchedulingShard it is a key of a
// Secret in the namespace KAI runs in: the operator mounts that key into the
// scheduler pod and gives the scheduler the path of the mounted file. A
// scheduler that runs without the operator sets the file paths itself.
type Auth struct {
	// BasicAuth authenticates with a username and a password.
	// +kubebuilder:validation:Optional
	BasicAuth *BasicAuth `yaml:"basicAuth,omitempty" json:"basicAuth,omitempty"`
	// Authorization authenticates with an "Authorization: <type> <credentials>"
	// header, such as a bearer token.
	// +kubebuilder:validation:Optional
	Authorization *Authorization `yaml:"authorization,omitempty" json:"authorization,omitempty"`
}

// BasicAuth holds the sources of a basic auth username and password. The
// username is required and has exactly one source; the password is optional.
type BasicAuth struct {
	// Username is the Secret key that holds the username.
	// +kubebuilder:validation:Optional
	Username *corev1.SecretKeySelector `yaml:"username,omitempty" json:"username,omitempty"`
	// Password is the Secret key that holds the password.
	// +kubebuilder:validation:Optional
	Password *corev1.SecretKeySelector `yaml:"password,omitempty" json:"password,omitempty"`
	// UsernameFile is the path of a file that holds the username. The operator
	// sets it in the scheduler configuration from Username.
	// +kubebuilder:validation:Optional
	UsernameFile string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty"`
	// PasswordFile is the path of a file that holds the password. The operator
	// sets it in the scheduler configuration from Password.
	// +kubebuilder:validation:Optional
	PasswordFile string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty"`
}

// Authorization holds the type and the source of the credentials of an
// Authorization header. The credentials have exactly one source.
type Authorization struct {
	// Type is the authentication type. Defaults to "Bearer". "Basic" is not
	// accepted: use BasicAuth.
	// +kubebuilder:validation:Optional
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Credentials is the Secret key that holds the credentials.
	// +kubebuilder:validation:Optional
	Credentials *corev1.SecretKeySelector `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	// CredentialsFile is the path of a file that holds the credentials. The
	// operator sets it in the scheduler configuration from Credentials.
	// +kubebuilder:validation:Optional
	CredentialsFile string `yaml:"credentialsFile,omitempty" json:"credentialsFile,omitempty"`
}

// GetType returns the authentication type, or DefaultAuthorizationType if none is set.
func (a *Authorization) GetType() string {
	if a == nil || strings.TrimSpace(a.Type) == "" {
		return DefaultAuthorizationType
	}
	return strings.TrimSpace(a.Type)
}

// Validate reports whether the configuration names a usable set of credential
// sources. A nil or empty Auth is valid and means no authentication.
func (a *Auth) Validate() error {
	if a == nil {
		return nil
	}
	if a.BasicAuth != nil && a.Authorization != nil {
		return errors.New("basicAuth and authorization are mutually exclusive")
	}
	if a.BasicAuth != nil {
		if err := a.BasicAuth.validate(); err != nil {
			return fmt.Errorf("basicAuth: %w", err)
		}
	}
	if a.Authorization != nil {
		if err := a.Authorization.validate(); err != nil {
			return fmt.Errorf("authorization: %w", err)
		}
	}
	return nil
}

func (b *BasicAuth) validate() error {
	hasUsername, err := hasOneSource("username", b.Username, "usernameFile", b.UsernameFile)
	if err != nil {
		return err
	}
	if !hasUsername {
		return errors.New("one of username and usernameFile is required")
	}
	_, err = hasOneSource("password", b.Password, "passwordFile", b.PasswordFile)
	return err
}

func (a *Authorization) validate() error {
	if strings.EqualFold(a.GetType(), "Basic") {
		return errors.New(`type "Basic" is not accepted, use basicAuth`)
	}
	hasCredentials, err := hasOneSource("credentials", a.Credentials, "credentialsFile", a.CredentialsFile)
	if err != nil {
		return err
	}
	if !hasCredentials {
		return errors.New("one of credentials and credentialsFile is required")
	}
	return nil
}

// hasOneSource reports whether a value has a source, and fails if it has both
// or if its Secret reference is incomplete.
func hasOneSource(
	selectorField string, selector *corev1.SecretKeySelector, fileField string, file string,
) (bool, error) {
	if selector != nil && file != "" {
		return false, fmt.Errorf("%s and %s are mutually exclusive", selectorField, fileField)
	}
	if selector != nil && (selector.Name == "" || selector.Key == "") {
		return false, fmt.Errorf("%s needs both a Secret name and a key", selectorField)
	}
	return selector != nil || file != "", nil
}

// DeepCopy returns a copy that shares no memory with the receiver.
func (a *Auth) DeepCopy() *Auth {
	if a == nil {
		return nil
	}
	out := new(Auth)
	out.BasicAuth = a.BasicAuth.DeepCopy()
	out.Authorization = a.Authorization.DeepCopy()
	return out
}

// DeepCopy returns a copy that shares no memory with the receiver.
func (b *BasicAuth) DeepCopy() *BasicAuth {
	if b == nil {
		return nil
	}
	out := new(BasicAuth)
	*out = *b
	if b.Username != nil {
		out.Username = b.Username.DeepCopy()
	}
	if b.Password != nil {
		out.Password = b.Password.DeepCopy()
	}
	return out
}

// DeepCopy returns a copy that shares no memory with the receiver.
func (a *Authorization) DeepCopy() *Authorization {
	if a == nil {
		return nil
	}
	out := new(Authorization)
	*out = *a
	if a.Credentials != nil {
		out.Credentials = a.Credentials.DeepCopy()
	}
	return out
}
