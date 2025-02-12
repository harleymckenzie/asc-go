package ssm

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/harleymckenzie/asc/pkg/shared/tableformat"
	"github.com/jedib0t/go-pretty/v6/table"
)

type SSMClientAPI interface {
	DescribeParameters(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
}

type SSMService struct {
	Client SSMClientAPI
}

type columnDef struct {
	id       string
	title    string
	getValue func(*types.ParameterMetadata) string
}

var availableColumns = []columnDef{
	{
		id:    "name",
		title: "Name",
		getValue: func(p *types.ParameterMetadata) string {
			return aws.ToString(p.Name)
		},
	},
	{
		id:    "type",
		title: "Type",
		getValue: func(p *types.ParameterMetadata) string {
			return string(p.Type)
		},
	},
}

func NewSSMService(ctx context.Context, profile string) (*SSMService, error) {
	var cfg aws.Config
	var err error

	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
	}

	if err != nil {
		return nil, err
	}

	client := ssm.NewFromConfig(cfg)
	return &SSMService{Client: client}, nil
}

func (svc *SSMService) ListParameters(ctx context.Context, selectedColumns []string) error {
	if len(selectedColumns) == 0 {
		selectedColumns = []string{"type", "name"}
	}

	params, err := svc.Client.DescribeParameters(ctx, &ssm.DescribeParametersInput{})
	if err != nil {
		return err
	}

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

	for _, param := range params.Parameters {
		row := make(table.Row, 0)
		for _, colID := range selectedColumns {
			for _, col := range availableColumns {
				if col.id == colID {
					row = append(row, col.getValue(&param))
					break
				}
			}
		}
		t.AppendRow(row)
	}

	tableformat.SetStyle(t, true, false, nil)
    fmt.Println("total", len(params.Parameters))
	t.Render()
	return nil
}
