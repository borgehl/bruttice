package configx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigRaw struct {
	AdminAccounts []string `yaml:"admin-accounts"`
	Timezone      string   `yaml:"timezone"`
	CookieSalt    string   `yaml:"cookie-salt"`
	IngestKey     string   `yaml:"ingest-key"`
}

type AdminAccount struct {
	Email    string
	Password string
}

type Config struct {
	AdminAccounts []AdminAccount
	TZLocation    *time.Location
	CookieSalt    string
	IngestKey     string
}

func ParseConfig(configFilename *string) Config {
	filename := "/app/config.yaml"
	if configFilename != nil {
		filename = *configFilename
	}
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	raw := ConfigRaw{}
	if err = yaml.Unmarshal(fileBytes, &raw); err != nil {
		panic(err)
	}

	ret := Config{
		AdminAccounts: make([]AdminAccount, 0, len(raw.AdminAccounts)),
		CookieSalt:    raw.CookieSalt,
		IngestKey:     raw.IngestKey,
	}

	for _, account := range raw.AdminAccounts {
		split := strings.Split(account, ";")
		if len(split) != 2 {
			panic(fmt.Errorf("malformed admin account: %s", account))
		}
		ret.AdminAccounts = append(ret.AdminAccounts, AdminAccount{
			Email:    split[0],
			Password: split[1],
		})
	}

	loc, err := time.LoadLocation(raw.Timezone)
	if err != nil {
		panic(err)
	}
	ret.TZLocation = loc

	return ret
}
