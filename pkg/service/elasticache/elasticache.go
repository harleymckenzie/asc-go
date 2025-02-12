package elasticache

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	"github.com/harleymckenzie/asc/pkg/shared/tableformat"
	"github.com/jedib0t/go-pretty/v6/table"
)

type ElasticacheClientAPI interface {
	DescribeCacheClusters(context.Context, *elasticache.DescribeCacheClustersInput, ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
}

// ElasticacheService is a struct that holds the Elasticache client.
type ElasticacheService struct {
	Client ElasticacheClientAPI
}

type columnDef struct {
	id       string
	title    string
	getValue func(*types.CacheCluster) string
}

var availableColumns = []columnDef{
	{
		id:    "name",
		title: "Cache name",
		getValue: func(i *types.CacheCluster) string {
			return aws.ToString(i.CacheClusterId)
		},
	},
	{
		id:    "status",
		title: "Status",
		getValue: func(i *types.CacheCluster) string {
			return tableformat.ResourceState(string(*i.CacheClusterStatus))
		},
	},
	{
		id:    "engine_version",
		title: "Engine version",
		getValue: func(i *types.CacheCluster) string {
			return fmt.Sprintf("%s (%s)", *i.EngineVersion, *i.Engine)
		},
	},
	{
		id:    "instance_type",
		title: "Configuration",
		getValue: func(i *types.CacheCluster) string {
			return string(*i.CacheNodeType)
		},
	},
	{
		id:    "endpoint",
		title: "Endpoint",
		getValue: func(i *types.CacheCluster) string {
			return string(*i.CacheNodes[0].Endpoint.Address)
		},
	},
}

func NewElasticacheService(ctx context.Context, profile string, region string) (*ElasticacheService, error) {
	var cfg aws.Config
	var err error

	opts := []func(*config.LoadOptions) error{}

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err = config.LoadDefaultConfig(ctx, opts...)

	if err != nil {
		return nil, err
	}

	client := elasticache.NewFromConfig(cfg)
	return &ElasticacheService{Client: client}, nil
}

func (svc *ElasticacheService) ListInstances(ctx context.Context, sortOrder []string, list bool, selectedColumns []string) error {

	// Set the default sort order to name if no sort order is provided
	if len(sortOrder) == 0 {
		sortOrder = []string{"Cache name"}
	}

	showEndpoint := false
	// Display endpoint if it is in the selected columns
	for _, col := range selectedColumns {
		if col == "endpoint" {
			showEndpoint = true
			break
		}
	}

	output, err := svc.Client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{
		ShowCacheNodeInfo: aws.Bool(showEndpoint),
	})
	if err != nil {
		return err
	}

	var instances []types.CacheCluster
	instances = append(instances, output.CacheClusters...)

	// Create the table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headerRow := make(table.Row, 0)
	for _, colID := range selectedColumns {
		for _, col := range availableColumns {
			if col.id == colID {
				headerRow = append(headerRow, col.title)
				break
			}
		}
	}
	t.AppendHeader(headerRow)

	// Start loop
	for _, instance := range instances {
		// Create empty row for selected instance. Iterate through selected columns
		row := make(table.Row, len(selectedColumns))
		for i, colID := range selectedColumns {
			// Iterate through available columns
			for _, col := range availableColumns {
				// If selected column = selected available column
				if col.id == colID {
					// Add value of getValue to index value (i) in row slice
					row[i] = col.getValue(&instance)
					break
				}
			}
		}
		t.AppendRow(row)
	}

	t.SortBy(tableformat.SortBy(sortOrder))
	tableformat.SetStyle(t, list, false, nil)
	t.Render()

	return nil
}
