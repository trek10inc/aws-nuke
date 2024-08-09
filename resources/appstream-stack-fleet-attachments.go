package resources

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamStackFleetAttachment struct {
	svc       *appstream.AppStream
	stackName *string
	fleetName *string
	stackTags map[string]*string
	fleetTags map[string]*string
}

func init() {
	register("AppStreamStackFleetAttachment", ListAppStreamStackFleetAttachments)
}

func ListAppStreamStackFleetAttachments(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}
	stacks := []*appstream.Stack{}
	params := &appstream.DescribeStacksInput{}

	for {
		output, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}

		stacks = append(stacks, output.Stacks...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	stackAssocParams := &appstream.ListAssociatedFleetsInput{}
	for _, stack := range stacks {

		stackAssocParams.StackName = stack.Name
		output, err := svc.ListAssociatedFleets(stackAssocParams)
		if err != nil {
			return nil, err
		}

		for _, name := range output.Names {
			describeFleetsResponse, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
				Names: []*string{name},
			})
			if err != nil {
				return nil, err
			}
			fleetArn := describeFleetsResponse.Fleets[0].Arn
			listStackTagsParams := &appstream.ListTagsForResourceInput{
				ResourceArn: stack.Arn,
			}
			stackTags, err := svc.ListTagsForResource(listStackTagsParams)
			if err != nil {
				return nil, err
			}
			listFleetTagsParams := &appstream.ListTagsForResourceInput{
				ResourceArn: fleetArn,
			}
			fleetTags, err := svc.ListTagsForResource(listFleetTagsParams)
			if err != nil {
				return nil, err
			}
			resources = append(resources, &AppStreamStackFleetAttachment{
				svc:       svc,
				stackName: stack.Name,
				fleetName: name,
				stackTags: stackTags.Tags,
				fleetTags: fleetTags.Tags,
			})
		}
	}

	return resources, nil
}

func (f *AppStreamStackFleetAttachment) Remove() error {

	_, err := f.svc.DisassociateFleet(&appstream.DisassociateFleetInput{
		StackName: f.stackName,
		FleetName: f.fleetName,
	})

	return err
}

func (f *AppStreamStackFleetAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("StackName", f.stackName)
	properties.Set("FleetName", f.fleetName)
	for key, val := range f.stackTags {
		properties.SetTagWithPrefix("stack", &key, val)
	}
	for key, val := range f.fleetTags {
		properties.SetTagWithPrefix("fleet", &key, val)
	}
	return properties
}

func (f *AppStreamStackFleetAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.stackName, *f.fleetName)
}
