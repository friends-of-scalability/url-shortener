package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/friends-of-scalability/url-shortener/cmd/config"
	"github.com/friends-of-scalability/url-shortener/internal/urlshortener"
	"github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"context"
)

func bindEnvironmentVariables() {
	viper.SetEnvPrefix("urlshortener")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("http.addr")
	viper.BindEnv("fakeload")
	viper.BindEnv("storage")
	viper.BindEnv("postgresql.host")
	viper.BindEnv("postgresql.port")
	viper.BindEnv("postgresql.user")
	viper.BindEnv("postgresql.password")
	viper.BindEnv("role")
	viper.BindEnv("sd.resolver")
	viper.BindEnv("sd.shortener")
}

func bindFlags(rootCmd *cobra.Command, c *config.Config) error {
	rootCmd.PersistentFlags().StringVar(&c.HTTPAddress, "http.addr", ":8080", "by default :8080")
	rootCmd.PersistentFlags().BoolVar(&c.EnableFakeLoad, "fakeload", false, "enable it if you want to generate fake load")
	rootCmd.PersistentFlags().StringVar(&c.StorageType, "storage", "inmemory", "storage backend to use [inmemory,postgres]")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.Host, "postgresql.host", "", "Postgres host")
	rootCmd.PersistentFlags().IntVar(&c.Postgresql.Port, "postgresql.port", 5432, "Postgres port")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.User, "postgresql.user", "", "Postgres user")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.Password, "postgresql.password", "", "Postgres password")
	rootCmd.PersistentFlags().StringVar(&c.ServiceDiscovery.Resolver, "sd.resolver", "", "DNS SRV for resolvers")
	rootCmd.PersistentFlags().StringVar(&c.ServiceDiscovery.Shortener, "sd.shortener", "", "DNS SRV for shorteners")
	rootCmd.PersistentFlags().StringVar(&c.Role, "role", "full", "which role will do this instance full|apigateway|resolver|shortener")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}
	bindEnvironmentVariables()

	return nil
}

func initializeConfig(c *config.Config) {
	c.HTTPAddress = viper.GetString("http.addr")
	host, port, err := net.SplitHostPort(c.HTTPAddress)
	if err != nil {
		c.HTTPAddress = ":8080"
		c.ExposedHost = ""
		c.ExposedPort = "8080"
	} else {
		c.ExposedHost = host
		c.ExposedPort = port
	}
	c.EnableFakeLoad = viper.GetBool("fakeload")
	c.StorageType = viper.GetString("storage")
	c.Role = viper.GetString("role")
	c.Postgresql.Host = viper.GetString("postgresql.host")
	c.Postgresql.Port = viper.GetInt("postgresql.port")
	c.Postgresql.User = viper.GetString("postgresql.user")
	c.Postgresql.Password = viper.GetString("postgresql.password")
	c.ServiceDiscovery.Resolver = viper.GetString("sd.resolver")
	c.ServiceDiscovery.Shortener = viper.GetString("sd.shortener")
}

func createCLI(c *config.Config) error {

	var rootCmd = &cobra.Command{
		Use:   "urlshortener",
		Short: "urlshortener CLI",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
	if err := bindFlags(rootCmd, c); err != nil {
		return err
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	initializeConfig(c)
	return nil

}

func main() {

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}
	var cfg config.Config
	{
		err := createCLI(&cfg)
		if err != nil {
			logger.Log("fatal", "config", "error", err)
		}
	}
	var ctx context.Context
	{
		ctx = context.Background()
	}

	var s urlshortener.Service
	{
		var err error
		s, err = urlshortener.NewService(&cfg)
		if err != nil {
			logger.Log("fatal", err)
			os.Exit(1)
		}
		s = urlshortener.NewLoggingService(logger, s)
	}

	var h http.Handler
	{
		switch cfg.Role {
		case "full":
			h = urlshortener.MakeHandler(ctx, s, log.With(logger, "component", "HTTP"))
		case "resolver":
			h = urlshortener.MakeResolverHandler(ctx, s, log.With(logger, "component", "HTTP"))
		case "shortener":
			h = urlshortener.MakeShortenerHandler(ctx, s, log.With(logger, "component", "HTTP"))
		case "apigateway":
			h = urlshortener.MakeAPIGWHandler(ctx, s, log.With(logger, "component", "HTTP"), &cfg)
		}
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "Hystrix Stream Server", "addr", ":9000", "STORAGE", cfg.StorageType, "FAKELOAD", cfg.EnableFakeLoad)

		hystrixStreamHandler := hystrix.NewStreamHandler()
		hystrixStreamHandler.Start()
		errs <- http.ListenAndServe(":9000", hystrixStreamHandler)
	}()
	go func() {
		logger.Log("transport", "HTTP", "addr", cfg.HTTPAddress, "STORAGE", cfg.StorageType, "FAKELOAD", cfg.EnableFakeLoad)
		errs <- http.ListenAndServe(cfg.HTTPAddress, h)

	}()

	logger.Log("exit", <-errs)
}
