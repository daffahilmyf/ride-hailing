package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "location",
	Short: "Location Service",
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
	rootCmd.PersistentFlags().String("grpc.addr", ":50053", "gRPC listen address")
	rootCmd.PersistentFlags().Int("shutdown.timeout", 10, "shutdown timeout in seconds")
	rootCmd.PersistentFlags().String("redis.addr", "", "Redis address")
	rootCmd.PersistentFlags().String("redis.password", "", "Redis password")
	rootCmd.PersistentFlags().Int("redis.db", 0, "Redis DB")
	rootCmd.PersistentFlags().Int("location.ttl", 60, "location TTL in seconds")
	rootCmd.PersistentFlags().String("location.key_prefix", "driver:location:", "location key prefix")
	rootCmd.PersistentFlags().String("location.geo_key", "drivers:geo", "location geo key")
	rootCmd.PersistentFlags().Bool("rate_limit.enabled", true, "enable rate limiting")
	rootCmd.PersistentFlags().Int("rate_limit.min_gap_ms", 300, "min gap between updates in ms")
	rootCmd.PersistentFlags().String("rate_limit.key_prefix", "driver:location:rate:", "rate limit key prefix")
	rootCmd.PersistentFlags().Bool("cleanup.enabled", true, "enable stale geo cleanup")
	rootCmd.PersistentFlags().Int("cleanup.interval_seconds", 60, "stale geo cleanup interval in seconds")
	rootCmd.PersistentFlags().Int("cleanup.batch_size", 1000, "stale geo cleanup batch size")
	rootCmd.PersistentFlags().String("nats.url", "", "NATS URL")
	rootCmd.PersistentFlags().Bool("events.enabled", true, "enable location events")
	rootCmd.PersistentFlags().Bool("internal_auth.enabled", false, "enable internal gRPC auth")
	rootCmd.PersistentFlags().String("internal_auth.token", "", "internal auth token")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("grpc.addr", rootCmd.PersistentFlags().Lookup("grpc.addr"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
	_ = viper.BindPFlag("redis.addr", rootCmd.PersistentFlags().Lookup("redis.addr"))
	_ = viper.BindPFlag("redis.password", rootCmd.PersistentFlags().Lookup("redis.password"))
	_ = viper.BindPFlag("redis.db", rootCmd.PersistentFlags().Lookup("redis.db"))
	_ = viper.BindPFlag("location.ttl_seconds", rootCmd.PersistentFlags().Lookup("location.ttl"))
	_ = viper.BindPFlag("location.key_prefix", rootCmd.PersistentFlags().Lookup("location.key_prefix"))
	_ = viper.BindPFlag("location.geo_key", rootCmd.PersistentFlags().Lookup("location.geo_key"))
	_ = viper.BindPFlag("rate_limit.enabled", rootCmd.PersistentFlags().Lookup("rate_limit.enabled"))
	_ = viper.BindPFlag("rate_limit.min_gap_ms", rootCmd.PersistentFlags().Lookup("rate_limit.min_gap_ms"))
	_ = viper.BindPFlag("rate_limit.key_prefix", rootCmd.PersistentFlags().Lookup("rate_limit.key_prefix"))
	_ = viper.BindPFlag("cleanup.enabled", rootCmd.PersistentFlags().Lookup("cleanup.enabled"))
	_ = viper.BindPFlag("cleanup.interval_seconds", rootCmd.PersistentFlags().Lookup("cleanup.interval_seconds"))
	_ = viper.BindPFlag("cleanup.batch_size", rootCmd.PersistentFlags().Lookup("cleanup.batch_size"))
	_ = viper.BindPFlag("nats.url", rootCmd.PersistentFlags().Lookup("nats.url"))
	_ = viper.BindPFlag("events.enabled", rootCmd.PersistentFlags().Lookup("events.enabled"))
	_ = viper.BindPFlag("internal_auth.enabled", rootCmd.PersistentFlags().Lookup("internal_auth.enabled"))
	_ = viper.BindPFlag("internal_auth.token", rootCmd.PersistentFlags().Lookup("internal_auth.token"))
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

	viper.SetEnvPrefix("LOCATION")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("service.name", infra.DefaultConfig().ServiceName)

	_ = viper.ReadInConfig()
}
