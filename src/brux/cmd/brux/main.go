package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ashishb/brux/src/brux/cmd/brux/cmd"
	"github.com/ashishb/brux/src/brux/internal/logger"
)

func main() {
	logger.ConfigureLogging()

	viper.SetDefault("author", "Ashish Bhatia")
	viper.AutomaticEnv()
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Warn().
			Err(err).
			Msg("Error executing root command")
		os.Exit(1)
	}
}
