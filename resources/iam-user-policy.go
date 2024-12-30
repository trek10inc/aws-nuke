package resources

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type IAMUserPolicy struct {
	svc        *iam.IAM
	user       iam.User
	userName   string
	policyName string
}

func init() {
	register("IAMUserPolicy", ListIAMUserPolicies)
}

func ListIAMUserPolicies(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)

	users, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, user := range users.Users {
		policies, err := svc.ListUserPolicies(&iam.ListUserPoliciesInput{
			UserName: user.UserName,
		})
		if err != nil {
			return nil, err
		}

		for _, policyName := range policies.PolicyNames {
			resources = append(resources, &IAMUserPolicy{
				svc:        svc,
				policyName: *policyName,
				userName:   *user.UserName,
				user:       *user,
			})
		}
	}

	return resources, nil
}

func (e *IAMUserPolicy) Remove() error {
	_, err := e.svc.DeleteUserPolicy(
		&iam.DeleteUserPolicyInput{
			UserName:   &e.userName,
			PolicyName: &e.policyName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserPolicy) Properties() types.Properties {
	properties := types.NewProperties().
		Set("PolicyName", e.policyName).
		Set("user:Arn", e.user.Arn).
		Set("user:UserName", e.user.UserName).
		Set("user:UserID", e.user.UserId)

	for _, tagValue := range e.user.Tags {
		properties.SetTagWithPrefix("user", tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *IAMUserPolicy) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.policyName)
}
