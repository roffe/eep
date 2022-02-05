/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// gendocsCmd represents the gendocs command
var gendocsCmd = &cobra.Command{
	Use:    "gendocs",
	Hidden: true,
	Short:  "generate markdown docs",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootCmd.Root().DisableAutoGenTag = true
		os.Mkdir("./docs", 0755)
		return doc.GenMarkdownTree(rootCmd, "./docs")
	},
}

func init() {
	rootCmd.AddCommand(gendocsCmd)
}
