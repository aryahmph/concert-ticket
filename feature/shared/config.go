package shared

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"os"
)

type pgConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     uint   `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"db_name" json:"db_name"`
	SslMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	MinConn  uint   `yaml:"min_conn" json:"min_conn"`
	MaxConn  uint   `yaml:"max_conn" json:"max_conn"`
}

func (p pgConfig) ConnStr() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%d database=%s sslmode=%s pool_min_conns=%d pool_max_conns=%d",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SslMode, p.MinConn, p.MaxConn,
	)
}

type httpConfig struct {
	Port         uint `yaml:"port" json:"port"`
	ReadTimeout  uint `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout uint `yaml:"write_timeout" json:"write_timeout"`
}

func (l httpConfig) Addr() string {
	return fmt.Sprintf(":%d", l.Port)
}

type redisConfig struct {
	Addr    string `yaml:"addr" json:"addr"`
	MinConn uint   `yaml:"min_conn" json:"min_conn"`
}

type queueConfig struct {
	Concurrent uint `yaml:"concurrent" json:"concurrent"`
}

type midtransConfig struct {
	ServerKey string `yaml:"server_key" json:"server_key"`
}

type order struct {
	CancellationDuration uint `yaml:"cancellation_duration_second" json:"cancellation_duration_second"`
}

type Config struct {
	Http     httpConfig     `yaml:"http" json:"http"`
	Database pgConfig       `yaml:"db" json:"db"`
	Cache    redisConfig    `yaml:"cache" json:"cache"`
	Queue    queueConfig    `yaml:"queue" json:"queue"`
	Midtrans midtransConfig `yaml:"midtrans" json:"midtrans"`
	Order    order          `yaml:"order" json:"order"`
}

func loadConfigFromReader(r io.Reader, c *Config) error {
	return yaml.NewDecoder(r).Decode(c)
}

func loadConfigFromFile(fn string, c *Config) error {
	_, err := os.Stat(fn)

	if err != nil {
		return err
	}

	f, err := os.Open(fn)

	if err != nil {
		return err
	}

	defer f.Close()

	return loadConfigFromReader(f, c)
}

func LoadConfig(fn string) Config {
	cfg := Config{}
	err := loadConfigFromFile(fn, &cfg)
	if err != nil {
		panic(err)
	}

	slog.Debug("config loaded", slog.Any("config", cfg))
	return cfg
}
