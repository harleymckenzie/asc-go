package cmd

import (
	"os"

	"github.com/harleymckenzie/asc/cmd/ec2"
	"github.com/harleymckenzie/asc/cmd/elasticache"
	"github.com/harleymckenzie/asc/cmd/rds"
	"github.com/harleymckenzie/asc/cmd/ssm"
	"github.com/spf13/cobra"
)

// Global variable to store the profile value
var (
	Profile string
	Version = "0.0.4"
)

func NewRootCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "asc",
		Short: "AWS Simple CLI (asc) - A simplified interface for AWS operations",
	}
	cmds.PersistentFlags().StringVarP(&Profile, "profile", "p", "", "AWS profile to use")
	cmds.Version = Version

	cmds.AddCommand(ec2.NewEC2Cmd())
	cmds.AddCommand(rds.NewRDSCmd())
	cmds.AddCommand(elasticache.NewElasticacheCmd())
	cmds.AddCommand(ssm.NewSSMCmd())
	return cmds
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
