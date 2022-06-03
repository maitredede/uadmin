package main

import (
	"github.com/maitredede/uadmin"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of uAdmin",
	// Long: "",
	Run: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	uadmin.Trail(uadmin.INFO, uadmin.Version)
}
