package resources

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rebuy-de/aws-nuke/v2/pkg/config"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

const DYNAMODB_MAX_DELETE_ATTEMPT = 3

type DynamoDBTable struct {
	svc               *dynamodb.DynamoDB
	id                string
	tags              []*dynamodb.Tag
	maxDeleteAttempts int
	featureFlags      config.FeatureFlags
}

func init() {
	register("DynamoDBTable", ListDynamoDBTables)
}

func ListDynamoDBTables(sess *session.Session) ([]Resource, error) {
	svc := dynamodb.New(sess)

	resp, err := svc.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}

	resources := make([]Resource, 0)
	for _, tableName := range resp.TableNames {
		tags, err := GetTableTags(svc, tableName)

		if err != nil {
			continue
		}

		resources = append(resources, &DynamoDBTable{
			svc:               svc,
			id:                *tableName,
			tags:              tags,
			maxDeleteAttempts: DYNAMODB_MAX_DELETE_ATTEMPT,
		})
	}

	return resources, nil
}

func (i *DynamoDBTable) FeatureFlags(ff config.FeatureFlags) {
	i.featureFlags = ff
}

func (i *DynamoDBTable) Remove() error {
	return i.removeWithAttempts(0)
}

func (i *DynamoDBTable) removeWithAttempts(attempt int) error {
	if err := i.doRemove(); err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "ValidationException" && awsErr.Message() == "Resource cannot be deleted as it is currently protected against deletion. Disable deletion protection first." {
			if i.featureFlags.DisableDeletionProtection.DynamoDBTable {
				logrus.Infof("DynamoDB Table name=%s attempt=%d maxAttempts=%d updating termination protection", i.id, attempt, i.maxDeleteAttempts)
				_, err = i.svc.UpdateTable(&dynamodb.UpdateTableInput{
					DeletionProtectionEnabled: aws.Bool(false),
					TableName:                 aws.String(i.id),
				})
				if err != nil {
					logrus.Errorf("DynamoDB Table name=%s attempt=%d maxAttempts=%d failed to disable deletion protection: %s", i.id, attempt, i.maxDeleteAttempts, err.Error())
					return err
				}
			} else {
				logrus.Warnf("DynamoDB Table name=%s attempt=%d maxAttempts=%d set feature flag to disable deletion protection", i.id, attempt, i.maxDeleteAttempts)
				return err
			}

		}
		if attempt >= i.maxDeleteAttempts {
			return errors.New("DynamoDB might not be deleted after this run.")
		} else {
			return i.removeWithAttempts(attempt + 1)
		}
	} else {
		return nil
	}
}

func (i *DynamoDBTable) doRemove() error {
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(i.id),
	}

	_, err := i.svc.DeleteTable(params)
	if err != nil {
		return err
	}

	return nil
}

func GetTableTags(svc *dynamodb.DynamoDB, tableName *string) ([]*dynamodb.Tag, error) {
	result, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(*tableName),
	})

	if err != nil {
		return make([]*dynamodb.Tag, 0), err
	}

	tags, err := svc.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: result.Table.TableArn,
	})

	if err != nil {
		return make([]*dynamodb.Tag, 0), err
	}

	return tags.Tags, nil
}

func (i *DynamoDBTable) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.id)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (i *DynamoDBTable) String() string {
	return i.id
}
