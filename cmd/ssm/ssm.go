package ssm

import (
	"context"
	"log"

	"github.com/harleymckenzie/asc/pkg/service/ssm"
	"github.com/spf13/cobra"
)


var (
	selectedColumns []string
)

func NewSSMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssm",
		Short: "Perform SSM operations",
	}

	// ls sub command
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List parameeters in Systems Manager Parameter Store",
        PreRun: func(cobraCmd *cobra.Command, args []string) {
            selectedColumns = []string{
                "type",
                "name",
            }
        },
        Run: func(cobraCmd *cobra.Command, args []string) {
            ctx := context.TODO()
            profile, _ := cobraCmd.Root().PersistentFlags().GetString("profile")

            svc, err := ssm.NewSSMService(ctx, profile)
            if err != nil {
                log.Fatalf("Failed to initialize SSM service: %v", err)
            }

            if err := svc.ListParameters(ctx, selectedColumns); err != nil {
                log.Fatalf("Failed to list parameters: %v", err)
            }
        },
	}
    cmd.AddCommand(lsCmd)

    return cmd
}