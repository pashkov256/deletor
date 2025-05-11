package rules

type Rules interface {
	UpdateRules(path, minSize string, extensions, exclude []string) error
	GetRules() (*defaultRules, error)
	SetupRulesConfig() error
	GetRulesPath() string
	Equals(other Rules) bool
}

type defaultRules struct {
	Path       string   `json:",omitempty"`
	Extensions []string `json:",omitempty"`
	Exclude    []string `json:",omitempty"`
	MinSize    string   `json:",omitempty"`
}

func NewRules() Rules {
	return &defaultRules{}
}
