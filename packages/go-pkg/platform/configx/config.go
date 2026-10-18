package configx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigRawVenue struct {
	Name   string `yaml:"name"`
	Link   string `yaml:"link"`
	Region string `yaml:"region"`
}

type ConfigRawAgeGroup struct {
	Name  string `yaml:"name"`
	Value uint8  `yaml:"value"`
}

type ConfigRawRegion struct {
	Name     string `yaml:"name"`
	Timezone string `yaml:"timezone"`
}

type ConfigRaw struct {
	AdminAccounts    []string            `yaml:"admin-accounts"`
	Timezone         string              `yaml:"timezone"`
	CookieSalt       string              `yaml:"cookie-salt"`
	IngestKey        string              `yaml:"ingest-key"`
	IngestTriggerURL string              `yaml:"ingest-trigger-url"`
	Tags             []string            `yaml:"tags"`
	AgeGroups        []ConfigRawAgeGroup `yaml:"age-groups"`
	Regions          []ConfigRawRegion   `yaml:"regions"`
	Venues           []ConfigRawVenue    `yaml:"venues"`
}

type AdminAccount struct {
	Email    string
	Password string
}

type InitialData struct {
	Tags      []string
	AgeGroups []ConfigRawAgeGroup
	Regions   []ConfigRawRegion
	Venues    []ConfigRawVenue
}

type Config struct {
	AdminAccounts []AdminAccount
	TZLocation    *time.Location
	CookieSalt    string
	IngestKey     string
	// IngestTriggerURL is the scraper service endpoint that starts a run, empty if there is none
	IngestTriggerURL string
	InitialData      InitialData
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
		AdminAccounts:    make([]AdminAccount, 0, len(raw.AdminAccounts)),
		CookieSalt:       raw.CookieSalt,
		IngestKey:        raw.IngestKey,
		IngestTriggerURL: raw.IngestTriggerURL,
		InitialData: InitialData{
			Tags:      raw.Tags,
			AgeGroups: raw.AgeGroups,
			Regions:   raw.Regions,
			Venues:    raw.Venues,
		},
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
