package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"slices"
	"strings"
	"sync"
	"synok/common"

	log "github.com/sirupsen/logrus"
)

type ProjectMembers struct {
	Data struct {
		Leaders        []Member `json:"leaders"`
		ActiveMembers  []Member `json:"activeMembers"`
		PassiveMembers []Member `json:"passiveMembers"`
	} `json:"data"`
}

type Member struct {
	Email []string `json:"email"`
}

type ProjectHeader struct {
	Data struct {
		Number      int    `json:"number,string"`
		NameRus     string `json:"nameRus"`
		StatusLabel string `json:"statusLabel"`
	} `json:"data"`
}

func GetCabinetState() map[string][]int {
	log.Info("Syncing projects...")
	var wg sync.WaitGroup
	emailToProjects := make(map[string][]int)
	mu := &sync.Mutex{}
	guard := make(chan struct{}, common.MaxGoroutines)

	for i := 0; i <= common.MaxProjectNum; i++ {
		wg.Add(1)
		guard <- struct{}{}
		go func(projectId int) {
			defer func() {
				wg.Done()
				<-guard
			}()

			if projectId%(common.MaxProjectNum/10) == 0 {
				log.Infof("Finished parsing about %d%% of projects", projectId/20)
			}

			respHead, err := http.Get(common.ProjectHeaderUrl + fmt.Sprint(projectId))
			if err != nil {
				log.Warnf("Error fetching project: %d - %v", projectId, err)
				return
			}
			defer respHead.Body.Close()

			data, _ := ioutil.ReadAll(respHead.Body)
			var projectHead ProjectHeader
			err = json.Unmarshal(data, &projectHead)
			if err != nil {
				log.Warnf("Error parsing JSON for project: %d - %v", projectId, err)
				return
			}
			if projectHead.Data.StatusLabel != "Готов к работе" && projectHead.Data.StatusLabel != "Рабочий" {
				log.Debugf("Project %d isn't labeled active: statusLabel is %s", projectId, projectHead.Data.StatusLabel)
				return
			}
			var projectNum int 
      projectNum = projectHead.Data.Number

			respMemb, err := http.Get(common.ProjectMembersURL + fmt.Sprint(projectId))
			if err != nil {
				log.Warnf("Error fetching project members: %d - %v", projectId, err)
				return
			}
			defer respMemb.Body.Close()

			data, _ = ioutil.ReadAll(respMemb.Body)

			var projectData ProjectMembers
			err = json.Unmarshal(data, &projectData)
			if err != nil {
				log.Warnf("Error parsing JSON for project members: %d - %v", projectId, err)
				return
			}
			mu.Lock()
			for _, member := range append(projectData.Data.Leaders, projectData.Data.ActiveMembers...) {
				emailSuffix := common.EmailPostfix
				for _, _email := range member.Email {
					if strings.HasSuffix(_email, common.WorkerEmailPostfix) {
						emailSuffix = common.WorkerEmailPostfix
					}
				}
				email := common.TrimEmail(member.Email[0]) + emailSuffix
				if !slices.Contains(emailToProjects[email], projectNum) {
					emailToProjects[email] = append(emailToProjects[email], projectNum)
				}
			}
			mu.Unlock()

			log.Debugf("Finished %d out of %d\r", projectId, common.MaxProjectNum)
		}(i)
	}

	wg.Wait()

	return emailToProjects
}
