# droplet-operator

Launch and manage digitalocean droplets using K8S. 

It acts as a wrapper and state manager for api calls to DO for droplet creation.

The project supports a CRD to launch a droplet.

## Instance
The instance type can be used to launch a droplet in your DigitalOcean account.

```yaml:
apiVersion: droplet.cattle.io/v1alpha1
kind: Instance
metadata:
  name: instance-sample
spec:
  # Add fields here
  name: sample
  secret: do-secret
  region: nyc3
  size: s-1vcpu-1gb
  image:
    slug: ubuntu-20-04-x64
```    

The secret `do-secret` needs to exist in the same namespace as the `Instance` object.

It only needs to contain one key `TOKEN` which contains a DigitalOcean API token
