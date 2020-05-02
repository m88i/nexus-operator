# Nexus Operator

A Nexus OSS Kubernetes Operator based on Operators SDK

## Pre Requisites

- [`kubectl` installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes or OpenShift cluster available (minishift, minikube or crc also supported)
- Cluster admin credentials to install the Operator

## Quick Install

The installation procedure will create a Namespace named `nexus` and a Nexus 3.x server for you:

```bash
make install
```

You can then edit or customize the installation as you pleased, just run:

```bash
kubectl edit nexus
```

If you're running on Kubernetes, edit the Nexus resource to add a [valid host for the Ingress](#network-on-kubernetes-114) to work.
### Clean up

Considering that you ran the install command above, to remove the operator completely from your cluster, just run:

```bash
make uninstall
```
## Networking

There are three flavours for exposing the Nexus server deployed with the Nexus Operator: `NodePort`, `Route` (for OpenShift) and `Ingress` (for Kubernetes).

In the future, we plan to give more options to this feature like adding custom certificates.

### Use NodePort

You can expose the Nexus server via [`NodePort`](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport) by setting the following parameters in the CR:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  (...)
  networking:
    expose: true
    exposeAs: "NodePort"
    nodePort: 31031
```

It's not the recommended approach, but fits whatever Kubernetes flavour you have.

### Network on OpenShift

On OpenShift, the Nexus server can be exposed via [Routes](https://docs.openshift.com/container-platform/3.11/architecture/networking/routes.html).
Set the following parameters in the CR:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  (...)
  networking:
    expose: true
``` 

### Network on Kubernetes 1.14+

On Kubernetes, we leverage from an [`Ingress`](https://kubernetes.io/docs/concepts/services-networking/ingress/) to expose the Nexus service:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  (...)
  networking:
    expose: true
    exposeAs: "Ingress"
    host: "nexus.example.com"
```

Please note that `host` is a required parameter when exposing via `Ingress`. 
Just make sure that that the host resolves to your cluster. 

If you're running on Minikube, take a look in the article ["Set up Ingress on Minikube with the NGINX Ingress Controller"](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/) 

## Persistence

### Minikube

On Minikube the dynamic PV [creation might fail](https://github.com/kubernetes/minikube/issues/7218). If this happens in your environment, **before creating the Nexus server**, create a PV with this template: [examples/pv-minikube.yaml](examples/pv-minikube.yaml). Then give the correct permissions to the directory in Minikube VM:

```sh
$ minikube ssh
                         _             _            
            _         _ ( )           ( )           
  ___ ___  (_)  ___  (_)| |/')  _   _ | |_      __  
/' _ ` _ `\| |/' _ `\| || , <  ( ) ( )| '_`\  /'__`\
| ( ) ( ) || || ( ) || || |\`\ | (_) || |_) )(  ___/
(_) (_) (_)(_)(_) (_)(_)(_) (_)`\___/'(_,__/'`\____)

$ sudo chown 200:200 -R /data/pv0001/

$ ls -la /data/
total 8
drwxr-xr-x  3 root root 4096 Apr 26 15:42 .
drwxr-xr-x 19 root root  500 Apr 26 20:47 ..
drwxr-xr-x  2  200  200 4096 Apr 26 15:42 pv0001
```

## Red Hat Certified Images

If you have access to [Red Hat Catalog](https://access.redhat.com/containers/#/registry.connect.redhat.com/sonatype/nexus-repository-manager), you might change the flag `spec.useRedHatImage` to `true`. 
**You'll have to set your Red Hat credentials** in the namespace where Nexus is deployed to be able to pull the image. 

[In future versions](https://github.com/m88i/nexus-operator/issues/14) the Operator will handle this step for you.

## Development

### Publish to OpenShift 4.x Marketplace

1. Run `make operator-verify`

2. Grab [Quay credentials](https://github.com/operator-framework/operator-courier/#authentication) with:

```
$ export QUAY_USERNAME=youruser
$ export QUAY_PASSWORD=yourpass

$ AUTH_TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{
    "user": {
        "username": "'"${QUAY_USERNAME}"'",
        "password": "'"${QUAY_PASSWORD}"'"
    }
}' | jq -r '.token')
``` 

3. Set courier variables:

```
$ export OPERATOR_DIR=build/_output/operatorhub/
$ export QUAY_NAMESPACE=m88i # should be different for you ;)
$ export PACKAGE_NAME=nexus-operator
$ export PACKAGE_VERSION=0.2.0
$ export TOKEN=$AUTH_TOKEN
```

4. Run `operator-courier` to publish the operator application to Quay:

```
operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
```

5. Check if the application was pushed successfuly in Quay.io. Bear in mind that the application should be **public**.
 
6. Publish the operator source there:

```
$ oc create -f deploy/olm-catalog/nexus-operator/nexus-operator-operatorsource.yaml
```

7. Wait a few minutes and the Nexus Operator should be available in the Marketplace. To check it's availability, run:

```
$ oc describe operatorsource.operators.coreos.com/nexus-operator-hub -n openshift-marketplace
```
