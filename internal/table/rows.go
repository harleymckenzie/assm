package table

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/charmbracelet/bubbles/table"
)

// SortInstancesByName sorts instances by their Name tag alphabetically
func SortInstancesByName(instances []types.Instance) {
	sort.Slice(instances, func(i, j int) bool {
		nameI := GetTagValue("Name", instances[i])
		nameJ := GetTagValue("Name", instances[j])
		return nameI < nameJ
	})
}

// BuildRows builds the rows for the table
func BuildRows(instances []types.Instance) ([]table.Row, error) {
	if len(instances) == 0 {
		return []table.Row{}, nil
	}

	SortInstancesByName(instances)

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

// GetTagValue gets the tag value for the instance
func GetTagValue(name string, instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == name {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}
