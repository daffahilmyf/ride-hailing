package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "user",
	Short: "User Service",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(serveCmd)

	rootCmd.PersistentFlags().String("config", "", "config file (default is ./config/config.yaml)")
	rootCmd.PersistentFlags().String("http.addr", ":8081", "HTTP listen address")
	rootCmd.PersistentFlags().Int("http.read_timeout_seconds", 5, "HTTP read timeout in seconds")
	rootCmd.PersistentFlags().Int("http.write_timeout_seconds", 5, "HTTP write timeout in seconds")
	rootCmd.PersistentFlags().String("grpc.addr", ":50054", "gRPC listen address")
	rootCmd.PersistentFlags().Int("shutdown.timeout", 10, "shutdown timeout in seconds")
	rootCmd.PersistentFlags().String("postgres.dsn", "", "Postgres DSN")
	rootCmd.PersistentFlags().String("auth.jwt_secret", "", "JWT secret")
	rootCmd.PersistentFlags().String("auth.issuer", "ride-hailing", "JWT issuer")
	rootCmd.PersistentFlags().String("auth.audience", "ride-hailing-clients", "JWT audience")
	rootCmd.PersistentFlags().Int("auth.access_ttl_seconds", 1800, "access token TTL in seconds")
	rootCmd.PersistentFlags().Int("auth.refresh_ttl_seconds", 2592000, "refresh token TTL in seconds")
	rootCmd.PersistentFlags().Bool("internal_auth.enabled", false, "enable internal auth")
	rootCmd.PersistentFlags().String("internal_auth.token", "", "internal auth token")
	rootCmd.PersistentFlags().Int("rate_limit.auth_requests", 30, "auth requests per window")
	rootCmd.PersistentFlags().Int("rate_limit.window_seconds", 60, "rate limit window seconds")
	rootCmd.PersistentFlags().Bool("observability.metrics_enabled", true, "enable metrics")
	rootCmd.PersistentFlags().String("observability.metrics_addr", ":9096", "metrics listen addr")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("http.addr", rootCmd.PersistentFlags().Lookup("http.addr"))
	_ = viper.BindPFlag("http.read_timeout_seconds", rootCmd.PersistentFlags().Lookup("http.read_timeout_seconds"))
	_ = viper.BindPFlag("http.write_timeout_seconds", rootCmd.PersistentFlags().Lookup("http.write_timeout_seconds"))
	_ = viper.BindPFlag("grpc.addr", rootCmd.PersistentFlags().Lookup("grpc.addr"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
	_ = viper.BindPFlag("postgres.dsn", rootCmd.PersistentFlags().Lookup("postgres.dsn"))
	_ = viper.BindPFlag("auth.jwt_secret", rootCmd.PersistentFlags().Lookup("auth.jwt_secret"))
	_ = viper.BindPFlag("auth.issuer", rootCmd.PersistentFlags().Lookup("auth.issuer"))
	_ = viper.BindPFlag("auth.audience", rootCmd.PersistentFlags().Lookup("auth.audience"))
	_ = viper.BindPFlag("auth.access_ttl_seconds", rootCmd.PersistentFlags().Lookup("auth.access_ttl_seconds"))
	_ = viper.BindPFlag("auth.refresh_ttl_seconds", rootCmd.PersistentFlags().Lookup("auth.refresh_ttl_seconds"))
	_ = viper.BindPFlag("internal_auth.enabled", rootCmd.PersistentFlags().Lookup("internal_auth.enabled"))
	_ = viper.BindPFlag("internal_auth.token", rootCmd.PersistentFlags().Lookup("internal_auth.token"))
	_ = viper.BindPFlag("rate_limit.auth_requests", rootCmd.PersistentFlags().Lookup("rate_limit.auth_requests"))
	_ = viper.BindPFlag("rate_limit.window_seconds", rootCmd.PersistentFlags().Lookup("rate_limit.window_seconds"))
	_ = viper.BindPFlag("observability.metrics_enabled", rootCmd.PersistentFlags().Lookup("observability.metrics_enabled"))
	_ = viper.BindPFlag("observability.metrics_addr", rootCmd.PersistentFlags().Lookup("observability.metrics_addr"))
}

func initConfig() {
	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
	}

	viper.SetEnvPrefix("USER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("service.name", infra.DefaultConfig().ServiceName)

	_ = viper.ReadInConfig()
}
