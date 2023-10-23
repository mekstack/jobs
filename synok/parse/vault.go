package parse

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"synok/common"

	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func vaultFormatToSlice(projects string) []int {
	stringSlice := strings.Split(projects, ";")
	var intSlice []int

	for _, v := range stringSlice {
		intv, _ := strconv.Atoi(v)
		intSlice = append(intSlice, intv)
	}
	return intSlice
}

func getEntityAliasName(c *vault.Client, entity_secret *vault.Secret, vault_mount_accessor string, entity_name string) (string, error) {
	entity_alias_info := entity_secret.Data["aliases"]
	if entity_alias_info == nil {
		return "", fmt.Errorf("No aliases found for %s, ignoring", entity_name)

	}

	entity_alias_info_provider := entity_alias_info.([]interface{})
	if len(entity_alias_info_provider) == 0 {
		return "", fmt.Errorf("No alias providers found for %s, ignoring", entity_name)
	}

	// TODO: when there is more than one identity provider, verify it fits mount accessor
	entity_alias_name := entity_alias_info_provider[0].(map[string]interface{})["name"].(string)
	log.Debugf("Successfully parsed email from vault: ", entity_alias_name)
	return entity_alias_name, nil
}

func readEntity(c *vault.Client, entityName string, vault_mount_accessor string) (common.UserProjects, error) {
	entitySecret, err := c.Logical().Read(common.EntityNamePath + entityName)
	if err != nil {
		return common.UserProjects{}, err
	}

	if entitySecret == nil {
		return common.UserProjects{}, fmt.Errorf("Cannot read info about %s, returning nil", entityName)
	}

	entityMetadata := entitySecret.Data["metadata"]

	if entityMetadata == nil {
		return common.UserProjects{}, fmt.Errorf("No metadata for %s, ignoring", entityName)
	}

	entityName, err = getEntityAliasName(c, entitySecret, vault_mount_accessor, entityName)

	if err != nil {
		return common.UserProjects{}, fmt.Errorf("No mount accessor for %s, ignoring", entityName)
	}

	entityProjects := entityMetadata.(map[string]interface{})["projects"].(string)

	projects := vaultFormatToSlice(entityProjects)

	var res common.UserProjects
	res.Name = entityName
	res.Projects = projects
	return res, nil

}

func GetVaultState(vaultClient *vault.Client, vaultMountAccessor string) map[string][]int {
	log.Infof("Syncing vault...")
	entitiesSecret, err := vaultClient.Logical().List(common.EntityNamePath)
	if err != nil {
		log.Fatalf("Cannot list entities - %v", err)
	}
	if entitiesSecret == nil {
		return make(map[string][]int)
	}

	var entities []interface{}
	if entities, present := entitiesSecret.Data["keys"]; !present || entities == nil {
		return make(map[string][]int)
	}
	entities = entitiesSecret.Data["keys"].([]interface{})

	wg := &sync.WaitGroup{}
	users := make(chan common.UserProjects)
	//guard := make(chan struct{}, common.MaxGoroutines)

	for _, entity := range entities {
		wg.Add(1)
		//guard <- struct{}{}
		go func(identityEntity string) {
			defer func() {
				wg.Done()
				//<-guard
			}()
			userProjects, err := readEntity(vaultClient, identityEntity, vaultMountAccessor)
			log.Debugf("Read %s successfully", identityEntity)
			if err != nil {
				log.Debugf("%v", err)
				return
			}
			users <- userProjects
		}(entity.(string))
	}

	go common.WaitGoroutines(wg, users)

	userProject := make(map[string][]int)
	for user := range users {
		userProject[user.Name] = user.Projects
	}

	return userProject
}
