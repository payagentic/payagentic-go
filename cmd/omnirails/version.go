package main

import (
	"fmt"
	"runtime"

	raistonpay "github.com/raiston/raistonpay-go"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "omnirails CLI v%s\n", raistonpay.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  SDK:      raistonpay-go v%s\n", raistonpay.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "  Go:       %s\n", runtime.Version())
		fmt.Fprintf(cmd.OutOrStdout(), "  OS/Arch:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
