package resources

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamFleet struct {
	svc          *appstream.AppStream
	name         *string
	state        *string
	instanceType *string
	fleetType    *string
	createdTime  *time.Time
	tags         map[string]*string
}

func init() {
	register("AppStreamFleet", ListAppStreamFleets)
}

func ListAppStreamFleets(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}

	params := &appstream.DescribeFleetsInput{}

	for {
		output, err := svc.DescribeFleets(params)
		if err != nil {
			return nil, err
		}

		for _, fleet := range output.Fleets {
			listTagsParams := &appstream.ListTagsForResourceInput{
				ResourceArn: fleet.Arn,
			}
			tags, err := svc.ListTagsForResource(listTagsParams)
			if err != nil {
				return nil, err
			}
			resources = append(resources, &AppStreamFleet{
				svc:          svc,
				name:         fleet.Name,
				state:        fleet.State,
				instanceType: fleet.InstanceType,
				fleetType:    fleet.FleetType,
				createdTime:  fleet.CreatedTime,
				tags:         tags.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamFleet) Remove() error {

	_, err := f.svc.StopFleet(&appstream.StopFleetInput{
		Name: f.name,
	})

	if err != nil {
		return err
	}

	_, err = f.svc.DeleteFleet(&appstream.DeleteFleetInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamFleet) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("State", f.state)
	properties.Set("InstanceType", f.instanceType)
	properties.Set("FleetType", f.fleetType)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))

	for key, val := range f.tags {
		properties.SetTag(&key, val)
	}
	return properties
}

func (f *AppStreamFleet) String() string {
	return *f.name
}
