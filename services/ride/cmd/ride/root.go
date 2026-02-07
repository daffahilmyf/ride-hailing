package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "ride",
	Short: "Ride Service",
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
	rootCmd.PersistentFlags().String("grpc.addr", ":50051", "gRPC listen address")
	rootCmd.PersistentFlags().String("postgres.dsn", "", "PostgreSQL DSN")
	rootCmd.PersistentFlags().Int("shutdown.timeout", 10, "shutdown timeout in seconds")
	rootCmd.PersistentFlags().Int("idempotency.ttl", 86400, "idempotency key TTL in seconds")
	rootCmd.PersistentFlags().String("nats.url", "", "NATS URL")
	rootCmd.PersistentFlags().Bool("outbox.enabled", true, "enable outbox publisher")
	rootCmd.PersistentFlags().Int("outbox.interval_millis", 2000, "outbox polling interval in milliseconds")
	rootCmd.PersistentFlags().Int("outbox.batch_size", 25, "outbox publish batch size")
	rootCmd.PersistentFlags().Int("outbox.max_attempts", 10, "outbox max attempts per message")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("grpc.addr", rootCmd.PersistentFlags().Lookup("grpc.addr"))
	_ = viper.BindPFlag("postgres.dsn", rootCmd.PersistentFlags().Lookup("postgres.dsn"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
	_ = viper.BindPFlag("idempotency.ttl_seconds", rootCmd.PersistentFlags().Lookup("idempotency.ttl"))
	_ = viper.BindPFlag("nats.url", rootCmd.PersistentFlags().Lookup("nats.url"))
	_ = viper.BindPFlag("outbox.enabled", rootCmd.PersistentFlags().Lookup("outbox.enabled"))
	_ = viper.BindPFlag("outbox.interval_millis", rootCmd.PersistentFlags().Lookup("outbox.interval_millis"))
	_ = viper.BindPFlag("outbox.batch_size", rootCmd.PersistentFlags().Lookup("outbox.batch_size"))
	_ = viper.BindPFlag("outbox.max_attempts", rootCmd.PersistentFlags().Lookup("outbox.max_attempts"))
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

	viper.SetEnvPrefix("RIDE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("service.name", infra.DefaultConfig().ServiceName)

	_ = viper.ReadInConfig()
}
