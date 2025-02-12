package ec2

import (
	"context"
	"log"

	"github.com/harleymckenzie/asc/pkg/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sortOrder       []string
	list            bool
	showLaunchTime  bool
	selectedColumns []string
)

func NewEC2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ec2",
		Short: "Perform EC2 operations",
	}

	// ls sub command
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all EC2 instances",
		Long:  "List all EC2 instances. Sort flags can be combined (e.g. -iTn) to define multiple sort orders, where the order of the flags determines the sort priority.",
		PreRun: func(cobraCmd *cobra.Command, args []string) {
			// Clear any existing sort order
			sortOrder = []string{}

			// Set default columns
			selectedColumns = []string{
				"name",
				"instance_id",
				"state",
				"instance_type",
				"public_ip",
			}

			if showLaunchTime {
				selectedColumns = append(selectedColumns, "launch_time")
			}

			// Visit flags in the order they appear in the command line
			cobraCmd.Flags().Visit(func(f *pflag.Flag) {
				switch f.Name {
				case "sort-name":
					sortOrder = append(sortOrder, "Name")
				case "sort-id":
					sortOrder = append(sortOrder, "Instance Id")
				case "sort-type":
					sortOrder = append(sortOrder, "Type")
				case "sort-launch-time":
					sortOrder = append(sortOrder, "Launch Time")
				}
			})
		},
		Run: func(cobraCmd *cobra.Command, args []string) {
			ctx := context.TODO()
			profile, _ := cobraCmd.Root().PersistentFlags().GetString("profile")
			region, _ := cobraCmd.Root().PersistentFlags().GetString("region")

			svc, err := ec2.NewEC2Service(ctx, profile, region)
			if err != nil {
				log.Fatalf("Failed to initialize EC2 service: %v", err)
			}

			if err := svc.ListInstances(ctx, sortOrder, list, selectedColumns); err != nil {
				log.Fatalf("Failed to list EC2 instances: %v", err)
			}
		},
	}
	cmd.AddCommand(lsCmd)

	// Add flags - Output
	lsCmd.Flags().BoolVarP(&list, "list", "l", false, "Outputs EC2 instances in list format.")
	lsCmd.Flags().BoolVarP(&showLaunchTime, "launch-time", "L", false, "Show the launch time of the instance.")

	// Add flags - Sorting
	lsCmd.Flags().BoolP("sort-name", "n", true, "Sort by descending EC2 instance name.")
	lsCmd.Flags().BoolP("sort-id", "i", false, "Sort by descending EC2 instance Id.")
	lsCmd.Flags().BoolP("sort-type", "T", false, "Sort by descending EC2 instance type.")
	lsCmd.Flags().BoolP("sort-launch-time", "t", false, "Sort by descending launch time (most recently launched first).")
	lsCmd.Flags().SortFlags = false

	return cmd
}
