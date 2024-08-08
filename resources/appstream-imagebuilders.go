package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamImageBuilder struct {
	svc         *appstream.AppStream
	name        *string
	tags        map[string]*string
	state       *string
	createdTime *time.Time
}

func init() {
	register("AppStreamImageBuilder", ListAppStreamImageBuilders)
}

func ListAppStreamImageBuilders(sess *session.Session) ([]Resource, error) {
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

			resources = append(resources, &AppStreamImageBuilder{
				svc:         svc,
				name:        imageBuilder.Name,
				tags:        tags.Tags,
				state:       imageBuilder.State,
				createdTime: imageBuilder.CreatedTime,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamImageBuilder) Remove() error {

	_, err := f.svc.DeleteImageBuilder(&appstream.DeleteImageBuilderInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamImageBuilder) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("State", f.state)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}
	return properties
}

func (f *AppStreamImageBuilder) String() string {
	return *f.name
}
