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

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("grpc.addr", rootCmd.PersistentFlags().Lookup("grpc.addr"))
	_ = viper.BindPFlag("postgres.dsn", rootCmd.PersistentFlags().Lookup("postgres.dsn"))
	_ = viper.BindPFlag("shutdown.timeout", rootCmd.PersistentFlags().Lookup("shutdown.timeout"))
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
