package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Key            []byte
	Level          string `default:"info"`
	Address        string `default:"127.0.0.1"`
	Port           int    `default:"7777"`
	DBPath         string `default:"./db.bolt"`
	AdminRole      string `default:"admin"`
	UserRole       string `default:"user"`
	SignAdminToken bool
	SignDuration   time.Duration
	SignEmail      string
	GenerateKey    bool
}

const (
	appEnv          = "APP_ENV"
	defaultKeyUsage = "file holding the signing key for JWT"
)

var (
	KeyFile        string
	defaultKeyPath string
)

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		home = "~"
	}

	defaultKeyPath = home + "/.todo-app.key"
}

func Load() (*Config, error) {
	env := os.Getenv(appEnv)
	if env == "" {
		fmt.Fprintf(os.Stderr, "env var %s is not set. Trying to load config from flags\n\n", appEnv)
	}

	var cfg Config
	err := envconfig.Process(env, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env vars > %w", err)
	}

	flag.Usage = showUsage
	flag.StringVar(&KeyFile, "k", defaultKeyPath, defaultKeyUsage)
	flag.StringVar(&KeyFile, "key-path", defaultKeyPath, defaultKeyUsage)
	flag.BoolFunc("h", "show this help", showHelp)
	flag.BoolFunc("help", "show this help", showHelp)
	flag.BoolVar(&cfg.SignAdminToken, "c", false, "create a new admin JWT token")
	flag.BoolVar(&cfg.SignAdminToken, "create-token", false, "create a new admin JWT token")
	flag.BoolVar(&cfg.GenerateKey, "g", false, fmt.Sprintf("create a new JWT signing key (%s)", defaultKeyPath))
	flag.BoolVar(&cfg.GenerateKey, "generate", false, fmt.Sprintf("create a new JWT signing key (%s)", defaultKeyPath))
	flag.DurationVar(&cfg.SignDuration, "d", time.Minute*15, "duration of the admin JWT token")
	flag.DurationVar(&cfg.SignDuration, "duration", time.Minute*15, "duration of the admin JWT token")
	flag.StringVar(&cfg.SignEmail, "e", "admin@localhost", "email address to use in the JWT token")
	flag.StringVar(&cfg.SignEmail, "email", "admin@localhost", "email address to use in the JWT token")
	flag.Parse()

	file, err := os.Open(KeyFile)
	if err == nil && len(cfg.Key) == 0 {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			key := scanner.Bytes()
			if len(key) > 5 {
				cfg.Key = key
			}
		}
	}

	if len(cfg.Key) == 0 && !cfg.GenerateKey {
		flag.Usage()
		return nil, fmt.Errorf("jwt signing key is missing")
	}

	return &cfg, nil
}

func showHelp(val string) error {
	flag.Usage()
	os.Exit(0)
	return nil
}

func showUsage() {
	fmt.Fprintf(os.Stderr, `%s%s`, banner, summary)
	flag.PrintDefaults()
}

const banner = `
 ____  __  ___    __        __   ___  ___ 
(_  _)/  \(   \  /  \  ___ (  ) (  ,\(  ,\
  )( ( () )) ) )( () )(___)/__\  ) _/ ) _/
 (__) \__/(___/  \__/     (_)(_)(_)  (_)  `

const summary = `


 todo-app is the backend of a web app for creating and sharing To-Do lists
 
 Usage:

 `
