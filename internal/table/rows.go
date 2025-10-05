package table

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/bubbles/table"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func BuildRows(instances []types.Instance) ([]table.Row, error) {
	if len(instances) == 0 {
		return []table.Row{}, nil
	}

	rows := make([]table.Row, len(instances))
	for i, instance := range instances {
		row := table.Row{
			GetTagValue("Name", instance),
			aws.ToString(instance.InstanceId),
			string(instance.State.Name),
			string(instance.InstanceType),
		}
		rows[i] = row
	}

	return rows, nil
}

func GetTagValue(name string, instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == name {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}