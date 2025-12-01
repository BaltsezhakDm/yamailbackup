// Proposed updates for yamailbackup: support custom mailbox folder and subject-based filters.

// ================= utils/config.go =================
package utils

import (
	"os"
	"strings"

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
		// Новое поле: имя папки (mailbox), по умолчанию INBOX
		Mailbox string `yaml:"mailbox"`
	} `yaml:"imap"`
	Mail MailConfig `yaml:"mail"`
}

type MailConfig struct {
	Emails  []string `yaml:"emails"`
	Exclude []string `yaml:"exclude"`
	// Новые поля: фильтрация по теме письма (подстрокой, регистронезависимо)
	SubjectInclude []string `yaml:"subject_include"`
	SubjectExclude []string `yaml:"subject_exclude"`
}

// ShouldProcessEmail решает, нужно ли обрабатывать письмо на основе
// адреса отправителя и темы письма.
func ShouldProcessEmail(cfg *Config, fromEmail, subject string) bool {
	if !shouldProcessByEmail(&cfg.Mail, fromEmail) {
		return false
	}
	return shouldProcessBySubject(&cfg.Mail, subject)
}

func shouldProcessByEmail(mail *MailConfig, email string) bool {
	e := strings.TrimSpace(strings.ToLower(email))

	// Если в emails указан "all", значит фильтруем только исключения
	if len(mail.Emails) == 1 && strings.ToLower(strings.TrimSpace(mail.Emails[0])) == "all" {
		for _, excl := range mail.Exclude {
			if strings.TrimSpace(strings.ToLower(excl)) == e {
				return false // В списке исключений — пропускаем
			}
		}
		return true // Загружаем все остальные
	}

	// Если указан список адресов, то загружаем только их
	for _, allowed := range mail.Emails {
		if strings.TrimSpace(strings.ToLower(allowed)) == e {
			return true
		}
	}

	return false // Не в списке — пропускаем
}

func shouldProcessBySubject(mail *MailConfig, subject string) bool {
	s := strings.ToLower(subject)

	// Сначала проверяем исключения по теме
	for _, excl := range mail.SubjectExclude {
		excl = strings.ToLower(strings.TrimSpace(excl))
		if excl != "" && strings.Contains(s, excl) {
			return false
		}
	}

	// Если список включений пустой — пропускаем всё, что не попало в exclude
	if len(mail.SubjectInclude) == 0 {
		return true
	}

	// Если есть include — тема должна содержать хотя бы одну подстроку
	for _, incl := range mail.SubjectInclude {
		incl = strings.ToLower(strings.TrimSpace(incl))
		if incl != "" && strings.Contains(s, incl) {
			return true
		}
	}

	return false
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
