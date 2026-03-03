package main

import (
	"github.com/celsiainternet/elvis/create/v2"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "go"}
	rootCmd.AddCommand(create.Create)
	rootCmd.Execute()
}
