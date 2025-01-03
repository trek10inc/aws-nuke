package resources

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
	"github.com/sirupsen/logrus"
)

type IAMVirtualMFADevice struct {
	svc          *iam.IAM
	user         *iam.User
	serialNumber string
}

func init() {
	register("IAMVirtualMFADevice", ListIAMVirtualMFADevices)
}

func ListIAMVirtualMFADevices(sess *session.Session) ([]Resource, error) {
	svc := iam.New(sess)

	resp, err := svc.ListVirtualMFADevices(&iam.ListVirtualMFADevicesInput{})
	if err != nil {
		return nil, err
	}

	resources := []Resource{}
	for _, out := range resp.VirtualMFADevices {
		resources = append(resources, &IAMVirtualMFADevice{
			svc:          svc,
			user:         out.User,
			serialNumber: *out.SerialNumber,
		})
	}

	return resources, nil
}

func (v *IAMVirtualMFADevice) Filter() error {
	if strings.HasSuffix(v.serialNumber, "/root-account-mfa-device") {
		return fmt.Errorf("Cannot delete root MFA device")
	}
	if v.user != nil && strings.HasSuffix(*v.user.Arn, ":root") {
		return fmt.Errorf("Cannot delete root MFA device")
	}
	return nil
}

func (v *IAMVirtualMFADevice) Remove() error {
	if v.user != nil {
		_, err := v.svc.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
			UserName: v.user.UserName, SerialNumber: &v.serialNumber,
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if strings.Contains(aerr.Message(), fmt.Sprintf("The user with name %s cannot be found.", *v.user.UserName)) {
					logrus.Warnf("User %s not found, skipping deactivation of MFA device", *v.user.UserName)
				} else {
					return err
				}
			} else {
				return err
			}
		}
	}

	_, err := v.svc.DeleteVirtualMFADevice(&iam.DeleteVirtualMFADeviceInput{
		SerialNumber: &v.serialNumber,
	})
	return err
}

func (e *IAMVirtualMFADevice) Properties() types.Properties {
	properties := types.NewProperties().
		Set("Serial", e.serialNumber)
	if e.user != nil {
		properties.Set("UserName", e.user.UserName)
		for _, tag := range e.user.Tags {
			properties.SetTagWithPrefix("user", tag.Key, tag.Value)
		}
	}
	return properties
}

func (v *IAMVirtualMFADevice) String() string {
	return v.serialNumber
}
