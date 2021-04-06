/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/digitalocean/godo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InstanceSpec defines the desired state of Instance and is a wrapper around DropletCreateRequest
type InstanceSpec struct {
	Secret            *string               `json:"secret"`
	Name              string                `json:"name"`
	Region            string                `json:"region"`
	Size              string                `json:"size"`
	Image             DropletCreateImage    `json:"image"`
	SSHKeys           []DropletCreateSSHKey `json:"ssh_keys,omitempty"`
	Backups           bool                  `json:"backups,omitempty"`
	IPv6              bool                  `json:"ipv6,omitempty"`
	PrivateNetworking bool                  `json:"private_networking,omitempty"`
	Monitoring        bool                  `json:"monitoring,omitempty"`
	UserData          string                `json:"user_data,omitempty,omitempty"`
	Volumes           []DropletCreateVolume `json:"volumes,omitempty,omitempty"`
	Tags              []string              `json:"tags,omitempty"`
	VPCUUID           string                `json:"vpc_uuid,omitempty,omitempty"`
}

type DropletCreateImage struct {
	ID   int    `json:"id,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type DropletCreateSSHKey struct {
	ID          int    `json:"id,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type DropletCreateVolume struct {
	ID string `json:"id,omitempty"`
	// Deprecated: You must pass a the volume's ID when creating a Droplet.
	Name string `json:"name,omitempty"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	Status     string `json:"status"`
	InstanceID int    `json:"instanceID"`
	PrivateIP  string `json:"privateIP"`
	PublicIP   string `json:"publicIP"`
	Message    string `json:"message"`
}

// +kubebuilder:object:root=true

// Instance is the Schema for the instances API
// +kubebuilder:printcolumn:name="InstanceId",type="string",JSONPath=`.status.instanceID`
// +kubebuilder:printcolumn:name="PublicIP",type="string",JSONPath=`.status.publicIP`
// +kubebuilder:printcolumn:name="PrivateIP",type="string",JSONPath=`.status.privateIP`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

func (i *InstanceSpec) GenerateRequest() (request *godo.DropletCreateRequest) {
	request = &godo.DropletCreateRequest{}
	request.Name = i.Name
	request.Region = i.Region
	request.Size = i.Size
	request.Image = godo.DropletCreateImage{ID: i.Image.ID, Slug: i.Image.Slug}
	request.SSHKeys = sshKeys(i.SSHKeys)
	request.Backups = i.Backups
	request.IPv6 = i.IPv6
	request.PrivateNetworking = i.PrivateNetworking
	request.Monitoring = i.Monitoring
	request.UserData = i.UserData
	request.Volumes = copyVolumes(i.Volumes)
	request.Tags = i.Tags
	return request
}

func sshKeys(in []DropletCreateSSHKey) (out []godo.DropletCreateSSHKey) {
	for _, v := range in {
		tmp := godo.DropletCreateSSHKey{
			ID:          v.ID,
			Fingerprint: v.Fingerprint,
		}
		out = append(out, tmp)
	}
	return out
}

func copyVolumes(in []DropletCreateVolume) (out []godo.DropletCreateVolume) {
	for _, v := range in {
		tmpVol := godo.DropletCreateVolume{
			ID:   v.ID,
			Name: v.Name,
		}
		out = append(out, tmpVol)
	}
	return out
}
