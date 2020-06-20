![Nexus Operator Integration Checks](https://github.com/m88i/nexus-operator/workflows/Nexus%20Operator%20Integration%20Checks/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/m88i/nexus-operator)](https://goreportcard.com/report/github.com/m88i/nexus-operator)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/m88i/nexus-operator?label=latest)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m88i/nexus-operator)

Table of Contents
=================

   * [Nexus Operator](#nexus-operator)
      * [Pre Requisites](#pre-requisites)
      * [Quick Install](#quick-install)
         * [Openshift](#openshift)
         * [Clean up](#clean-up)
      * [Networking](#networking)
         * [Use NodePort](#use-nodeport)
         * [Network on OpenShift](#network-on-openshift)
         * [Network on Kubernetes 1.14 ](#network-on-kubernetes-114)
         * [TLS/SSL](#tlsssl)
      * [Persistence](#persistence)
         * [Minikube](#minikube)
      * [Service Account](#service-account)
      * [Control Random Admin Password Generation](#control-random-admin-password-generation)
      * [Red Hat Certified Images](#red-hat-certified-images)
      * [Image Pull Policy](#image-pull-policy)
      * [Contributing](#contributing)

# Nexus Operator

A Nexus OSS Kubernetes Operator based on Operators SDK.

You can find us at [OperatorHub](https://operatorhub.io/operator/nexus-operator-m88i) or at the ["Operators" tab in your OpenShift 4.x web console](https://docs.openshift.com/container-platform/4.4/operators/olm-adding-operators-to-cluster.html), just search for "Nexus". If you don't have access to [OLM](https://github.com/operator-framework/operator-lifecycle-manager), try installing it manually [following our quick installation guide](#quick-install).

If you have any questions please either [open an issue](https://github.com/m88i/nexus-operator/issues) or send an email to the mailing list: [nexus-operator@googlegroups.com](mailto:nexus-operator@googlegroups.com).

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

### Openshift

If you're running the Operator on Openshift (3.11 or 4.x+) and you're not using [Red Hat certified images](#red-hat-certified-images) it's also necessary to configure a [Security Context Constraints](https://docs.openshift.com/container-platform/3.11/admin_guide/manage_scc.html) (SCC) resource.

This is necessary because the Nexus image requires its container to be ran as UID 200. The use of the `restricted` default SCC in Openshift results in a failure when starting the pods, as seen in [Issue #41](https://github.com/m88i/nexus-operator/issues/41) and [Issue #51](https://github.com/m88i/nexus-operator/issues/51) (see this issue for more details on why can't the Operator handle this for you as things are now).

Valid SCC resources can be found at the `examples/` directory. You must associate the SCC with the `ServiceAccount` in use and change the SCC's `metadata.name` field (search for "<Change Me!>" in the file).

For persistent configurations:

```
$ oc apply -f examples/scc-persistent.yaml
```

For volatile configurations:

```
$ oc apply -f examples/scc-volatile.yaml
```

> **Note**: you must choose one or the other, applying both will result in using the one applied last.

Once the SCC has been created, run:

```
$ oc adm policy add-scc-to-user <SCCName> -z <ServiceAccountName>
```

This command will bind the SCC we just created with the `ServiceAccount` being used to create the Pods.

If you're [using a custom ServiceAccount](#service-account), replace "`<ServiceAccountName>`" with the name of that account. If you're not using a custom `ServiceAccount`, the operator has created a default one which has the same name as your Nexus CR, replace "`<ServiceAccountName>`" with that.

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

## Service Account

It is possible to use a custom [`ServiceAccount`](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) to perform your Deployments with the Nexus Operator via:

  - `spec.serviceAccountName` (*string*): ServiceAccountName is the name of the ServiceAccount used to run the Pods. If left blank, a default ServiceAccount is created with the same name as the Nexus CR.

**Important**: the Operator handles the creation of default resources necessary to run. If you choose to use a custom ServiceAccount be sure to also configure [`Role`](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole) and [`RoleBinding`](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) resources.

## Control Random Admin Password Generation

By default, from version 0.3.0 the Nexus Operator **does not** generate a random password for the `admin` user. This means that you can login in the server right away with the default administrator credentials (admin/admin123). **Comes in handy for development purposes, but consider changing this password right away on production environments**.

To enable random password generation, you can set the attribute `generateRandomAdminPassword` in the Nexus CR spec to `true`. Then the Nexus service will create a random password in the file system. You have to grab the password from a file inside the Nexus Server container in order to login in the web console:

```
$ kubectl exec <nexus server pod name> -- cat /nexus-data/admin.password
```

Use this password to login into the web console with the username `admin`. 

## Red Hat Certified Images

If you have access to [Red Hat Catalog](https://access.redhat.com/containers/#/registry.connect.redhat.com/sonatype/nexus-repository-manager), you might change the flag `spec.useRedHatImage` to `true`.
**You'll have to set your Red Hat credentials** in the namespace where Nexus is deployed to be able to pull the image.

[In future versions](https://github.com/m88i/nexus-operator/issues/14) the Operator will handle this step for you.

## Image Pull Policy

You can control the pods Image Pull Policy using the `spec.imagePullPolicy` field. It accepts either of the following values:

  - `Always`
  - `IfNotPresent`
  - `Never` 

If this field is set to an invalid value this configuration will be omitted, deferring to [Kubernetes default behavior](https://kubernetes.io/docs/concepts/containers/images/#updating-images), which is `Always` if the image's tag is "latest" and `IfNotPresent` otherwise.

Leaving this field blank will also result in deferring to Kubernetes default behavior.

## Contributing

Please read our [Contribution Guide](CONTRIBUTING.md).
