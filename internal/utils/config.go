package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backup struct {
		Interval     string `yaml:"interval"`
		SavePath     string `yaml:"save_path"`
		ClientId     string `yaml:"clientId"`
		ClientSecret string `yaml:"clientSecret"`
		Host         string `yaml:"host"`
		AuthKey      string `yaml:"authKey"`
	} `yaml:"backup"`
	IMAP struct {
		Server   string `yaml:"server"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"imap"`
	Mail struct {
		Emails  []string `yaml:"emails"`
		Exclude []string `yaml:"exclude"`
	} `yaml:"mail"`
}

func ShouldProcessEmail(cfg *Config, email string) bool {
	// Если в emails указан "all", значит фильтруем только исключения
	if len(cfg.Mail.Emails) == 1 && cfg.Mail.Emails[0] == "all" {
		for _, excl := range cfg.Mail.Exclude {
			if excl == email {
				return false // В списке исключений — пропускаем
			}
		}
		return true // Загружаем все остальные
	}

	// Если указан список адресов, то загружаем только их
	for _, allowed := range cfg.Mail.Emails {
		if allowed == email {
			return true
		}
	}

	return false // Не в списке — пропускаем
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
