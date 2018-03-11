package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
}

func bindFlags(rootCmd *cobra.Command, c *config.Config) error {
	rootCmd.PersistentFlags().StringVar(&c.HTTPAddress, "http.addr", ":8080", "by default :8080")
	rootCmd.PersistentFlags().BoolVar(&c.EnableFakeLoad, "fakeload", false, "enable it if you want to generate fake load")
	rootCmd.PersistentFlags().StringVar(&c.StorageType, "storage", "inmemory", "storage backend to use [inmemory,postgres]")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.Host, "postgresql.host", "", "Postgres host")
	rootCmd.PersistentFlags().IntVar(&c.Postgresql.Port, "postgresql.port", 5432, "Postgres port")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.User, "postgresql.user", "", "Postgres user")
	rootCmd.PersistentFlags().StringVar(&c.Postgresql.Password, "postgresql.password", "", "Postgres password")
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		return err
	}
	bindEnvironmentVariables()
	c.HTTPAddress = viper.GetString("http.addr")
	c.EnableFakeLoad = viper.GetBool("fakeload")
	c.StorageType = viper.GetString("storage")
	c.Postgresql.Host = viper.GetString("postgresql.host")
	c.Postgresql.Port = viper.GetInt("postgresql.port")
	c.Postgresql.User = viper.GetString("postgresql.user")
	c.Postgresql.Password = viper.GetString("postgresql.password")
	return nil
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
		fmt.Printf("%+v\n", cfg)
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
		h = urlshortener.MakeHandler(ctx, s, log.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", cfg.HTTPAddress)
		errs <- http.ListenAndServe(cfg.HTTPAddress, h)

	}()

	logger.Log("exit", <-errs)
}
