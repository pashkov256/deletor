package rules

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

var (
	AppDirName   = "deletor"
	RuleFileName = "rule.json"
)

type Rules struct {
	Path       string   `json:",omitempty"`
	Extensions []string `json:",omitempty"`
	Excludes   []string `json:",omitempty"`
	MinSize    string   `json:",omitempty"`
}

func UpdateRules(path, minSize string, extensions []string) {
	r := &Rules{Path: path, Extensions: extensions, MinSize: minSize}
	rulesJSON, _ := json.Marshal(r)
	os.WriteFile(GetRulesPath(), rulesJSON, 0644)
}

func GetRules() *Rules {
	jsonRules, _ := os.ReadFile(GetRulesPath())
	r := &Rules{}
	json.Unmarshal(jsonRules, r)
	return r
}

func SetupRulesConfig() {
	filePathRuleConfig := GetRulesPath()
	os.MkdirAll(filepath.Dir(filePathRuleConfig), 0755)

	_, err := os.Stat(filePathRuleConfig)

	if err != nil {
		baseRules := `
{
"path":"",
"extensions":[],
"min_size":null
}`

		err := os.WriteFile(filePathRuleConfig, []byte(baseRules), 0644)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetRulesPath() string {
	userConfigDir, _ := os.UserConfigDir()
	filePathRuleConfig := filepath.Join(userConfigDir, AppDirName, RuleFileName)

	return filePathRuleConfig
}
