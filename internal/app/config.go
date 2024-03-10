package app

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"os"
	"path/filepath"
)

const defaultErrorLevelName = "debug"

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevelName         string `env:"LOG_LEVEL"`
	ProjectRootDir       string `env:"GM_ROOT_DIR"`
	//SecretKeyFilePath    string `env:"SECRET_KEY_FILE_PATH"`
	// Время жизни токена
	//TokenTTL time.Duration `env:"TOKEN_TTL"`
}

var defaultConfig = Config{
	RunAddress:           ":8080",
	DatabaseURI:          "postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable",
	AccrualSystemAddress: "http://localhost:8082", // @todo
	LogLevelName:         defaultErrorLevelName,
	ProjectRootDir:       "",
	//SecretKeyFilePath:    "./secret.key",
	//TokenTTL:             time.Duration(time.Hour),
}

func GetConfig() (cnf Config, err error) {
	flag.StringVar(&cnf.RunAddress, "a", defaultConfig.RunAddress, "run address")
	flag.StringVar(&cnf.DatabaseURI, "d", defaultConfig.DatabaseURI, "db uri")
	flag.StringVar(&cnf.AccrualSystemAddress, "r", defaultConfig.AccrualSystemAddress, "accrual system address")

	flag.StringVar(&cnf.LogLevelName, "log-level", defaultConfig.LogLevelName, "log level(error,warning, debug e.t.c)")
	flag.StringVar(&cnf.ProjectRootDir, "root-dir", defaultConfig.ProjectRootDir, "project root directory")

	flag.Parse()

	envConfig := Config{}
	if err := env.Parse(&envConfig); err != nil {
		return Config{}, fmt.Errorf("can't parse env vars: %w", err)
	} else {
		if envConfig.RunAddress != "" {
			cnf.RunAddress = envConfig.RunAddress
		}
		if envConfig.DatabaseURI != "" {
			cnf.DatabaseURI = envConfig.DatabaseURI
		}
		if envConfig.AccrualSystemAddress != "" {
			cnf.AccrualSystemAddress = envConfig.AccrualSystemAddress
		}
		if envConfig.LogLevelName != "" {
			cnf.LogLevelName = envConfig.LogLevelName
		}
		if envConfig.ProjectRootDir != "" {
			cnf.ProjectRootDir = envConfig.ProjectRootDir
		}
	}

	if cnf.ProjectRootDir == defaultConfig.ProjectRootDir {
		rootDir, err := getProjectRootDir()
		if err != nil {
			return Config{}, fmt.Errorf("can't evalute project root dir: %w", err)
		}
		cnf.ProjectRootDir = rootDir
	}
	return cnf, nil
}

func getProjectRootDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	rootDir, err := filepath.EvalSymlinks(filepath.Dir(execPath) + "/../..")
	if err != nil {
		return "", fmt.Errorf("can't evalute root path. Executable path: %s. Error: %w", execPath, err)
	}
	return rootDir, nil
}
