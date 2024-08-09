package resources

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type AppStreamImage struct {
	svc        *appstream.AppStream
	name       *string
	visibility *string
	inUse      *bool
}

func init() {
	register("AppStreamImage", ListAppStreamImages)
}

func ListAppStreamImages(sess *session.Session) ([]Resource, error) {
	svc := appstream.New(sess)
	resources := []Resource{}

	params := &appstream.DescribeImagesInput{}

	imagesInUse := make(map[string]*bool)

	// mark which images are in use by fleets, image builders
	fleetParams := &appstream.DescribeFleetsInput{}
	for {
		output, err := svc.DescribeFleets(fleetParams)
		if err != nil {
			return nil, err
		}
		for _, fleet := range output.Fleets {
			inUse := true
			if fleet.ImageArn != nil {
				imagesInUse[*fleet.ImageArn] = &inUse
			}
		}

		if output.NextToken == nil {
			break
		}
		fleetParams.NextToken = output.NextToken
	}

	imageBuilderParams := &appstream.DescribeImageBuildersInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.DescribeImageBuilders(imageBuilderParams)
		if err != nil {
			return nil, err
		}
		for _, imageBuilder := range output.ImageBuilders {
			inUse := true
			imagesInUse[*imageBuilder.ImageArn] = &inUse
		}
		if output.NextToken == nil {
			break
		}
		imageBuilderParams.NextToken = output.NextToken
	}

	for {
		output, err := svc.DescribeImages(params)
		if err != nil {
			return nil, err
		}
		for _, image := range output.Images {
			inUse := false
			if imagesInUse[*image.Arn] != nil {
				inUse = *imagesInUse[*image.Arn]
			}
			resources = append(resources, &AppStreamImage{
				svc:        svc,
				name:       image.Name,
				visibility: image.Visibility,
				inUse:      &inUse,
			})
		}
		if output.NextToken == nil {
			break
		}
		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamImage) Remove() error {

	_, err := f.svc.DeleteImage(&appstream.DeleteImageInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamImage) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)
	properties.Set("Visibility", f.visibility)
	properties.Set("InUse", f.inUse)
	return properties
}

func (f *AppStreamImage) String() string {
	return *f.name
}

func (f *AppStreamImage) Filter() error {
	if strings.ToUpper(*f.visibility) == "PUBLIC" {
		return fmt.Errorf("cannot delete public AWS images")
	}
	return nil
}
