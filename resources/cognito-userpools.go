package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/rebuy-de/aws-nuke/v2/pkg/config"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

const COGNITO_MAX_DELETE_ATTEMPT = 3

type CognitoUserPool struct {
	svc               *cognitoidentityprovider.CognitoIdentityProvider
	name              *string
	id                *string
	tags              *cognitoidentityprovider.ListTagsForResourceOutput
	maxDeleteAttempts int
	featureFlags      config.FeatureFlags
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
			logrus.Errorf("Unable to list Cognito user pools: %v", err)
			continue
		}

		for _, pool := range output.UserPools {
			userPoolDescription, err := svc.DescribeUserPool(&cognitoidentityprovider.DescribeUserPoolInput{UserPoolId: pool.Id})
			if err != nil {
				continue
			}
			tags, err := svc.ListTagsForResource(&cognitoidentityprovider.ListTagsForResourceInput{
				ResourceArn: userPoolDescription.UserPool.Arn,
			})
			if err != nil {
				continue
			}

			resources = append(resources, &CognitoUserPool{
				svc:               svc,
				name:              pool.Name,
				id:                pool.Id,
				tags:              tags,
				maxDeleteAttempts: DYNAMODB_MAX_DELETE_ATTEMPT,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *CognitoUserPool) FeatureFlags(ff config.FeatureFlags) {
	f.featureFlags = ff
}

func (f *CognitoUserPool) Remove() error {
	return f.removeWithAttempts(0)
}

func (f *CognitoUserPool) removeWithAttempts(attempt int) error {
	err := f.doRemove()
	if err == nil {
		return nil
	}
	awsErr, ok := err.(awserr.Error)
	logrus.Errorf("Cognito User Pool name=%s attempt=%d maxAttempts=%d error=%v", *f.name, attempt, f.maxDeleteAttempts, err)
	if ok && awsErr.Code() == "InvalidParameterException" && awsErr.Message() == "The user pool cannot be deleted because deletion protection is activated. Deletion protection must be inactivated first." {
		if f.featureFlags.DisableDeletionProtection.CognitoUserPool {
			logrus.Infof("Cognito User Pool name=%s attempt=%d maxAttempts=%d updating termination protection", *f.name, attempt, f.maxDeleteAttempts)
			_, err := f.svc.UpdateUserPool(&cognitoidentityprovider.UpdateUserPoolInput{
				UserPoolId:         f.id,
				DeletionProtection: aws.String("INACTIVE"),
			})
			if err != nil {
				logrus.Errorf("Cognito User Pool name=%s attempt=%d maxAttempts=%d failed to disable deletion protection: %v", *f.name, attempt, f.maxDeleteAttempts, err)
				return err
			}
			return f.removeWithAttempts(attempt + 1)
		} else {
			logrus.Errorf("Cognito User Pool name=%s attempt=%d maxAttempts=%d disable deletion protection feature flag is not enabled", *f.name, attempt, f.maxDeleteAttempts)
			return err
		}
	}
	if attempt >= f.maxDeleteAttempts {
		return err
	}
	return f.removeWithAttempts(attempt + 1)
}

func (f *CognitoUserPool) doRemove() error {
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
