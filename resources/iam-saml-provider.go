package resources

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type IAMSAMLProvider struct {
	svc  *iam.IAM
	arn  string
	tags []*iam.Tag
}

func init() {
	register("IAMSAMLProvider", ListIAMSAMLProvider)
}

func ListIAMSAMLProvider(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)
	params := &iam.ListSAMLProvidersInput{}
	resources := make([]Resource, 0)

	resp, err := svc.ListSAMLProviders(params)
	if err != nil {
		return nil, err
	}

	for _, out := range resp.SAMLProviderList {
		params := &iam.GetSAMLProviderInput{
			SAMLProviderArn: out.Arn,
		}
		resp, err := svc.GetSAMLProvider(params)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &IAMSAMLProvider{
			svc:  svc,
			arn:  *out.Arn,
			tags: resp.Tags,
		})
	}

	return resources, nil
}

func (e *IAMSAMLProvider) Remove() error {
	_, err := e.svc.DeleteSAMLProvider(&iam.DeleteSAMLProviderInput{
		SAMLProviderArn: &e.arn,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMSAMLProvider) String() string {
	return e.arn
}

func (e *IAMSAMLProvider) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Arn", e.arn)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
