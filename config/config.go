package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
)

type (
	Config struct {
		// data represents the actual configuration information
		Data ConfigData

		configLocation string
	}

	ConfigData struct {
		SelectedMap      string   `json:"selectedMap"`
		ChatLogDirectory string   `json:"chatLogDirectory"`
		Channels         []string `json:"channels"`
		ClearWords       []string `json:"clearWords"`
	}
)

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) GetConfigDirectory() string {
	return filepath.Dir(cfg.configLocation)
}

func (cfg *Config) LoadConfig() error {

	loc, err := cfg.location()
	if err != nil {
		return err
	}

	// Check if the directory exists
	dir := filepath.Dir(loc)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Check if the config file exists
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		cfg.CreateDefaultConfig(loc)
	}

	// Load the config file!
	f, err := os.Open(loc)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	cd := ConfigData{}
	err = dec.Decode(&cd)
	if err != nil {
		return err
	}

	cfg.Data = cd
	cfg.configLocation = loc

	return nil
}

func (cfg *Config) SaveConfig() error {
	f, err := os.Create(cfg.configLocation)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(cfg.Data)

	return err
}

func (cfg *Config) CreateDefaultConfig(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Set the defaults for the config
	cd := ConfigData{
		SelectedMap:      "Providence",
		ChatLogDirectory: cfg.logDirHint(),
		Channels:         []string{"int.testing", "asdf"},
		ClearWords:       []string{"clear", "clr", "blue"},
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(cd)

	return err
}

func (cfg *Config) SetConfig(s map[string]interface{}) error {

	cd := ConfigData{}

	err := mapstructure.Decode(s, &cd)
	if err != nil {
		return err
	}

	cfg.Data = cd

	return cfg.SaveConfig()
}

func (cfg *Config) GetData() (string, error) {
	b, err := json.Marshal(cfg.Data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
