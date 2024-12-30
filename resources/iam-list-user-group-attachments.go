package resources

import (
	"fmt"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

type IAMUserGroupAttachment struct {
	svc       *iam.IAM
	groupName string
	userName  string
	userTags  []*iam.Tag
}

func init() {
	register("IAMUserGroupAttachment", ListIAMUserGroupAttachments)
}

func ListIAMUserGroupAttachments(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)

	resp, err := svc.ListUsers(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, user := range resp.Users {
		resp, err := svc.ListGroupsForUser(
			&iam.ListGroupsForUserInput{
				UserName: user.UserName,
			})
		if err != nil {
			return nil, err
		}

		for _, grp := range resp.Groups {
			resources = append(resources, &IAMUserGroupAttachment{
				svc:       svc,
				groupName: *grp.GroupName,
				userName:  *user.UserName,
				userTags:  user.Tags,
			})
		}
	}

	return resources, nil
}

func (e *IAMUserGroupAttachment) Remove() error {
	_, err := e.svc.RemoveUserFromGroup(
		&iam.RemoveUserFromGroupInput{
			GroupName: &e.groupName,
			UserName:  &e.userName,
		})
	if err != nil {
		return err
	}

	return nil
}

func (e *IAMUserGroupAttachment) String() string {
	return fmt.Sprintf("%s -> %s", e.userName, e.groupName)
}

func (e *IAMUserGroupAttachment) Properties() types.Properties {
	properties := types.NewProperties().
		Set("GroupName", e.groupName).
		Set("UserName", e.userName)
	for _, tag := range e.userTags {
		properties.SetTagWithPrefix("user", tag.Key, tag.Value)
	}
	return properties
}
