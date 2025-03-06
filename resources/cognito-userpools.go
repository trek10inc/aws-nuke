package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

type CognitoUserPool struct {
	svc  *cognitoidentityprovider.CognitoIdentityProvider
	name *string
	id   *string
	tags *cognitoidentityprovider.ListTagsForResourceOutput
}

func init() {
	register("CognitoUserPool", ListCognitoUserPools)
}

func ListCognitoUserPools(sess *session.Session) ([]Resource, error) {
	svc := cognitoidentityprovider.New(sess)
	resources := []Resource{}

	params := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListUserPools(params)
		if err != nil {
			logrus.Errorf("Unable to list Cognito user pools: %w", err)
			continue
		}

		for _, pool := range output.UserPools {
			userPoolDescription, err := svc.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{UserPoolId: pool.Id})
			if err != nil {
				return nil, err
			}
			tags, err := svc.ListTagsForResource(&cognitoidentityprovider.ListTagsForResourceInput{
				ResourceArn: userPoolDescription.UserPool.Arn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &CognitoUserPool{
				svc:  svc,
				name: pool.Name,
				id:   pool.Id,
				tags: tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *CognitoUserPool) Remove() error {

	_, err := f.svc.DeleteUserPool(&cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: f.id,
	})

	return err
}

func (i *CognitoUserPool) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)
	properties.Set("Identifier", i.id)
	for key, tag := range i.tags.Tags {
		properties.SetTag(&key, tag)
	}
	return properties
}

func (f *CognitoUserPool) String() string {
	return *f.name
}
