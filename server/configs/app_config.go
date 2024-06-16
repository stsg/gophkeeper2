package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/stsg/gophkeeper2/pkg/logger"
)

const (
	defaultPort     = ":42000"
	defaultDBConfig = ""
)

type AppConfig struct {
	log              *zap.SugaredLogger
	ServerPort       string `env:"SERVER_PORT" json:"server_port"`
	TokenKey         string `env:"TOKEN_KEY" json:"token_key"`
	DBConnection     string `env:"DV_CONNECTION" json:"db_connection"`
	DBMaxConnections int    `env:"DB_MAX_CONNECTIONS" json:"db_max_connections"`
}

func InitAppConfig(configPath string) (*AppConfig, error) {
	config, err := readConfig(configPath)
	if err != nil {
		return nil, err
	}
	config.setupConfigByFlags()
	return config, nil
}

func readConfig(configFilePath string) (*AppConfig, error) {
	if configFilePath == "" {
		return nil, errors.New("failed to init configuration: file path is not specified")
	}
	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configFile by '%s': %v", configFilePath, err)
	}
	var config AppConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config json '%s': %v", string(configBytes), err)
	}
	config.log = logger.NewLogger("app-config")
	return &config, nil
}

func (cfg *AppConfig) setupConfigByFlags() {
	cfg.log.Infof("Reading config flags")
	var serverPortF string
	pflag.StringVarP(&serverPortF, "a", "a", defaultPort, "Port of the proto server")

	var dbDsnF string
	pflag.StringVarP(&dbDsnF, "d", "d", defaultDBConfig, "Postgres DB DSN")

	var cryptoKeyF string
	pflag.StringVarP(&cryptoKeyF, "crypto-key", "c", "", "Path to private key")

	var dbMaxConnF string
	pflag.StringVarP(&dbMaxConnF, "t", "t", "", "DB Max connections")

	var tokenKeyF string
	pflag.StringVarP(&tokenKeyF, "tk", "k", "", "Token key")

	pflag.Parse()

	if cfg.ServerPort != "" && serverPortF != "" {
		cfg.ServerPort = serverPortF
	}
	if dbDsnF != "" {
		cfg.DBConnection = dbDsnF
	}
	if dbMaxConnF != "" {
		cfg.DBMaxConnections, _ = strconv.Atoi(dbMaxConnF)
	}
	if tokenKeyF != "" {
		cfg.TokenKey = tokenKeyF
	}
}
