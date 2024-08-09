package resources

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamImageBuilderWaiter struct {
	svc         *appstream.AppStream
	name        *string
	state       *string
	tags        map[string]*string
	createdTime *time.Time
}

func init() {
	register("AppStreamImageBuilderWaiter", ListAppStreamImageBuilderWaiters)
}

func ListAppStreamImageBuilderWaiters(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}

	params := &appstream.DescribeImageBuildersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeImageBuilders(params)
		if err != nil {
			return nil, err
		}

		for _, imageBuilder := range output.ImageBuilders {
			listTagsParams := &appstream.ListTagsForResourceInput{
				ResourceArn: imageBuilder.Arn,
			}
			tags, err := svc.ListTagsForResource(listTagsParams)
			if err != nil {
				return nil, err
			}

			resources = append(resources, &AppStreamImageBuilderWaiter{
				svc:         svc,
				name:        imageBuilder.Name,
				state:       imageBuilder.State,
				createdTime: imageBuilder.CreatedTime,
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

func (f *AppStreamImageBuilderWaiter) Remove() error {

	return nil
}

func (f *AppStreamImageBuilderWaiter) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("State", f.state)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}
	return properties
}

func (f *AppStreamImageBuilderWaiter) String() string {
	return *f.name
}

func (f *AppStreamImageBuilderWaiter) Filter() error {
	if *f.state == "STOPPED" {
		return fmt.Errorf("already stopped")
	} else if *f.state == "RUNNING" {
		return fmt.Errorf("already running")
	} else if *f.state == "DELETING" {
		return fmt.Errorf("already being deleted")
	}

	return nil
}
