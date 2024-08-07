package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type EC2VPCEndpoint struct {
	svc               *ec2.EC2
	id                *string
	vpcId             *string
	state             *string
	ownerId           *string
	serviceName       *string
	creationTimestamp *time.Time
	tags              []*ec2.Tag
}

func init() {
	register("EC2VPCEndpoint", ListEC2VPCEndpoints)
}

func ListEC2VPCEndpoints(sess *session.Session) ([]Resource, error) {
	svc := ec2.New(sess)

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, vpc := range resp.Vpcs {
		params := &ec2.DescribeVpcEndpointsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []*string{vpc.VpcId},
				},
			},
		}

		resp, err := svc.DescribeVpcEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, vpcEndpoint := range resp.VpcEndpoints {
			resources = append(resources, &EC2VPCEndpoint{
				svc:               svc,
				id:                vpcEndpoint.VpcEndpointId,
				tags:              vpcEndpoint.Tags,
				vpcId:             vpcEndpoint.VpcId,
				state:             vpcEndpoint.State,
				ownerId:           vpcEndpoint.OwnerId,
				serviceName:       vpcEndpoint.ServiceName,
				creationTimestamp: vpcEndpoint.CreationTimestamp,
			})
		}
	}

	return resources, nil
}

func (endpoint *EC2VPCEndpoint) Remove() error {
	params := &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: []*string{endpoint.id},
	}

	_, err := endpoint.svc.DeleteVpcEndpoints(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2VPCEndpoint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", e.id)
	properties.Set("VpcId", e.vpcId)
	properties.Set("State", e.state)
	properties.Set("OwnerId", e.ownerId)
	properties.Set("ServiceName", e.serviceName)
	properties.Set("CreationTimestamp", e.creationTimestamp.Format(time.RFC3339))
	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (endpoint *EC2VPCEndpoint) String() string {
	return *endpoint.id
}
