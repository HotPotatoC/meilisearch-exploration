package config

// Config is the global config for the application
type Config struct {
	AppName     string
	Host        string
	Port        string
	Environment string

	MeiliMasterkey string
	MeiliHost      string
}

var config *Config

// Init initializes the config package
func Init() {
	config = &Config{
		AppName:     LookupEnv("APP_NAME", "meilisearch-exploration"),
		Host:        LookupEnv("HOST", "0.0.0.0"),
		Port:        LookupEnv("PORT", "9000"),
		Environment: LookupEnv("ENVIRONMENT", "development"),

		MeiliMasterkey: LookupEnv("MEILI_MASTER_KEY", ""),
		MeiliHost:      LookupEnv("MEILI_HOST", ""),
	}
}

// GetConfig returns the global config
func GetConfig() *Config { return config }

// AppName returns the app name
func AppName() string { return config.AppName }

// Host returns the host
func Host() string { return config.Host }

// Port returns the port
func Port() string { return config.Port }

// Environment returns the environment
func Environment() string { return config.Environment }

func MeiliMasterkey() string { return config.MeiliMasterkey }
func MeiliHost() string      { return config.MeiliHost }
