package main

import (
	"os"

	"synok/common"
	"synok/diff"
	"synok/parse"
	"synok/render"
	"synok/sync"
	"synok/update"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"

	vault "github.com/hashicorp/vault/api"

	log "github.com/sirupsen/logrus"
)

func getRequiredEnv(env_var string) (env_value string) {
	auth_info, info_passed := os.LookupEnv(env_var)
	if !info_passed {
		log.Fatalf("%s unset - terminating", env_var)
	}

	return auth_info
}

func checkOpenstackEnvVar(env_var string) {
	_, info_passed := os.LookupEnv(env_var)
	if !info_passed {
		log.Fatalf("%s unset - terminating", env_var)
	}
}

func getOpenstackIdentityV3Client() *gophercloud.ServiceClient {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		log.Fatalf("%v", err)
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalf("%v", err)
	}

	client, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Fatalf("%v", err)
	}

	return client
}

func getVaultClient(vault_addr string, vault_token string) *vault.Client {
	config := vault.DefaultConfig()
	config.Address = vault_addr

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("Unable to initialize Vault client: %v", err)
	}

	client.SetToken(vault_token)

	return client
}

func setLogLevel() {

	switch level, _ := os.LookupEnv("SYNC_LOG_LEVEL"); level {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

}

func main() {
	setLogLevel()

	vaultAddr := getRequiredEnv("VAULT_ADDR")
	vaultToken := getRequiredEnv("VAULT_TOKEN")

	vaultMountAccessor := getRequiredEnv("VAULT_ACCESSOR")
	mappingID := getRequiredEnv("MAPPING_ID")

	vaultClient := getVaultClient(vaultAddr, vaultToken)

	for _, env_var := range common.OpenstackRequieredEnv {
		checkOpenstackEnvVar(env_var)
	}
	openstackClient := getOpenstackIdentityV3Client()

	cabinetData := parse.GetCabinetState()
  log.Debugf("%v", cabinetData)
	//cabinetData := make(map[string][]int) // <- this can be used to force-clear vault: just uncomment it and comment out previous line

	vaultData := parse.GetVaultState(vaultClient, vaultMountAccessor)
  log.Debugf("%v", vaultData)

	diff := diff.ProjectDelta(cabinetData, vaultData)
  log.Debugf("%v", diff)

	sync.CommitUserDeltaToVault(vaultClient, vaultMountAccessor, diff)

	newVaultData := parse.GetVaultState(vaultClient, vaultMountAccessor)

	render.RenderRules(newVaultData)

	update.UpdateMapping(openstackClient, mappingID, common.RulesLocaton)

	log.Info("Finished successfully")
}
