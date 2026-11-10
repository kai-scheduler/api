// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package usagedb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func secretKey(name, key string) *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{Name: name},
		Key:                  key,
	}
}

func TestAuth_Validate(t *testing.T) {
	tests := []struct {
		name    string
		auth    *Auth
		wantErr string
	}{
		{name: "nil auth means no authentication", auth: nil},
		{name: "empty auth means no authentication", auth: &Auth{}},
		{
			name: "basic auth from secret keys",
			auth: &Auth{BasicAuth: &BasicAuth{
				Username: secretKey("usage-db", "username"),
				Password: secretKey("usage-db", "password"),
			}},
		},
		{
			name: "basic auth from files",
			auth: &Auth{BasicAuth: &BasicAuth{UsernameFile: "/auth/username", PasswordFile: "/auth/password"}},
		},
		{
			name: "basic auth without a password",
			auth: &Auth{BasicAuth: &BasicAuth{Username: secretKey("usage-db", "username")}},
		},
		{
			name: "bearer token from a secret key",
			auth: &Auth{Authorization: &Authorization{Credentials: secretKey("usage-db", "token")}},
		},
		{
			name: "custom authorization type from a file",
			auth: &Auth{Authorization: &Authorization{Type: "Token", CredentialsFile: "/auth/credentials"}},
		},
		{
			name: "basic auth and authorization together",
			auth: &Auth{
				BasicAuth:     &BasicAuth{Username: secretKey("usage-db", "username")},
				Authorization: &Authorization{Credentials: secretKey("usage-db", "token")},
			},
			wantErr: "basicAuth and authorization are mutually exclusive",
		},
		{
			name:    "basic auth without a username",
			auth:    &Auth{BasicAuth: &BasicAuth{Password: secretKey("usage-db", "password")}},
			wantErr: "basicAuth: one of username and usernameFile is required",
		},
		{
			name: "username from both a secret key and a file",
			auth: &Auth{BasicAuth: &BasicAuth{
				Username:     secretKey("usage-db", "username"),
				UsernameFile: "/auth/username",
			}},
			wantErr: "basicAuth: username and usernameFile are mutually exclusive",
		},
		{
			name: "password from both a secret key and a file",
			auth: &Auth{BasicAuth: &BasicAuth{
				Username:     secretKey("usage-db", "username"),
				Password:     secretKey("usage-db", "password"),
				PasswordFile: "/auth/password",
			}},
			wantErr: "basicAuth: password and passwordFile are mutually exclusive",
		},
		{
			name:    "secret reference without a key",
			auth:    &Auth{BasicAuth: &BasicAuth{Username: secretKey("usage-db", "")}},
			wantErr: "basicAuth: username needs both a Secret name and a key",
		},
		{
			name:    "secret reference without a name",
			auth:    &Auth{Authorization: &Authorization{Credentials: secretKey("", "token")}},
			wantErr: "authorization: credentials needs both a Secret name and a key",
		},
		{
			name:    "authorization without credentials",
			auth:    &Auth{Authorization: &Authorization{Type: "Bearer"}},
			wantErr: "authorization: one of credentials and credentialsFile is required",
		},
		{
			name: "credentials from both a secret key and a file",
			auth: &Auth{Authorization: &Authorization{
				Credentials:     secretKey("usage-db", "token"),
				CredentialsFile: "/auth/credentials",
			}},
			wantErr: "authorization: credentials and credentialsFile are mutually exclusive",
		},
		{
			name:    "basic as an authorization type",
			auth:    &Auth{Authorization: &Authorization{Type: "basic", Credentials: secretKey("usage-db", "token")}},
			wantErr: `authorization: type "Basic" is not accepted, use basicAuth`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestAuthorization_GetType(t *testing.T) {
	var unset *Authorization
	assert.Equal(t, DefaultAuthorizationType, unset.GetType())
	assert.Equal(t, DefaultAuthorizationType, (&Authorization{}).GetType())
	assert.Equal(t, DefaultAuthorizationType, (&Authorization{Type: "  "}).GetType())
	assert.Equal(t, "Token", (&Authorization{Type: " Token "}).GetType())
}

func TestUsageDBConfig_DeepCopy_Auth(t *testing.T) {
	t.Run("no auth stays nil", func(t *testing.T) {
		copied := (&UsageDBConfig{ClientType: "prometheus"}).DeepCopy()
		assert.Nil(t, copied.Auth)
	})

	t.Run("basic auth is copied, not shared", func(t *testing.T) {
		original := &UsageDBConfig{
			ClientType: "prometheus",
			Auth: &Auth{BasicAuth: &BasicAuth{
				Username:     secretKey("usage-db", "username"),
				Password:     secretKey("usage-db", "password"),
				UsernameFile: "/auth/username",
				PasswordFile: "/auth/password",
			}},
		}
		copied := original.DeepCopy()
		require.Equal(t, original, copied)

		copied.Auth.BasicAuth.Username.Name = "other"
		copied.Auth.BasicAuth.Password.Key = "other"
		copied.Auth.BasicAuth.PasswordFile = "/other"
		assert.Equal(t, "usage-db", original.Auth.BasicAuth.Username.Name)
		assert.Equal(t, "password", original.Auth.BasicAuth.Password.Key)
		assert.Equal(t, "/auth/password", original.Auth.BasicAuth.PasswordFile)
	})

	t.Run("authorization is copied, not shared", func(t *testing.T) {
		original := &UsageDBConfig{
			ClientType: "prometheus",
			Auth: &Auth{Authorization: &Authorization{
				Type:        "Bearer",
				Credentials: secretKey("usage-db", "token"),
			}},
		}
		copied := original.DeepCopy()
		require.Equal(t, original, copied)

		copied.Auth.Authorization.Credentials.Key = "other"
		copied.Auth.Authorization.Type = "Token"
		assert.Equal(t, "token", original.Auth.Authorization.Credentials.Key)
		assert.Equal(t, "Bearer", original.Auth.Authorization.Type)
	})
}

// The scheduler configuration and the SchedulingShard are both read as JSON
// (sigs.k8s.io/yaml), so the JSON field names are the contract.
func TestUsageDBConfig_Auth_JSON(t *testing.T) {
	const document = `{
		"clientType": "prometheus",
		"connectionString": "https://metrics.example.com/prometheus",
		"auth": {
			"basicAuth": {
				"username": {"name": "usage-db", "key": "username"},
				"password": {"name": "usage-db", "key": "password"}
			}
		}
	}`

	var config UsageDBConfig
	require.NoError(t, json.Unmarshal([]byte(document), &config))
	require.NotNil(t, config.Auth)
	require.NotNil(t, config.Auth.BasicAuth)
	assert.Equal(t, secretKey("usage-db", "username"), config.Auth.BasicAuth.Username)
	assert.Equal(t, secretKey("usage-db", "password"), config.Auth.BasicAuth.Password)
	assert.NoError(t, config.Auth.Validate())

	encoded, err := json.Marshal(&UsageDBConfig{
		ClientType: "prometheus",
		Auth:       &Auth{Authorization: &Authorization{CredentialsFile: "/auth/credentials"}},
	})
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"clientType":"prometheus","auth":{"authorization":{"credentialsFile":"/auth/credentials"}}}`,
		string(encoded))
}
