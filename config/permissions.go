package config

type Permission struct {
	Host   string   `yaml:"host" validate:"required"`
	Prefix string   `yaml:"prefix" validate:"required"`
	UIDs   []string `yaml:"uids"`
	GIDs   []string `yaml:"gids"`
}
