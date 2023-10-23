package update

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/federation"
	log "github.com/sirupsen/logrus"
)

func UpdateMapping(client *gophercloud.ServiceClient, mappingID string, mapping_file string) {
	rules, err := ioutil.ReadFile(mapping_file)
	if err != nil {
		log.Fatalf("Couldn't read rules - %v", err)
	}
	var update federation.UpdateMappingOpts
	json.Unmarshal([]byte(rules), &update)

	res := federation.UpdateMapping(client, "mapping", update)
	log.Info(res)
}
