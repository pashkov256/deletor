package config

import "time"

type Config struct {
	Directory      string
	Extensions     []string
	MinSize        int64
	MaxSize        int64
	Exclude        []string
	IncludeSubdirs bool
	ShowProgress   bool
	IsCLIMode      bool
	HaveProgress   bool
	ConfirmDelete  bool
	OlderThan      time.Time
	NewerThan      time.Time
}

func LoadConfig() *Config {
	return GetFlags()
}

func (c *Config) GetConfig() *Config {
	return c
}
