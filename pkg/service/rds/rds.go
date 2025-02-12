package rds

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/harleymckenzie/asc/pkg/shared/tableformat"
	"github.com/jedib0t/go-pretty/v6/table"
)

type RDSClientAPI interface {
	DescribeDBInstances(context.Context, *rds.DescribeDBInstancesInput, ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribeDBClusters(context.Context, *rds.DescribeDBClustersInput, ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
}

type RDSService struct {
	Client RDSClientAPI
	ctx    context.Context
}

type columnDef struct {
	id       string
	title    string
	getValue func(*types.DBInstance, []types.DBCluster) string
}

var availableColumns = []columnDef{
	{
		id:    "cluster_identifier",
		title: "Cluster Identifier",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			if i.DBClusterIdentifier != nil {
				return aws.ToString(i.DBClusterIdentifier)
			}
			return "None"
		},
	},
	{
		id:    "identifier",
		title: "Identifier",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return aws.ToString(i.DBInstanceIdentifier)
		},
	},
	{
		id:    "status",
		title: "Status",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return tableformat.ResourceState(aws.ToString(i.DBInstanceStatus))
		},
	},
	{
		id:    "engine",
		title: "Engine",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return string(*i.Engine)
		},
	},
	{
		id:    "size",
		title: "Size",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return string(*i.DBInstanceClass)
		},
	},
	{
		id:    "role",
		title: "Role",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return getDBInstanceRole(*i, clusters)
		},
	},
	{
		id:    "endpoint",
		title: "Endpoint",
		getValue: func(i *types.DBInstance, clusters []types.DBCluster) string {
			return aws.ToString(i.Endpoint.Address)
		},
	},
}

func NewRDSService(ctx context.Context, profile string, region string) (*RDSService, error) {
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

	client := rds.NewFromConfig(cfg)
	return &RDSService{Client: client, ctx: ctx}, nil
}

func (svc *RDSService) ListInstances(ctx context.Context, sortOrder []string, list bool, selectedColumns []string) error {
	// Set the default sort order to name if no sort order is provided
	if len(sortOrder) == 0 {
		sortOrder = []string{"Cluster Identifier", "Identifier"}
	}

	output, err := svc.Client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return err
	}

	var instances []types.DBInstance
	instances = append(instances, output.DBInstances...)

	clusterOutput, err := svc.Client.DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{})
	if err != nil {
		log.Printf("Failed to describe clusters: %v", err)
		return err
	}

	var clusters []types.DBCluster
	clusters = append(clusters, clusterOutput.DBClusters...)

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

	// The following loop is the same across different services, and will eventually
	// be replaced with a shared function.
	for _, instance := range instances {
		// Create empty row for selected instance. Iterate through selected columns
		row := make(table.Row, len(selectedColumns))
		for i, colID := range selectedColumns {
			// Iterate through available columns
			for _, col := range availableColumns {
				// If selected column = selected available column
				if col.id == colID {
					// Add value of getValue to index value (i) in row slice
					row[i] = col.getValue(&instance, clusters)
					break
				}
			}
		}
		t.AppendRow(row)
	}

	var (
		separateRows = true
		mergeColumn  = "Cluster Identifier"
	)

	t.SortBy(tableformat.SortBy(sortOrder))
	tableformat.SetStyle(t, list, separateRows, &mergeColumn)
	t.Render()

	return nil
}

func getDBInstanceRole(instance types.DBInstance, clusters []types.DBCluster) string {
	// If ReadReplicaSourceDBInstanceIdentifier is set, then this is a replica. If
	// if ReadReplicaDBInstanceIdentifiers is set, then this is a primary.
	if instance.ReadReplicaSourceDBInstanceIdentifier != nil {
		return "Replica"
	}

	if len(instance.ReadReplicaDBInstanceIdentifiers) > 0 {
		return "Primary"
	}

	if instance.DBClusterIdentifier == nil {
		return "None"
	}

	for _, cluster := range clusters {
		for _, member := range cluster.DBClusterMembers {
			if aws.ToString(member.DBInstanceIdentifier) == aws.ToString(instance.DBInstanceIdentifier) {
				if member.IsClusterWriter != nil && *member.IsClusterWriter {
					return "Writer"
				}
				return "Reader"
			}
		}
	}

	return "Unknown"
}
