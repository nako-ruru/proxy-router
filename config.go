package main

import "time"

type Config struct {
	Server   Server     `yaml:"server"`
	Backends []BackendC `yaml:"backends"`
	Health   HealthC     `yaml:"health"`
	Selector string     `yaml:"selector"`
}
type TLS struct {
	KeyStore         interface{} `yaml:"key-store"`
	KeyStorePassword interface{} `yaml:"key-store-password"`
	KeyStoreType     interface{} `yaml:"key-store-type"`
}
type Listen struct {
	Address string `yaml:"address"`
	TLS     TLS `yaml:"tls,omitempty"`
}
type Server struct {
	Listen  []Listen `yaml:"listen"`
	Verbose string      `yaml:"verbose"`
}
type BackendC struct {
	Server   string      `yaml:"server"`
	Type     string      `yaml:"type"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}
type HealthC struct {
	HealthInterval       time.Duration `yaml:"health-interval"`
	HealthTimeout        time.Duration `yaml:"health-timeout"`
	HealthURI            string `yaml:"health-uri"`
	HealthResponseStatus string `yaml:"health-response-status"`
	Threads              int    `yaml:"threads"`
	UserAgent            string `yaml:"user-agent"`
}