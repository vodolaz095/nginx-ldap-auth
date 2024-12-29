package config

type Permission struct {
	Host   string   `yaml:"host"`
	Prefix string   `yaml:"prefix"`
	UIDs   []string `yaml:"uids"`
	GIDs   []string `yaml:"gids"`
}
