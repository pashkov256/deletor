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
	Exclude    []string `json:",omitempty"`
	MinSize    string   `json:",omitempty"`
}

func UpdateRules(path, minSize string, extensions []string, exclude []string) {
	r := &Rules{Path: path, Extensions: extensions, Exclude: exclude, MinSize: minSize}
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
"exclude":[],
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

func (r *Rules) Equals(other *Rules) bool {
	if r.Path != other.Path || r.MinSize != other.MinSize {
		return false
	}

	if len(r.Extensions) != len(other.Extensions) {
		return false
	}

	if len(r.Exclude) != len(other.Exclude) {
		return false
	}

	for i := range r.Extensions {
		if r.Extensions[i] != other.Extensions[i] {
			return false
		}
	}

	for i := range r.Exclude {
		if r.Exclude[i] != other.Exclude[i] {
			return false
		}
	}

	return true
}
