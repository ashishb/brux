package cmd

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ashishb/brux/src/brux/internal/brurunner"
)

var (
	_filePath       string
	_saveOutput     bool
	_outputFilePath *string
	_envName        *string
	_prettyPrint    *bool
)

var _runCmd = &cobra.Command{
	Use:   "run <bruFilePath>",
	Short: "Run a Bru file",
	Long:  `Run a Bru file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_filePath = args[0]
		log.Debug().
			Str("filePath", _filePath).
			Msg("Running bru file")
		cfg, err := brurunner.NewConfig(_filePath, _saveOutput, *_outputFilePath, *_envName, *_prettyPrint)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error creating config")
			os.Exit(1)
		}
		if err := brurunner.Run(context.Background(), *cfg); err != nil {
			log.Error().
				Err(err).
				Msg("Error running bru file")
		}
	},
}

func init() {
	_runCmd.Flags().BoolVarP(&_saveOutput, "save-output", "s", true, "Save output to a file")
	_outputFilePath = _runCmd.Flags().StringP("output-file", "o", "", "Output file path (defaults to a file in tmp dir")
	_envName = _runCmd.Flags().StringP("env", "e", "", "Environment name (name of the sub-dir under the 'environments' directory)")
	_prettyPrint = _runCmd.Flags().BoolP("pretty-print", "p", true, "Pretty print the output")
	RootCmd.AddCommand(_runCmd)
}
