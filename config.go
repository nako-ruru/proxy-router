package main

type Config struct {
	Server   Server   `yaml:"server"`
	Backends Backends `yaml:"backends"`
	Health   Health   `yaml:"health"`
	Selector string   `yaml:"selector"`
}
type Listen struct {
	Address  string         `yaml:"address"`
	CertFile string `yaml:"cert-file,omitempty"`
	KeyFile  string `yaml:"key-file,omitempty"`
	AutoCert string         `yaml:"auto-cert,omitempty"`
}
type Server struct {
	Listen  []Listen `yaml:"listen"`
	Verbose string      `yaml:"verbose"`
}
type Target struct {
	Server   string `yaml:"server"`
	Type     string `yaml:"type"`
	Username string	`yaml:"username"`
	Password string	`yaml:"password"`
}
type TargetsFromDocument struct {
	FilePath       string      `yaml:"file-path"`
	URL         string `yaml:"url"`
	ExtractType string `yaml:"extract-type"`
	Delimiter   string `yaml:"delimiter"`
	Type           string      `yaml:"type"`
	ScriptEntrance string      `yaml:"script-entrance"`
	ExtractScript  string      `yaml:"extract-script"`
}
type Backends struct {
	Targets             []Target              `yaml:"targets"`
	TargetsFromDocument []TargetsFromDocument `yaml:"targets-from-document"`
}
type Health struct {
	HealthInterval       string `yaml:"health-interval"`
	HealthTimeout        string `yaml:"health-timeout"`
	HealthURL            string `yaml:"health-url"`
	HealthResponseStatus string `yaml:"health-response-status"`
	Threads              int    `yaml:"threads"`
	UserAgent            string `yaml:"user-agent"`
}