package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type RedshiftServerlessNamespace struct {
	svc       *redshiftserverless.RedshiftServerless
	namespace *redshiftserverless.Namespace
	tags      []*redshiftserverless.Tag
}

func init() {
	register("RedshiftServerlessNamespace", ListRedshiftServerlessNamespaces)
}

func ListRedshiftServerlessNamespaces(sess *session.Session) ([]Resource, error) {
	svc := redshiftserverless.New(sess)
	resources := []Resource{}

	params := &redshiftserverless.ListNamespacesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListNamespaces(params)
		if err != nil {
			return nil, err
		}

		for _, namespace := range output.Namespaces {
			tag_output, err := svc.ListTagsForResource(&redshiftserverless.ListTagsForResourceInput{
				ResourceArn: namespace.NamespaceArn,
			})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &RedshiftServerlessNamespace{
				svc:       svc,
				namespace: namespace,
				tags:      tag_output.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (n *RedshiftServerlessNamespace) Properties() types.Properties {
	properties := types.NewProperties().
		Set("CreationDate", n.namespace.CreationDate).
		Set("NamespaceName", n.namespace.NamespaceName)

	for _, tag := range n.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (n *RedshiftServerlessNamespace) Remove() error {
	_, err := n.svc.DeleteNamespace(&redshiftserverless.DeleteNamespaceInput{
		NamespaceName: n.namespace.NamespaceName,
	})

	return err
}

func (n *RedshiftServerlessNamespace) String() string {
	return *n.namespace.NamespaceName
}
