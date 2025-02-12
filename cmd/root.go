package cmd

import (
	"os"

	"github.com/harleymckenzie/asc/cmd/asg"
	"github.com/harleymckenzie/asc/cmd/ec2"
	"github.com/harleymckenzie/asc/cmd/elasticache"
	"github.com/harleymckenzie/asc/cmd/rds"
	"github.com/spf13/cobra"
)

// Global variable to store the profile value
var (
	Profile string
	Region  string
	Version = "0.0.5"
)

func NewRootCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "asc",
		Short: "AWS Simple CLI (asc) - A simplified interface for AWS operations",
	}
	cmds.PersistentFlags().StringVarP(&Profile, "profile", "p", "", "AWS profile to use")
	cmds.PersistentFlags().StringVarP(&Region, "region", "r", "", "AWS region to use")
	cmds.Version = Version

	cmds.AddCommand(asg.NewASGCmd())
	cmds.AddCommand(ec2.NewEC2Cmd())
	cmds.AddCommand(rds.NewRDSCmd())
	cmds.AddCommand(elasticache.NewElasticacheCmd())

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
