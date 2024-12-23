package config

import (
	"time"

	"gopkg.in/yaml.v3"

	"github.com/vodolaz095/nginx-ldap-auth/pkg/tracing"
	"github.com/vodolaz095/nginx-ldap-auth/pkg/zerologger"
)

type WebServer struct {
	Network                string        `yaml:"network" validate:"required"`
	Listen                 string        `yaml:"listen" validate:"required"`
	CookieName             string        `yaml:"cookie_name" validate:"required"`
	SessionSecret          string        `yaml:"session_secret" validate:"required"`
	SessionMaxAgeInSeconds time.Duration `yaml:"session_max_age"`
}

type Authenticator struct {
	// TTL depicts how long user profile is cached in session, when it expires, it is reloaded from ldap
	TTL time.Duration `yaml:"ttl"`
	//ConnectionString depicts how we dial LDAP server, something like ldap://127.0.0.1:389 or ldaps://ldap.example.org:636
	ConnectionString string `yaml:"connection_string" validate:"required"`
	// StartTLS shows, do we need to execute StartTLS or not
	StartTLS bool `yaml:"start_tls"`
	// InsecureTLS bypass ldap server tls cert verification
	InsecureTLS bool `yaml:"insecure_tls"`
	// ReadonlyDN is distinguished name used for authorization as readonly user,
	// who has access to listing groups of user. For example, "cn=readonly,dc=vodolaz095,dc=ru"
	ReadonlyDN string `yaml:"readonly_dn" validate:"required"`
	// ReadonlyPasswd is password for readonly user, who has access to listing groups
	ReadonlyPasswd string `yaml:"readonly_passwd" validate:"required"`
	// UserBaseTpl is template to extract user profiles by UID, for example
	// "uid=%s,ou=people,dc=vodolaz095,dc=ru" or
	// "email=%s,ou=people,dc=vodolaz095,dc=ru"
	UserBaseTpl string `yaml:"user_base_tpl" validate:"required"`
	// GroupsOU depicts organization unit for groups, usually "ou=groups,dc=vodolaz095,dc=ru"
	GroupsOU string `yaml:"groups_ou" validate:"required"`
}

type Cfg struct {
	WebServer     WebServer      `yaml:"webserver"`
	Authenticator Authenticator  `yaml:"authenticator"`
	Log           zerologger.Log `yaml:"log"`
	Tracing       tracing.Config `yaml:"tracing"`
	Realm         string         `yaml:"realm"`
}

func (c *Cfg) Dump() ([]byte, error) {
	return yaml.Marshal(c)
}
