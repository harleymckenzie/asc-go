package elasticache

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	"github.com/harleymckenzie/asc-go/pkg/shared/tableformat"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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

func NewElasticacheService(ctx context.Context, profile string) (*ElasticacheService, error) {
	var cfg aws.Config
	var err error

	if profile != "" {
		// Load the configuration for the specified profile
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	} else {
		// Use the default configuration
		cfg, err = config.LoadDefaultConfig(ctx)
	}

	if err != nil {
		return nil, err
	}

	client := elasticache.NewFromConfig(cfg)
	return &ElasticacheService{Client: client}, nil
}

func (svc *ElasticacheService) ListInstances(ctx context.Context, sortOrder []string, list bool, showEndpoint bool) error {
	selectedColumns := []string{"name", "status", "engine_version", "instance_type"}

	output, err := svc.Client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{
		ShowCacheNodeInfo: aws.Bool(showEndpoint),
	})
	if err != nil {
		log.Printf("Failed to describe instances: %v", err)
		return err
	}

	var instances []types.CacheCluster
	instances = append(instances, output.CacheClusters...)

	if showEndpoint {
		selectedColumns = append(selectedColumns, "endpoint")
	}

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

	t.SortBy(sortBy(sortOrder))
	setStyle(t, list)
	t.Render()

	return nil
}

func sortBy(sortOrder []string) []table.SortBy {
	sortBy := []table.SortBy{}

	if len(sortOrder) == 0 {
		sortOrder = []string{"name"}
	}

	for _, sortField := range sortOrder {
		sortBy = append(sortBy, table.SortBy{Name: sortField, Mode: table.Asc})
	}
	return sortBy
}

func setStyle(t table.Writer, list bool) {
	var tableStyle table.Style

	if list {
		tableStyle = table.StyleRounded
		fmt.Println("List style")
	} else {
		tableStyle = table.StyleRounded
	}
	t.SetStyle(tableStyle)
	t.Style().Format.Header = text.FormatTitle
	if list {
		t.Style().Options.DrawBorder = false
		t.Style().Options.SeparateColumns = false
		t.Style().Options.SeparateHeader = false
		t.Style().Format.Header = text.FormatUpper
	}
}
