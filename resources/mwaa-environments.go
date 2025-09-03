package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type MWAAEnvironment struct {
	svc  *mwaa.MWAA
	name *string
	arn  *string
	tags map[string]*string
}

func init() {
	register("MWAAEnvironment", ListMWAAEnvironments)
}

func ListMWAAEnvironments(sess *session.Session) ([]Resource, error) {
	svc := mwaa.New(sess)
	resources := []Resource{}

	params := &mwaa.ListEnvironmentsInput{}

	for {
		output, err := svc.ListEnvironments(params)
		if err != nil {
			return nil, err
		}

		for _, environmentName := range output.Environments {
			env, err := svc.GetEnvironment(&mwaa.GetEnvironmentInput{
				Name: environmentName,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &MWAAEnvironment{
				svc:  svc,
				name: environmentName,
				arn:  env.Environment.Arn,
				tags: env.Environment.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *MWAAEnvironment) Remove() error {
	_, err := f.svc.DeleteEnvironment(&mwaa.DeleteEnvironmentInput{
		Name: f.name,
	})

	return err
}

func (f *MWAAEnvironment) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("ARN", f.arn)

	for tagKey, tagValue := range f.tags {
		properties.SetTag(aws.String(tagKey), tagValue)
	}

	return properties
}
