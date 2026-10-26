package outscalevault_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	stepwise "github.com/hashicorp/vault-testing-stepwise"
	dockerEnvironment "github.com/hashicorp/vault-testing-stepwise/environments/docker"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/helper/template"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
	"github.com/outscale/osc-sdk-go/v3/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestStepwise(t *testing.T) {
	t.Parallel()
	testAccPreCheck(t)

	p := profile.Profile{
		AccessKey: os.Getenv("OSC_ACCESS_KEY"),
		SecretKey: os.Getenv("OSC_SECRET_KEY"),
		Region:    os.Getenv("OSC_REGION"),
		Protocol:  "https",
	}

	envOptions := &stepwise.MountOptions{
		RegistryName:    "outscale",
		PluginType:      api.PluginTypeSecrets,
		PluginName:      "vault-plugin-secrets-outscale",
		MountPathPrefix: "outscale",
	}

	roleName := "vault-stepwise-user-role"
	rootRoleName := "vault-stepwise-root-role"
	secretKey := new(string)
	accessKey := new(string)

	stepwise.Run(t, stepwise.Case{
		Environment: dockerEnvironment.NewEnvironment("outscale", envOptions),
		Steps: []stepwise.Step{
			testAccConfig(t, p),
			testAccUserRole(t, roleName, p),
			testAccUserRoleRead(t, roleName),
			testAccUserCredRead(t, roleName, accessKey, secretKey),
			testAccRootRole(t, rootRoleName),
			testAccRootRoleRead(t, rootRoleName),
			testAccRootCredRead(t, rootRoleName),
		},
	})
}

var initSetup sync.Once

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("VAULT_ACC") != "1" {
		t.Skip("Acceptance tests skipped unless VAULT_ACC=1 is set")
	}

	initSetup.Do(func() {
		if v := os.Getenv("OSC_SECRET_KEY"); v == "" {
			t.Skip(fmt.Printf("%s not set", "OSC_SECRET_KEY"))
		}
		if v := os.Getenv("OSC_ACCESS_KEY"); v == "" {
			t.Skip(fmt.Printf("%s not set", "OSC_ACCESS_KEY"))
		}
		if v := os.Getenv("OSC_REGION"); v == "" {
			t.Skip(fmt.Printf("%s not set", "OSC_REGION"))
		}
	})
}

func testAccConfig(t *testing.T, p profile.Profile) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "config",
		Data: map[string]any{
			"access_key": p.AccessKey,
			"secret_key": p.SecretKey,
			"region":     p.Region,
			"protocol":   p.Protocol,
		},
	}
}

func testAccUserRole(t *testing.T, roleName string, p profile.Profile) stepwise.Step {
	t.Helper()

	client, err := osc.NewClient(&p)
	require.NoError(t, err)

	groupTemplate, err := template.NewTemplate(template.Template("vault-test-{{random 20}}"))
	require.NoError(t, err)
	groupName, err := groupTemplate.Generate(nil)
	require.NoError(t, err)

	_, err = client.CreateUserGroup(t.Context(), osc.CreateUserGroupRequest{UserGroupName: groupName})
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		_, err := client.DeleteUserGroup(ctx, osc.DeleteUserGroupRequest{UserGroupName: groupName})
		if err != nil {
			t.Errorf("failed to delete test user group: %v", err)
		}
	})

	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "roles/" + roleName,
		Data: map[string]any{
			"user_group_names": []string{groupName},
			"ttl":              "1m",
			"max_ttl":          "5m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, resp)
			require.NoError(t, err)
			return nil
		},
	}
}

func testAccUserRoleRead(t *testing.T, roleName string) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "roles/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NoError(t, err)
			require.NotNil(t, resp)
			return nil
		},
	}
}

func testAccUserCredRead(
	t *testing.T,
	roleName string,
	accessKey, secretKey *string,
) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["access_key"])
			require.NotEmpty(t, resp.Data["secret_key"])

			*accessKey = resp.Data["access_key"].(string)
			*secretKey = resp.Data["secret_key"].(string)

			return nil
		},
	}
}

func testAccRootRole(t *testing.T, roleName string) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "root-role/" + roleName,
		Data: map[string]any{
			"ttl": "1m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.NoError(t, err)
			require.Nil(t, resp)
			return nil
		},
	}
}

func testAccRootRoleRead(t *testing.T, roleName string) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "root-role/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NoError(t, err)
			require.NotNil(t, resp)
			return nil
		},
	}
}

func testAccRootCredRead(t *testing.T, roleName string) stepwise.Step {
	t.Helper()

	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "root-creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["access_key"])
			require.NotEmpty(t, resp.Data["secret_key"])

			return nil
		},
	}
}
