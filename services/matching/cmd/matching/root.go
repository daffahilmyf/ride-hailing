package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/daffahilmyf/ride-hailing/services/matching/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "matching",
	Short: "Matching Service",
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
	rootCmd.PersistentFlags().String("grpc.addr", ":50052", "gRPC listen address")
	rootCmd.PersistentFlags().Int("shutdown.timeout", 10, "shutdown timeout in seconds")
	rootCmd.PersistentFlags().String("redis.addr", "", "Redis address")
	rootCmd.PersistentFlags().String("redis.password", "", "Redis password")
	rootCmd.PersistentFlags().Int("redis.db", 0, "Redis DB")
	rootCmd.PersistentFlags().String("matching.geo_key", "drivers:geo", "geo key")
	rootCmd.PersistentFlags().String("matching.status_key", "drivers:status", "driver status key")
	rootCmd.PersistentFlags().String("matching.available_key", "drivers:available", "available set key")
	rootCmd.PersistentFlags().String("matching.offer_key_prefix", "driver:offer:", "offer key prefix")
	rootCmd.PersistentFlags().Int("matching.offer_ttl_seconds", 10, "offer TTL seconds")
	rootCmd.PersistentFlags().Float64("matching.radius_meters", 3000, "matching radius meters")
	rootCmd.PersistentFlags().Int("matching.limit", 5, "matching candidate limit")
	rootCmd.PersistentFlags().String("nats.url", "", "NATS URL")
	rootCmd.PersistentFlags().Bool("events.enabled", true, "enable event consumption")
	rootCmd.PersistentFlags().String("events.ride_requested_subject", "ride.requested", "ride requested subject")
	rootCmd.PersistentFlags().String("events.driver_location_subject", "driver.location.updated", "driver location subject")
	rootCmd.PersistentFlags().Bool("internal_auth.enabled", false, "enable internal gRPC auth")
	rootCmd.PersistentFlags().String("internal_auth.token", "", "internal auth token")
	rootCmd.PersistentFlags().String("ride.addr", "", "ride service address")
	rootCmd.PersistentFlags().String("ride.internal_token", "", "ride service internal token")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("grpc.addr", rootCmd.PersistentFlags().Lookup("grpc.addr"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
	_ = viper.BindPFlag("redis.addr", rootCmd.PersistentFlags().Lookup("redis.addr"))
	_ = viper.BindPFlag("redis.password", rootCmd.PersistentFlags().Lookup("redis.password"))
	_ = viper.BindPFlag("redis.db", rootCmd.PersistentFlags().Lookup("redis.db"))
	_ = viper.BindPFlag("matching.geo_key", rootCmd.PersistentFlags().Lookup("matching.geo_key"))
	_ = viper.BindPFlag("matching.status_key", rootCmd.PersistentFlags().Lookup("matching.status_key"))
	_ = viper.BindPFlag("matching.available_key", rootCmd.PersistentFlags().Lookup("matching.available_key"))
	_ = viper.BindPFlag("matching.offer_key_prefix", rootCmd.PersistentFlags().Lookup("matching.offer_key_prefix"))
	_ = viper.BindPFlag("matching.offer_ttl_seconds", rootCmd.PersistentFlags().Lookup("matching.offer_ttl_seconds"))
	_ = viper.BindPFlag("matching.radius_meters", rootCmd.PersistentFlags().Lookup("matching.radius_meters"))
	_ = viper.BindPFlag("matching.limit", rootCmd.PersistentFlags().Lookup("matching.limit"))
	_ = viper.BindPFlag("nats.url", rootCmd.PersistentFlags().Lookup("nats.url"))
	_ = viper.BindPFlag("events.enabled", rootCmd.PersistentFlags().Lookup("events.enabled"))
	_ = viper.BindPFlag("events.ride_requested_subject", rootCmd.PersistentFlags().Lookup("events.ride_requested_subject"))
	_ = viper.BindPFlag("events.driver_location_subject", rootCmd.PersistentFlags().Lookup("events.driver_location_subject"))
	_ = viper.BindPFlag("internal_auth.enabled", rootCmd.PersistentFlags().Lookup("internal_auth.enabled"))
	_ = viper.BindPFlag("internal_auth.token", rootCmd.PersistentFlags().Lookup("internal_auth.token"))
	_ = viper.BindPFlag("ride.addr", rootCmd.PersistentFlags().Lookup("ride.addr"))
	_ = viper.BindPFlag("ride.internal_token", rootCmd.PersistentFlags().Lookup("ride.internal_token"))
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

	viper.SetEnvPrefix("MATCHING")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("service.name", infra.DefaultConfig().ServiceName)

	_ = viper.ReadInConfig()
}
