package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamDirectoryConfig struct {
	svc         *appstream.AppStream
	name        *string
	createdTime *time.Time
	inUse       *bool
}

func init() {
	register("AppStreamDirectoryConfig", ListAppStreamDirectoryConfigs)
}

func ListAppStreamDirectoryConfigs(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}

	// mark which directory configs are in use by fleets
	directoryConfigsInUse := make(map[string]*bool)
	fleetParams := &appstream.DescribeFleetsInput{}
	for {
		output, err := svc.DescribeFleets(fleetParams)
		if err != nil {
			return nil, err
		}
		for _, fleet := range output.Fleets {
			inUse := true
			if fleet.DomainJoinInfo != nil && fleet.DomainJoinInfo.DirectoryName != nil {
				directoryConfigsInUse[*fleet.DomainJoinInfo.DirectoryName] = &inUse
			}
		}

		if output.NextToken == nil {
			break
		}
		fleetParams.NextToken = output.NextToken
	}

	params := &appstream.DescribeDirectoryConfigsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDirectoryConfigs(params)
		if err != nil {
			return nil, err
		}

		for _, directoryConfig := range output.DirectoryConfigs {
			inUse := false
			if directoryConfigsInUse[*directoryConfig.DirectoryName] != nil {
				inUse = *directoryConfigsInUse[*directoryConfig.DirectoryName]
			}
			resources = append(resources, &AppStreamDirectoryConfig{
				svc:         svc,
				name:        directoryConfig.DirectoryName,
				inUse:       &inUse,
				createdTime: directoryConfig.CreatedTime,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamDirectoryConfig) Remove() error {

	_, err := f.svc.DeleteDirectoryConfig(&appstream.DeleteDirectoryConfigInput{
		DirectoryName: f.name,
	})

	return err
}

func (f *AppStreamDirectoryConfig) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))
	properties.Set("InUse", f.inUse)
	return properties
}

func (f *AppStreamDirectoryConfig) String() string {
	return *f.name
}
