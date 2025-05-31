package rules

type Rules interface {
	UpdateRules(options ...RuleOption) error
	GetRules() (*defaultRules, error)
	SetupRulesConfig() error
	GetRulesPath() string
	Equals(other Rules) bool
}

type defaultRules struct {
	Path                  string   `json:",omitempty"`
	Extensions            []string `json:",omitempty"`
	Exclude               []string `json:",omitempty"`
	MinSize               string   `json:",omitempty"`
	MaxSize               string   `json:",omitempty"`
	OlderThan             string   `json:",omitempty"`
	NewerThan             string   `json:",omitempty"`
	ShowHiddenFiles       bool     `json:",omitempty"`
	ConfirmDeletion       bool     `json:",omitempty"`
	IncludeSubfolders     bool     `json:",omitempty"`
	DeleteEmptySubfolders bool     `json:",omitempty"`
	SendFilesToTrash      bool     `json:",omitempty"`
	LogOperations         bool     `json:",omitempty"`
	LogToFile             bool     `json:",omitempty"`
	ShowStatistics        bool     `json:",omitempty"`
	ExitAfterDeletion     bool     `json:",omitempty"`
}

func NewRules() Rules {
	return &defaultRules{}
}
