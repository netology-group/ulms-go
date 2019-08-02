package main

import (
	"github.com/netology-group/ulms-go/pkg/app"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"os"
)

var (
	config      string
	environment string
	dir         string
	table       string
	schema      string
)

func init() {
	pflag.StringVarP(&config, "config", "c", "configs/config.yml", "path to config file")
	pflag.StringVarP(&environment, "environment", "e", "development", "sql-migrate environment to use")
	pflag.StringVarP(&dir, "dir", "d", "migrations", "path to dir with migrations")
	pflag.StringVarP(&table, "table", "t", "gorp_migrations", "table storing applied migrations")
	pflag.StringVarP(&schema, "schema", "s", "", "DB schema to use when applying migrations")
}

func main() {
	pflag.Parse()
	config, err := app.LoadConfig(config)
	if err != nil {
		logrus.WithError(err).Fatal("can't read config file")
	}
	dbConfig := map[string]*migrationsEnvironment{
		environment: {
			Dialect:    config.Db.Driver,
			DataSource: config.Db.DataSource,
			Dir:        dir,
			TableName:  table,
			SchemaName: schema,
		},
	}
	encoder := yaml.NewEncoder(os.Stdout)
	if err := encoder.Encode(dbConfig); err != nil {
		logrus.WithError(err).Fatal("can't write config file")
	}
}

type migrationsEnvironment struct {
	Dialect    string `yaml:"dialect"`
	DataSource string `yaml:"datasource"`
	Dir        string `yaml:"dir"`
	TableName  string `yaml:"table"`
	SchemaName string `yaml:"schema"`
}
