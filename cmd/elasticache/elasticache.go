package elasticache

import (
	"context"
	"log"

	"github.com/harleymckenzie/asc/pkg/service/elasticache"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	sortOrder       []string
	list            bool
	selectedColumns []string
	showEndpoint    bool
)

func NewElasticacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "elasticache",
		Short: "Perform Elasticache operations",
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all Elasticache clusters",
		PreRun: func(cobraCmd *cobra.Command, args []string) {
			// Clear any existing sort order
			sortOrder = []string{}
			
			// Set default columns
			selectedColumns = []string{
				"name",
				"status",
				"engine_version",
				"instance_type",
			}
			if showEndpoint {
				selectedColumns = append(selectedColumns, "endpoint")
			}

			// Visit flags in the order they appear in the command line
			cobraCmd.Flags().Visit(func(f *pflag.Flag) {
				switch f.Name {
				case "sort-name":
					sortOrder = append(sortOrder, "Cache Name")
				case "sort-type":
					sortOrder = append(sortOrder, "Configuration")
				case "sort-status":
					sortOrder = append(sortOrder, "Status")
				case "sort-engine":
					sortOrder = append(sortOrder, "Engine Version")
				}
			})
		},
		Run: func(cobraCmd *cobra.Command, args []string) {
			ctx := context.TODO()
			profile, _ := cobraCmd.Root().PersistentFlags().GetString("profile")
			region, _ := cobraCmd.Root().PersistentFlags().GetString("region")

			svc, err := elasticache.NewElasticacheService(ctx, profile, region)
			if err != nil {
				log.Fatalf("Failed to initialize Elasticache service: %v", err)
			}

			err = svc.ListInstances(ctx, sortOrder, list, selectedColumns)
			if err != nil {
				log.Fatalf("Error describing clusters: %v", err)
			}
		},
	}

	cmd.AddCommand(lsCmd)

	// Add flags - Output
	lsCmd.Flags().BoolVarP(&list, "list", "l", false, "Outputs Elasticache clusters in list format.")
	lsCmd.Flags().BoolVarP(&showEndpoint, "endpoint", "e", false, "Show the endpoint of the cluster")

	// Add flags - Sorting
	lsCmd.Flags().BoolP("sort-name", "n", true, "Sort by descending Elasticache cluster name.")
	lsCmd.Flags().BoolP("sort-type", "T", false, "Sort by descending Elasticache cluster type.")
	lsCmd.Flags().BoolP("sort-status", "s", false, "Sort by descending Elasticache cluster status.")
	lsCmd.Flags().BoolP("sort-engine", "E", false, "Sort by descending Elasticache cluster engine version.")
	lsCmd.Flags().SortFlags = false

	return cmd
}
