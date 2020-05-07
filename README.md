![Nexus Operator Integration Checks](https://github.com/m88i/nexus-operator/workflows/Nexus%20Operator%20Integration%20Checks/badge.svg)

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

### TLS/SSL

For details about TLS configuration check out our [TLS guide](https://github.com/m88i/nexus-operator/tree/master/docs/TLS.md).

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
