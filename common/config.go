package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Name used in $HOME/.config/${Name}
const Name = "clash"

type Server struct {
	Host   string `toml:"host"`
	Port   string `toml:"port"`
	Secret string `toml:"secret"`
	HTTPS  bool   `toml:"https"`
}

func (s Server) URL() url.URL {
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", s.Host, s.Port),
	}

	if s.HTTPS {
		u.Scheme = "https"
	}

	return u
}

func (s Server) WebsocketURL() url.URL {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%s", s.Host, s.Port),
	}

	if s.HTTPS {
		u.Scheme = "wss"
	}

	return u
}

type Config struct {
	Servers  map[string]Server `toml:"servers"`
	Selected string            `toml:"selected"`
}

// Init create config if not exist
func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	homeDir = filepath.Join(homeDir, ".config", Name)
	// initial homedir
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		if err := os.MkdirAll(homeDir, 0o755); err != nil {
			return fmt.Errorf("can't create config directory %s: %s", homeDir, err.Error())
		}
	}

	cfgFile := filepath.Join(homeDir, "ctl.toml")
	// initial config.yaml
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		f, err := os.OpenFile(cfgFile, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("can't create file %s: %s", cfgFile, err.Error())
		}
		f.Close()
	}

	return nil
}

func GetCfgPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cfgFile := filepath.Join(homeDir, ".config", Name, "ctl.toml")
	return cfgFile, nil
}

func ReadCfg() (*Config, error) {
	cfgFile, err := GetCfgPath()
	if err != nil {
		return nil, err
	}

	buf, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Servers: make(map[string]Server),
	}
	if err := toml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func SaveCfg(cfg *Config) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	cfgPath, err := GetCfgPath()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(cfgPath, buf, 0o666)
}

func GetCurrentServer(cfg *Config) (string, *Server, error) {
	current := cfg.Selected
	if current == "" {
		return "", nil, errors.New("not select any server")
	}

	server, ok := cfg.Servers[current]
	if !ok {
		return "", nil, fmt.Errorf("selected %s but no in server list", current)
	}

	return current, &server, nil
}
