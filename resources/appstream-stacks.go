package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamStack struct {
	svc         *appstream.AppStream
	name        *string
	createdTime *time.Time
	tags        map[string]*string
}

func init() {
	register("AppStreamStack", ListAppStreamStacks)
}

func ListAppStreamStacks(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}

	params := &appstream.DescribeStacksInput{}

	for {
		output, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}

		for _, stack := range output.Stacks {
			listTagsParams := &appstream.ListTagsForResourceInput{
				ResourceArn: stack.Arn,
			}
			tags, err := svc.ListTagsForResource(listTagsParams)
			if err != nil {
				return nil, err
			}
			resources = append(resources, &AppStreamStack{
				svc:         svc,
				name:        stack.Name,
				createdTime: stack.CreatedTime,
				tags:        tags.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamStack) Remove() error {

	_, err := f.svc.DeleteStack(&appstream.DeleteStackInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamStack) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))
	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}
	return properties
}

func (f *AppStreamStack) String() string {
	return *f.name
}
