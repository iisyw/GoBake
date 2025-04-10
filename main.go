package main

import (
	"fmt"
	"os"

	"gobake/internal/cli"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gobake",
		Short: "GoBake - Go Multi-Platform Build Tool",
		Long: `GoBake is a flexible Go build tool that supports multiple platforms and architectures.
It provides an interactive CLI interface for building Go projects for different
target platforms (Windows/Linux) and architectures (AMD64/ARM64).`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.StartInteractiveBuild(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}