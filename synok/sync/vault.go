package sync

import (
	"fmt"
	"strings"
	"sync"
	"synok/common"

	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func toVaultEntityObject(userName string, userProjects []int) map[string]interface{} {
	projString := "projects=" + strings.Trim(strings.Join(strings.Fields(fmt.Sprint(userProjects)), ";"), "[]")
	resultingObject := map[string]interface{}{
		"name":     common.TrimEmail(userName),
		"metadata": projString,
	}

	return resultingObject

}
func toVaultAliasObject(userName string, canonicalID string, mountAccessor string) map[string]interface{} {
	resulting_object := map[string]interface{}{
		"name":           userName,
		"canonical_id":   canonicalID,
		"mount_accessor": mountAccessor,
	}
	return resulting_object
}

func commitEntity(c *vault.Client, vaultMountAccessor string, userName string, userProjects []int) {
	log.Debugf("Writing %s", userName)
	entity := toVaultEntityObject(userName, userProjects)
	secret, err := c.Logical().Write(common.EntityPath, entity)

	if err != nil {
		log.Warnf("Unable to write entity to Vault: %s - %v", userName, err)
		return
	}
	if secret == nil {
		log.Warnf("Update reply empty for %s", userName)
		return
	}

	var entityID interface{}
	if _, present := secret.Data["id"]; !present {
		return
	}
	entityID = secret.Data["id"]
	entity_alias := toVaultAliasObject(userName, entityID.(string), vaultMountAccessor)
	_, err = c.Logical().Write(common.EntityAliasPath, entity_alias)
	if err != nil {
		log.Warnf("Unable to write entity alias to Vault: %v", err)
		return
	}

}

func deleteEntityByName(c *vault.Client, userName string) {
	log.Debugf("Deleting %s\n", userName)
	_, err := c.Logical().Delete(common.EntityNamePath + common.TrimEmail(userName))
	if err != nil {
		log.Warnf("Unable to delete %s from vault: %v", userName, err)
	}
}

func CommitUserDeltaToVault(vaultClient *vault.Client, vaultMountAccessor string, delta map[string][]int) error {
	log.Infof("Commiting changes")

	var wg sync.WaitGroup
	guard := make(chan struct{}, common.MaxGoroutines)

	for name, projects := range delta {
		wg.Add(1)
		guard <- struct{}{}
		go func(userName string, userProjects []int) {
			defer func() {
				wg.Done()
				<-guard
			}()

			if len(userProjects) > 0 {
				commitEntity(vaultClient, vaultMountAccessor, userName, userProjects)
			} else {
				deleteEntityByName(vaultClient, userName)
			}

		}(name, projects)
	}

	wg.Wait()

	return nil
}
