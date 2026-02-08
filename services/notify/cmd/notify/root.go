package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/daffahilmyf/ride-hailing/services/notify/internal/infra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "notify",
	Short: "Notification Service",
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
	rootCmd.PersistentFlags().String("http.addr", ":8090", "HTTP listen address")
	rootCmd.PersistentFlags().Int("shutdown.timeout", 10, "shutdown timeout in seconds")
	rootCmd.PersistentFlags().String("nats.url", "", "NATS URL")
	rootCmd.PersistentFlags().Bool("nats.self_heal", true, "enable NATS self-heal")
	rootCmd.PersistentFlags().Bool("events.enabled", true, "enable event consumption")
	rootCmd.PersistentFlags().String("events.ride_subject", "ride.>", "ride events subject")
	rootCmd.PersistentFlags().String("events.driver_subject", "driver.>", "driver events subject")
	rootCmd.PersistentFlags().Int("sse.buffer_size", 64, "SSE channel buffer size")
	rootCmd.PersistentFlags().Int("sse.keepalive_seconds", 15, "SSE keepalive interval in seconds")
	rootCmd.PersistentFlags().Int("sse.replay_buffer_size", 256, "SSE replay buffer size")
	rootCmd.PersistentFlags().Bool("observability.metrics_enabled", true, "enable metrics endpoint")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("http.addr", rootCmd.PersistentFlags().Lookup("http.addr"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
	_ = viper.BindPFlag("nats.url", rootCmd.PersistentFlags().Lookup("nats.url"))
	_ = viper.BindPFlag("nats.self_heal", rootCmd.PersistentFlags().Lookup("nats.self_heal"))
	_ = viper.BindPFlag("events.enabled", rootCmd.PersistentFlags().Lookup("events.enabled"))
	_ = viper.BindPFlag("events.ride_subject", rootCmd.PersistentFlags().Lookup("events.ride_subject"))
	_ = viper.BindPFlag("events.driver_subject", rootCmd.PersistentFlags().Lookup("events.driver_subject"))
	_ = viper.BindPFlag("sse.buffer_size", rootCmd.PersistentFlags().Lookup("sse.buffer_size"))
	_ = viper.BindPFlag("sse.keepalive_seconds", rootCmd.PersistentFlags().Lookup("sse.keepalive_seconds"))
	_ = viper.BindPFlag("sse.replay_buffer_size", rootCmd.PersistentFlags().Lookup("sse.replay_buffer_size"))
	_ = viper.BindPFlag("observability.metrics_enabled", rootCmd.PersistentFlags().Lookup("observability.metrics_enabled"))
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

	viper.SetEnvPrefix("NOTIFY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("service.name", infra.DefaultConfig().ServiceName)

	_ = viper.ReadInConfig()
}
