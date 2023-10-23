package diff

import (
	"slices"
	"sync"
	"synok/common"

	log "github.com/sirupsen/logrus"
)

func sliceValuesEqual(x []int, y []int) bool {
	if len(x) != len(y) {
		return false
	}

	for _, v := range x {
		if !slices.Contains(y, v) {
			return false
		}
	}

	return true

}

func ProjectDelta(cabinetData map[string][]int, vaultData map[string][]int) map[string][]int {
	log.Info("Building Delta...")
	wg := &sync.WaitGroup{}
	delta := make(chan common.UserProjects)

	for Email := range cabinetData {
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			var userDiff common.UserProjects

			if vaultValue, presentInVault := vaultData[email]; presentInVault {
				log.Debugf("Analyzing: %s vs %s\n", cabinetData[email], vaultValue)
				if sliceValuesEqual(cabinetData[email], vaultValue) {
					return
				}
			} else {
				userDiff.Name = email
				userDiff.Projects = cabinetData[email]
				log.Debugf("Verdict: %s:%s to be committed", userDiff.Name, userDiff.Projects)
			}
			delta <- userDiff
		}(Email)
	}

	for Email := range vaultData {
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			if _, presentInCabinet := cabinetData[email]; !presentInCabinet {
				var userDiff common.UserProjects
				userDiff.Name = email
				userDiff.Projects = []int{}
				delta <- userDiff
				log.Debugf("Verdict: %s to be deleted", userDiff.Name)
			}
		}(Email)
	}

	go common.WaitGoroutines(wg, delta)

	deltaMap := make(map[string][]int)
	for userdelta := range delta {
		deltaMap[userdelta.Name] = userdelta.Projects
	}

	return deltaMap
}
