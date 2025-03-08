package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "brux",
	Short: "Brux is a CLI tool for interacting with Bru files",
	Long:  `Brux is a CLI tool for interacting with Bru files by Ashish Bhatia (https://ashishb.net/)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
