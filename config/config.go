package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	FilePath string = "./config.yaml"
)

type Database struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
	Name    string `yaml:"name"`
	SslMode string `yaml:"ssl_mode"`
	Migrate bool   `yaml:"migrate"`
}

type Keyval struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Pw   string `yaml:"pw"`
	Db   int    `yaml:"db"`
}

type Observer struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
	UseSSL   bool   `yaml:"use_ssl"`
}

type Token struct {
	PublicKeyFile  string `yaml:"public_key_file"`
	PrivateKeyFile string `yaml:"private_key_file"`
	Project        string `yaml:"project"`
	SignMethod     string `yaml:"sign_method"`
}

type EventStream struct {
	StreamUrl string `yaml:"stream_url"`
}

type Turnstile struct {
	Secret string `yaml:"secret"`
	Skip   bool   `yaml:"skip"`
}

type Rest struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Domain       string `yaml:"domain"`
	AllowMethods string `yaml:"allowed_methods"`
	AllowHeaders string `yaml:"allowed_headers"`
	AllowOrigins string `yaml:"allowed_origins"`
	ExposeHeader string `yaml:"expose_headers"`
	AllowCred    bool   `yaml:"allow_credentials"`
}

type Rpc struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	Domain string `yaml:"domain"`
	UseSSL bool   `yaml:"use_ssl"`
}

type I18n struct {
	Locales []string `yaml:"locales"`
	Default string   `yaml:"default"`
}

type App struct {
	Database  Database    `yaml:"database"`
	Keyval    Keyval      `yaml:"keyval"`
	Observer  Observer    `yaml:"observer"`
	Token     Token       `yaml:"token"`
	Event     EventStream `yaml:"event"`
	Rest      Rest        `yaml:"rest"`
	Rpc       Rpc         `yaml:"rpc"`
	Turnstile Turnstile   `yaml:"turnstile"`
	I18n      I18n        `yaml:"i18n"`
}

func Bind(v interface{}) error {
	filename, err := filepath.Abs(FilePath)
	if err != nil {
		return err
	}
	cleanedDst := filepath.Clean(filename)
	yamlFile, err := os.ReadFile(cleanedDst)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlFile, v); err != nil {
		return err
	}
	return nil
}
