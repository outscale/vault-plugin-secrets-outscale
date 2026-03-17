package outscalevault_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	stepwise "github.com/hashicorp/vault-testing-stepwise"
	dockerEnvironment "github.com/hashicorp/vault-testing-stepwise/environments/docker"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

func TestAcc(t *testing.T) {
	t.Parallel()

	envOptions := &stepwise.MountOptions{
		RegistryName:    "outscale",
		PluginType:      api.PluginTypeSecrets,
		PluginName:      "vault-plugin-secrets-outscale",
		MountPathPrefix: "outscale",
	}

	roleName := "vault-stepwise-user-role"
	secretKey := new(string)
	accessKey := new(string)

	stepwise.Run(t, stepwise.Case{
		Precheck:    func() { testAccPreCheck(t) },
		Environment: dockerEnvironment.NewEnvironment("outscale", envOptions),
		Steps: []stepwise.Step{
			testAccConfig(t),
			testAccUserRole(t, roleName),
			testAccUserRoleRead(t, roleName),
			testAccUserCredRead(t, roleName, accessKey, secretKey),
		},
	})
}

var initSetup sync.Once

func testAccPreCheck(t *testing.T) {
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

func testAccConfig(t *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "config",
		Data: map[string]any{
			"access_key": os.Getenv("OSC_ACCESS_KEY"),
			"secret_key": os.Getenv("OSC_SECRET_KEY"),
			"region":     os.Getenv("OSC_REGION"),
		},
	}
}

func testAccUserRole(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "roles/" + roleName,
		Data: map[string]any{
			"ttl":     "1m",
			"max_ttl": "5m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, resp)
			require.Nil(t, err)
			return nil
		},
	}
}

func testAccUserRoleRead(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "roles/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
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
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["access_key"])
			require.NotEmpty(t, resp.Data["secret_key"])
			*accessKey = resp.Data["access_key"].(string)
			*secretKey = resp.Data["secret_key"].(string)
			return nil
		},
	}
}
