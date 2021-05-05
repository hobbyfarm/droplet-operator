package do

import (
	"context"
	"fmt"
	"net"

	"github.com/digitalocean/godo"
	dropletv1alpha1 "github.com/ibrokethecloud/droplet-operator/pkg/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// DOClient is the wrapper to DO go client
type DOClient struct {
	*godo.Client
}

const (
	Submitted   = "submitted"
	Provisioned = "provisioned"
)

// NewClient creates a new DOClient struct that can now be used for
// performing api calls to DigitalOcean services
func NewClient(secret *corev1.Secret) (c *DOClient, err error) {
	token, ok := secret.Data["TOKEN"]
	if !ok {
		return nil, fmt.Errorf("no key TOKEN found in instance secret")
	}
	c = &DOClient{}
	c.Client = godo.NewFromToken(string(token))
	return c, err
}

// CreateDroplet uses the instance object to create the instance //
func (c *DOClient) CreateDroplet(ctx context.Context, instance *dropletv1alpha1.Instance) (status *dropletv1alpha1.InstanceStatus, err error) {
	status = &dropletv1alpha1.InstanceStatus{}
	droplet := &godo.Droplet{}
	var ok bool
	availableLabels := make(map[string]string)
	if instance.GetLabels() != nil {
		availableLabels = instance.GetLabels()
		_, ok = availableLabels["requestSubmitted"]
	}

	// only create if no URN is present
	request := instance.Spec.GenerateRequest()
	if !ok && status.InstanceID == 0 {
		droplet, _, err = c.Droplets.Create(ctx, request)
		if err != nil {
			return nil, err
		}
	}

	logrus.Info(droplet.ID)
	availableLabels["requestSubmitted"] = "true"
	instance.SetLabels(availableLabels)
	status.Status = Submitted
	status.InstanceID = droplet.ID
	return status, err
}

func (c *DOClient) FetchDetails(ctx context.Context, instance *dropletv1alpha1.Instance) (status *dropletv1alpha1.InstanceStatus, err error) {
	status = instance.Status.DeepCopy()
	if err != nil {
		return status, err
	}
	droplet, _, err := c.Droplets.Get(ctx, status.InstanceID)
	if err != nil {
		return status, err
	}

	privateIP, err := droplet.PrivateIPv4()
	if err != nil {
		return status, err
	}

	if net.ParseIP(privateIP) == nil {
		return status, fmt.Errorf("no valid private ip available")
	}
	status.PrivateIP = privateIP

	if !instance.Spec.PrivateNetworking {
		publicIP, err := droplet.PublicIPv4()
		if err != nil {
			return status, err
		}

		if net.ParseIP(publicIP) == nil {
			return status, fmt.Errorf("no valid public ip available")
		}

		status.PublicIP = publicIP
	}

	status.Status = Provisioned

	return status, nil
}

func (c *DOClient) DeleteInstance(ctx context.Context, instance *dropletv1alpha1.Instance) (err error) {
	status := instance.Status.DeepCopy()
	if err != nil {
		return err
	}
	_, err = c.Droplets.Delete(ctx, status.InstanceID)
	return err
}

func (c *DOClient) CreateKeyPair(ctx context.Context, key *dropletv1alpha1.ImportKeyPair) (status *dropletv1alpha1.ImportKeyPairStatus, err error) {
	status = key.Status.DeepCopy()
	k := godo.KeyCreateRequest{PublicKey: key.Spec.PublicKey, Name: key.Name}
	retKey, _, err := c.Keys.Create(ctx, &k)
	if err != nil {
		return status, err
	}

	status.ID = retKey.ID
	status.FingerPrint = retKey.Fingerprint
	status.Status = Provisioned
	return status, nil
}

func (c *DOClient) RemoveKeyPair(ctx context.Context, key *dropletv1alpha1.ImportKeyPair) (err error) {
	_, err = c.Keys.DeleteByID(ctx, key.Status.ID)
	return err
}
