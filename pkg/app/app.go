package app

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/stdlib" // PostgreSQL driver
	"github.com/jmoiron/sqlx"
	"github.com/netology-group/ulms-auth-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// App is an application instance
type App struct {
	Config     *Config
	auth       auth.Auth
	Db         *sqlx.DB
	router     *mux.Router
	goroutines sync.WaitGroup
}

// New instance of application
func New(configFile string) *App {
	config, err := LoadConfig(configFile)
	if err != nil {
		logrus.WithError(err).Panic("can't load config")
	}
	if err := sentry.Init(config.Sentry); err != nil {
		logrus.Errorf("Sentry initialization failed: %v", err)
	}
	authConfig, err := loadAuthConfig(config.Auth)
	if err != nil {
		logrus.WithError(err).Panic("can't load auth config")
	}
	app := &App{
		Config: config,
		auth:   authConfig,
		Db:     sqlx.MustOpen(config.Db.Driver, config.Db.DataSource),
		router: mux.NewRouter(),
	}
	app.Db.SetMaxOpenConns(config.Db.MaxOpenConns)
	app.Db.SetMaxIdleConns(config.Db.MaxIdleConns)
	app.setRoutes()
	return app
}

// Run application on specified port
func (app *App) Run(port string) (err error) {
	server := &http.Server{Addr: port, Handler: app.router}
	stopChannel := make(chan os.Signal, 1)
	go func() {
		if err = server.ListenAndServe(); err != nil {
			stopChannel <- os.Interrupt
		}
	}()
	signal.Notify(stopChannel, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stopChannel
	if err != nil {
		return
	}
	logrus.WithField("signal", stopSignal).Info("received stop signal")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err = server.Shutdown(ctx); err == nil {
		err = app.wait(ctx)
	}
	return
}

func (app *App) wait(ctx context.Context) error {
	finished := make(chan bool, 1)
	go func() {
		app.goroutines.Wait()
		finished <- true
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-finished:
		return nil
	}
}

// Config stores App configuration parameters
type Config struct {
	Auth string `json:"auth"`
	Db   struct {
		Driver       string `yaml:"driver"`
		DataSource   string `yaml:"source"`
		MaxOpenConns int    `yaml:"max-open-conns"`
		MaxIdleConns int    `yaml:"max-idle-conns"`
	}
	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
		MaxAge         int      `yaml:"max_age"`
	} `yaml:"cors"`
	Sentry sentry.ClientOptions `yaml:"sentry"`
}

// LoadConfig reads configuration parameters from the specified file
func LoadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func loadAuthConfig(file string) (auth.Auth, error) {
	authConfig := &auth.TenantAuth{}
	if _, err := toml.DecodeFile(file, authConfig); err != nil {
		return nil, err
	}
	return authConfig, nil
}
