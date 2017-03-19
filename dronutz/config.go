package dronutz

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Host              string `yaml:"host"`
	ApiPort           int    `yaml:"api_port"`
	KitchenPort       int    `yaml:"kitchen_port"`
	PublicDir         string `yaml:"public_directory"`
	Tracer            string `yaml:"tracer"`
	TracerHost        string `yaml:"tracer_host"`
	TracerPort        int    `yaml:"tracer_port"`
	TracerAccessToken string `yaml:"tracer_access_token"`
}

func (c *Config) Validate() error {
	switch {
	case c.Host == "":
		return errors.New("host is not set")
	case c.PublicDir == "":
		return errors.New("public_directory is not set")
	case c.ApiPort == 0:
		return errors.New("api_port is not set")
	case c.KitchenPort == 0:
		return errors.New("kitchen_port is not set")
	default:
		return nil
	}
}

func (c *Config) APIAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.ApiPort)
}

func (c *Config) KitchenAddress() string {
	return fmt.Sprintf(":%d", c.KitchenPort)
}

func NewConfigFromPath(path string) (Config, error) {
	cfg := Config{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	err = cfg.Validate()
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

var configJSTemplate = template.Must(template.New("config").Parse(`
var Config = {
	tracer: "{{.Tracer}}",
	tracer_host: "{{.TracerHost}}",
	tracer_port: {{.TracerPort}},
	tracer_access_token: "{{.TracerAccessToken}}",
}
`))
