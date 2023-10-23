package render

import (
	"os"
	"synok/common"

	"github.com/kluctl/go-jinja2"
	log "github.com/sirupsen/logrus"
)

func RenderRules(userProjects map[string][]int) {
	log.Infof("Rendering rules")
	j2, err := jinja2.NewJinja2("rules", common.MaxGoroutines, jinja2.WithGlobal("userProjects", userProjects))
	if err != nil {
		log.Fatalf("Couldn't render rules - %v", err)
	}
	rules, err := j2.RenderFile(os.Getenv("PWD") + "/render/rules.json.j2")
	if err != nil {
		log.Fatalf("Couldn't render rules - %v", err)
	}

	os.WriteFile(common.RulesLocaton, []byte(rules), 0644)
}
