package common

const MaxProjectNum = 2000
const MaxGoroutines = 15
const EntityPath = "identity/entity/"
const EntityNamePath = "identity/entity/name/"
const EntityAliasPath = "identity/entity-alias/"
const EmailPostfix = "@edu.hse.ru"
const WorkerEmailPostfix = "@hse.ru"
const ProjectMembersURL = "https://cabinet.miem.hse.ru/public-api/project/students/"
const ProjectHeaderUrl = "https://cabinet.miem.hse.ru/public-api/project/header/"
const RulesTemplateLocation = "render/rules.json.j2"
const RulesLocaton = "rules.json"

var OpenstackRequieredEnv = []string{"OS_USERNAME", "OS_PASSWORD", "OS_AUTH_URL", "OS_PROJECT_NAME", "OS_DOMAIN_NAME", "OS_REGION_NAME"}
