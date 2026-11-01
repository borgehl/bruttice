package configx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultPagesDir is where pages are read from unless config.yaml says otherwise.
const DefaultPagesDir = "/app/pages"

// AdminAccountsEnv, when set, replaces the admin-accounts list of config.yaml. It holds
// "email1;password1,email2;password2": accounts are separated by "," and the email from the
// password by ";", so neither may appear in a password.
const AdminAccountsEnv = "ADMIN_ACCOUNTS"

// splitAdminAccountsEnv splits the value of AdminAccountsEnv into email;password entries,
// trimming whitespace and skipping empty ones such as after a trailing comma.
func splitAdminAccountsEnv(value string) []string {
	var accounts []string
	for _, entry := range strings.Split(value, ",") {
		if entry = strings.TrimSpace(entry); entry != "" {
			accounts = append(accounts, entry)
		}
	}
	return accounts
}

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

// ConfigRawSite is what differs between deployments of the site, shown to visitors.
type ConfigRawSite struct {
	Name         string `yaml:"name"`
	Region       string `yaml:"region"`
	ContactEmail string `yaml:"contact-email"`
	Description  string `yaml:"description"`
}

type ConfigRaw struct {
	AdminAccounts    []string            `yaml:"admin-accounts"`
	Timezone         string              `yaml:"timezone"`
	CookieSalt       string              `yaml:"cookie-salt"`
	IngestKey        string              `yaml:"ingest-key"`
	IngestTriggerURL string              `yaml:"ingest-trigger-url"`
	PagesDir         string              `yaml:"pages-dir"`
	Site             ConfigRawSite       `yaml:"site"`
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
	Site             ConfigRawSite
	// PagesDir holds the deployment's own pages, see package pagesx
	PagesDir    string
	InitialData InitialData
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

	pagesDir := raw.PagesDir
	if pagesDir == "" {
		pagesDir = DefaultPagesDir
	}

	ret := Config{
		AdminAccounts:    make([]AdminAccount, 0, len(raw.AdminAccounts)),
		CookieSalt:       raw.CookieSalt,
		IngestKey:        raw.IngestKey,
		IngestTriggerURL: raw.IngestTriggerURL,
		Site:             raw.Site,
		PagesDir:         pagesDir,
		InitialData: InitialData{
			Tags:      raw.Tags,
			AgeGroups: raw.AgeGroups,
			Regions:   raw.Regions,
			Venues:    raw.Venues,
		},
	}

	accounts := raw.AdminAccounts
	if env := os.Getenv(AdminAccountsEnv); strings.TrimSpace(env) != "" {
		accounts = splitAdminAccountsEnv(env)
	}
	for i, account := range accounts {
		split := strings.Split(account, ";")
		if len(split) != 2 {
			// no value in the message: it holds a password
			panic(fmt.Errorf("malformed admin account #%d, want email;password", i+1))
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
