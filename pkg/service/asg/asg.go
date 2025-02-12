package asg

import (
	"context"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"

	"github.com/harleymckenzie/asc/pkg/shared/tableformat"
	"github.com/jedib0t/go-pretty/v6/table"
)

type AutoScalingClientAPI interface {
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

// AutoScalingService is a struct that holds the AutoScaling client.
type AutoScalingService struct {
	Client AutoScalingClientAPI
}

// ColumnDef is a definition of a column to display in the table
type columnDef struct {
	id       string
	title    string
	getValue func(*types.AutoScalingGroup) string
}

var availableColumns = []columnDef{
	{
		id:    "name",
		title: "Name",
		getValue: func(i *types.AutoScalingGroup) string {
			return aws.ToString(i.AutoScalingGroupName)
		},
	},
	{
		id:    "instances",
		title: "Instances",
		getValue: func(i *types.AutoScalingGroup) string {
			// TODO: Return count of Instances (Instance[])
			return strconv.Itoa(len(i.Instances))
		},
	},
	{
		id:    "desired_capacity",
		title: "Desired",
		getValue: func(i *types.AutoScalingGroup) string {
			return strconv.Itoa(int(*i.DesiredCapacity))
		},
	},
	{
		id:    "min_capacity",
		title: "Min",
		getValue: func(i *types.AutoScalingGroup) string {
			return strconv.Itoa(int(*i.MinSize))
		},
	},
	{
		id:    "max_capacity",
		title: "Max",
		getValue: func(i *types.AutoScalingGroup) string {
			return strconv.Itoa(int(*i.MaxSize))
		},
	},
	{
		id:    "arn",
		title: "ARN",
		getValue: func(i *types.AutoScalingGroup) string {
			return aws.ToString(i.AutoScalingGroupARN)
		},
	},
}

func NewAutoScalingService(ctx context.Context, profile string, region string) (*AutoScalingService, error) {
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

	client := autoscaling.NewFromConfig(cfg)
	return &AutoScalingService{Client: client}, nil
}

func (svc *AutoScalingService) ListAutoScalingGroups(ctx context.Context, sortOrder []string, list bool, selectedColumns []string) error {
	// Set the default sort order to name if no sort order is provided
	if len(sortOrder) == 0 {
		sortOrder = []string{"Name"}
	}

	output, err := svc.Client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return err
	}

	// At this point we have our Auto Scaling Groups
	autoScalingGroups := output.AutoScalingGroups

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
	for _, asg := range autoScalingGroups {
		// Create empty row for selected instance. Iterate through selected columns
		row := make(table.Row, len(selectedColumns))
		for i, colID := range selectedColumns {
			// Iterate through available columns
			for _, col := range availableColumns {
				// If selected column = selected available column
				if col.id == colID {
					// Add value of getValue to index value (i) in row slice
					row[i] = col.getValue(&asg)
					break
				}
			}
		}
		t.AppendRow(row)
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "Name",
			WidthMax: 40,
		},
		{
			Name:     "Instances",
			WidthMin: 8,
			WidthMax: 15,
		},
		{
			Name:     "Desired",
			WidthMin: 8,
			WidthMax: 15,
		},
		{
			Name:     "Min",
			WidthMin: 8,
			WidthMax: 15,
		},
		{
			Name:     "Max",
			WidthMin: 8,
			WidthMax: 15,
		},
	})

	t.SortBy(tableformat.SortBy(sortOrder))
	tableformat.SetStyle(t, list, false, nil)
	t.Render()
	return nil
}

func (svc *AutoScalingService) ListAutoScalingGroupInstances(ctx context.Context, autoScalingGroupName string, sortOrder []string, list bool, selectedColumns []string) error {
	// Input provided Auto Scaling Group Name into DescribeAutoScalingGroupsInput
	output, err := svc.Client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{autoScalingGroupName},
	})
	if err != nil {
		return err
	}

	var instances []types.Instance
	for _, autoScalingGroups := range output.AutoScalingGroups {
		instances = append(instances, autoScalingGroups.Instances...)
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
