package common

type Config struct {
	Http      Http       `yaml:"Http"`
	Endpoints []Endpoint `yaml:"Endpoints"`
	Log       Log        `yaml:"Log"`
	Forbidden Forbidden  `yaml:"Forbidden"`
}

type Forbidden struct {
	ForbiddenAccountNotFound    bool `yaml:"ForbiddenAccountNotFound"`
	ForbiddenProxyCredentialErr bool `yaml:"ForbiddenProxyCredentialErr"`
}

type Audit struct {
	Output  string `yaml:"Output"`
	Enabled bool   `yaml:"Enabled"`
	MaxAge  int    `yaml:"MaxAge"`
	MaxSize int    `yaml:"MaxSize"`
}

type Log struct {
	Output  string `yaml:"Output"`
	Level   string `yaml:"Level"`
	MaxAge  int    `yaml:"MaxAge"`
	MaxSize int    `yaml:"MaxSize"`
}

type Http struct {
	Address string `yaml:"Address"`
	Tls     Tls    `yaml:"Tls"`
}

type Tls struct {
	Address  string `yaml:"Address"`
	Enabled  bool   `yaml:"Enabled"`
	CertFile string `yaml:"CertFile"`
	KeyFile  string `yaml:"KeyFile"`
}

type Endpoint struct {
	CloudAccountName string      `yaml:"CloudAccountName"`
	Vendor           string      `yaml:"Vendor"`
	Credentials      Credentials `yaml:"Credentials"`
}

type Credentials struct {
	Proxy Credential `yaml:"Proxy"`
	Real  Credential `yaml:"Real"`
}

type Credential struct {
	AccessKey    string `yaml:"AccessKey"`
	SecretKey    string `yaml:"SecretKey"`
	AccessToken  string `yaml:"AccessToken"`
	ClientToken  string `yaml:"ClientToken"`
	ClientSecret string `yaml:"ClientSecret"`
}
