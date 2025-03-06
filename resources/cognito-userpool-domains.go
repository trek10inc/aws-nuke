package resources

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

type CognitoUserPoolDomain struct {
	svc          *cognitoidentityprovider.CognitoIdentityProvider
	name         *string
	userPoolName *string
	userPoolId   *string
	userPoolTags *cognitoidentityprovider.ListTagsForResourceOutput
}

func init() {
	register("CognitoUserPoolDomain", ListCognitoUserPoolDomains)
}

func ListCognitoUserPoolDomains(sess *session.Session) ([]Resource, error) {
	svc := cognitoidentityprovider.New(sess)

	userPools, poolErr := ListCognitoUserPools(sess)
	if poolErr != nil {
		return nil, poolErr
	}

	resources := make([]Resource, 0)
	for _, userPoolResource := range userPools {
		userPool, ok := userPoolResource.(*CognitoUserPool)
		if !ok {
			logrus.Errorf("Unable to case CognitoUserPool")
			continue
		}

		describeParams := &cognitoidentityprovider.DescribeUserPoolInput{
			UserPoolId: userPool.id,
		}
		userPoolDetails, err := svc.DescribeUserPool(describeParams)
		if err != nil {
			return nil, err
		}
		if userPoolDetails.UserPool.Domain == nil {
			// No domain on this user pool so skip
			continue
		}

		resources = append(resources, &CognitoUserPoolDomain{
			svc:          svc,
			name:         userPoolDetails.UserPool.Domain,
			userPoolName: userPool.name,
			userPoolId:   userPool.id,
			userPoolTags: userPool.tags,
		})
	}

	return resources, nil
}

func (f *CognitoUserPoolDomain) Remove() error {
	params := &cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     f.name,
		UserPoolId: f.userPoolId,
	}
	_, err := f.svc.DeleteUserPoolDomain(params)

	return err
}

func (p *CognitoUserPoolDomain) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", p.name)
	properties.Set("UserPoolName", p.userPoolName)
	for key, tag := range p.userPoolTags.Tags {
		properties.SetTag(&key, tag)
	}
	return properties
}

func (f *CognitoUserPoolDomain) String() string {
	return *f.userPoolName + " -> " + *f.name
}
