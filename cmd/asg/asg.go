package asg

import (
	"context"
	"log"

	"github.com/harleymckenzie/asc/pkg/service/asg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sortOrder       []string
	list            bool
	showARNs        bool
	selectedColumns []string
)

func NewASGCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asg",
		Short: "Perform ASG operations",
	}

	// ls sub command
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all Auto-Scaling Groups",
		Long:  "List all Auto-Scaling Groups. Sort flags can be combined to define multiple sort orders, where the order of the flags determines the sort priority.",
		PreRun: func(cobraCmd *cobra.Command, args []string) {
			// Clear any existing sort order
			sortOrder = []string{}

			// Set default columns
			selectedColumns = []string{
				"name",
				"instances",
				"desired_capacity",
				"min_capacity",
				"max_capacity",
			}
			if showARNs {
				selectedColumns = append(selectedColumns, "arn")
			}

			// Visit flags in the order they appear in the command line
			cobraCmd.Flags().Visit(func(f *pflag.Flag) {
				switch f.Name {
				case "sort-name":
					sortOrder = append(sortOrder, "Name")
				case "sort-instances":
					sortOrder = append(sortOrder, "Instances")
				case "sort-desired-capacity":
					sortOrder = append(sortOrder, "Desired Capacity")
				case "sort-min-capacity":
					sortOrder = append(sortOrder, "Min Capacity")
				case "sort-max-capacity":
					sortOrder = append(sortOrder, "Max Capacity")
				}
			})
		},
		Run: func(cobraCmd *cobra.Command, args []string) {
			ctx := context.TODO()
			profile, _ := cobraCmd.Root().PersistentFlags().GetString("profile")
			region, _ := cobraCmd.Root().PersistentFlags().GetString("region")

			svc, err := asg.NewAutoScalingService(ctx, profile, region)
			if err != nil {
				log.Fatalf("Failed to initialize ASG service: %v", err)
			}

			if err := svc.ListAutoScalingGroups(ctx, sortOrder, list, selectedColumns); err != nil {
				log.Fatalf("Failed to list Auto-Scaling Groups: %v", err)
			}
		},
	}
	cmd.AddCommand(lsCmd)

	// Add flags - Output
	lsCmd.Flags().BoolVarP(&list, "list", "l", false, "Outputs Auto-Scaling Groups in list format.")
	lsCmd.Flags().BoolVar(&showARNs, "arn", false, "Show ARNs for each Auto-Scaling Group.")

	// Add flags - Sorting
	lsCmd.Flags().BoolP("sort-name", "n", true, "Sort by descending ASG name.")
	lsCmd.Flags().BoolP("sort-instances", "i", false, "Sort by descending number of instances.")
	lsCmd.Flags().BoolP("sort-desired-capacity", "d", false, "Sort by descending desired capacity.")
	lsCmd.Flags().BoolP("sort-min-capacity", "m", false, "Sort by descending min capacity.")
	lsCmd.Flags().BoolP("sort-max-capacity", "M", false, "Sort by descending max capacity.")
	lsCmd.Flags().SortFlags = false

	return cmd
}
