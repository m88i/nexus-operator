![Nexus Operator Integration Checks](https://github.com/m88i/nexus-operator/workflows/Nexus%20Operator%20Integration%20Checks/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/m88i/nexus-operator)](https://goreportcard.com/report/github.com/m88i/nexus-operator)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/m88i/nexus-operator?label=latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m88i/nexus-operator)

Table of Contents
=================

<!--ts-->
   * [Table of Contents](#table-of-contents)
   * [Nexus Operator](#nexus-operator)
      * [Pre Requisites](#pre-requisites)
      * [Quick Install](#quick-install)
         * [Openshift](#openshift)
         * [Clean up](#clean-up)
      * [Automatic Updates](#automatic-updates)
         * [Successful Updates](#successful-updates)
         * [Failed Updates](#failed-updates)
      * [Custom Configuration](#custom-configuration)
      * [Networking](#networking)
         * [Use NodePort](#use-nodeport)
         * [Network on OpenShift](#network-on-openshift)
         * [Network on Kubernetes 1.14 ](#network-on-kubernetes-114)
            * [NGINX Ingress troubleshooting](#nginx-ingress-troubleshooting)
         * [Ignoring external changes to Ingress/Route resources](#ignoring-external-changes-to-ingressroute-resources)
         * [TLS/SSL](#tlsssl)
         * [Annotations and Labels](#annotations-and-labels)
      * [Persistence](#persistence)
         * [Extra volumes](#extra-volumes)
         * [Minikube](#minikube)
      * [Service Account](#service-account)
      * [Control Random Admin Password Generation](#control-random-admin-password-generation)
      * [Red Hat Certified Images](#red-hat-certified-images)
      * [Image Pull Policy](#image-pull-policy)
      * [Repositories Auto Creation](#repositories-auto-creation)
      * [Scaling](#scaling)
      * [Contributing](#contributing)


<!--te-->


# Nexus Operator

A Nexus OSS Kubernetes Operator based on the [Operator SDK](https://github.com/operator-framework/operator-sdk).

You can find us at [OperatorHub](https://operatorhub.io/operator/nexus-operator-m88i) or at the ["Operators" tab in your OpenShift 4.x web console](https://docs.openshift.com/container-platform/4.4/operators/olm-adding-operators-to-cluster.html), just search for "Nexus". If you don't have access to [OLM](https://github.com/operator-framework/operator-lifecycle-manager), try installing it manually [following our quick installation guide](#quick-install).

If you have any questions please either [open an issue](https://github.com/m88i/nexus-operator/issues) or send an email to the mailing list: [nexus-operator@googlegroups.com](mailto:nexus-operator@googlegroups.com).

## Pre Requisites

- [`kubectl` installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes (1.16+) or OpenShift (4.5+) cluster available (minikube or crc also supported)
- Cluster admin credentials to install the Operator

> Note: since version 0.6.0 we do not support OpenShift 3.11 or Kubernetes 1.11 anymore.
> If you need to install in these clusters, please use version [0.5.0](https://github.com/m88i/nexus-operator/releases/tag/v0.5.0) instead.

## Quick Install

The installation procedure will create a Namespace named `nexus-operator-system` and will install every resources needed for the operator to run:

```bash
# requires python and kubectl
bash <(curl -s https://github.com/m88i/nexus-operator/blob/main/hack/install.sh)
```

Alternatively, you can manually elect a [released version](https://github.com/m88i/nexus-operator/releases):

```bash
VERSION=<version from GitHub releases page>

kubectl apply -f https://github.com/m88i/nexus-operator/releases/download/${VERSION}/nexus-operator.yaml
```

You can choose any flavors of Nexus 3.x server from our [`examples`](examples) directory and apply the YAML in any namespace in your cluster.
Use these examples as a starting point to customize the server to meet your requirements.

### Openshift

If you're running the Operator on Openshift (4.5+) and **you're not using Red Hat image with persistence enabled**, that's anything other than `spec.useRedHatImage: true` and `spec.persistence.persistent: true`,
it's also necessary to configure a [Security Context Constraints](https://docs.openshift.com/container-platform/3.11/admin_guide/manage_scc.html) (SCC) resource.

This is necessary because the Nexus image requires its container to be ran as UID 200. 
The use of the `restricted` default SCC in Openshift results in a failure when starting the pods, as seen in [Issue #41](https://github.com/m88i/nexus-operator/issues/41) and [Issue #51](https://github.com/m88i/nexus-operator/issues/51) (see this issue for more details on why can't the Operator handle this for you as things are now).

Valid SCC resources can be found at the `examples/` directory. You must associate the SCC with the `ServiceAccount` in use.

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
$ oc adm policy add-scc-to-user allow-nexus-userid-200 -z <ServiceAccountName>
```

This command will bind the SCC we just created with the `ServiceAccount` being used to create the Pods.

If you're [using a custom ServiceAccount](#service-account), replace "`<ServiceAccountName>`" with the name of that account. 
If you're not using a custom `ServiceAccount`, the operator has created a default one which has the same name as your Nexus CR, replace "`<ServiceAccountName>`" with that.

### Clean up

Considering that you ran the install command above, to remove the operator completely from your cluster, just run:

```bash
make uninstall
```

## Automatic Updates

The Nexus Operator is capable of conducting automatic updates within a minor (the `y` in `x.y.z`) when using the community default image (`docker.io/sonatype/nexus3`). In the future Red Hat images will also be supported by this feature.
> **Note**: custom images will not be supported as there is no guarantee that they follow [semantic versioning](https://semver.org/) and as such, updates within the same minor may be disruptive.

Two fields within the Nexus CR control this behavior:

  - `spec.automaticUpdate.disabled` (*boolean*): Whether the Operator should perform automatic updates. Defaults to `false` (auto updates are enabled). Is set to `false` if `spec.image` is not empty and is different from the default community image.
  - `spec.automaticUpdate.minorVersion` (*integer*): The Nexus image minor version the deployment should stay in. If left blank and automatic updates are enabled the latest minor is set.

> **Note**: if you wish to set a specific tag when using the default community image you must first disable automatic updates.

> **Important**: a change of minors will *not* be monitored or acted upon as an automatic update. Changing the minor is a manual process initiated by the human operator and as such must be monitored by the human operator.
 
The state of ongoing updates is written to `status.updateConditions`, which can be easily accessed with `kubectl`:

```
$ kubectl describe nexus
# (output omitted)
  Update Conditions:
    Starting automatic update from 3.26.0 to 3.26.1
    Successfully updated from 3.26.0 to 3.26.1
Events:
  Type    Reason         Age   From    Message
  ----    ------         ----  ----    -------
  Normal  UpdateSuccess  59s   nexus3  Successfully updated to 3.26.1
```

> **Note**: do *not* modify these conditions manually, the Operator reconstructs the update state from these.

### Successful Updates

Once an update finishes successfully, an [Event](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#event-v1-core) is raised. You may view the events from a particular Nexus CR by describing it:

```
$ kubectl describe <Nexus CR>
```

Or you may query all events:

```
$ kubectl get events
```

A successful update event looks like:

```yaml
apiVersion: v1
count: 1
eventTime: null
firstTimestamp: "2020-08-26T13:56:16Z"
involvedObject:
  apiVersion: apps.m88i.io/v1alpha1
  kind: Nexus
  name: nexus3
  namespace: update
  resourceVersion: "66087"
  uid: f017e60f-21b5-4b14-b67c-341e029afae3
kind: Event
lastTimestamp: "2020-08-26T13:56:16Z"
message: Successfully updated to 3.26.1
# (output omitted)
reason: UpdateSuccess
reportingComponent: ""
reportingInstance: ""
source:
  component: nexus3
type: Normal
```

```
$ kubectl get events         
LAST SEEN   TYPE      REASON              OBJECT                         MESSAGE
12m         Normal    UpdateSuccess       nexus/nexus3                   Successfully updated to 3.26.1
# (output omitted)
```

### Failed Updates

When an update fails, since the Deployments produced by the Operator use a [Rolling Deployment Strategy](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment) there is no disruption and the previous version is still available. 
The Operator will then:
 
   1. disable automatic updates
   2. set `spec.image` to the version that was set before the update began
   3. raise a failure event

A failed update event looks like:

```yaml
apiVersion: v1
count: 1
eventTime: null
firstTimestamp: "2020-08-21T18:29:11Z"
involvedObject:
  apiVersion: apps.m88i.io/v1alpha1
  kind: Nexus
  name: nexus3
  namespace: update
  resourceVersion: "51602"
  uid: 2e9ef49a-7d37-4c96-bfae-0642a9487c95
kind: Event
lastTimestamp: "2020-08-21T18:29:11Z"
message: Failed to update to 3.26.1. Human intervention may be required
# (output omitted)
reason: UpdateFailed
reportingComponent: ""
reportingInstance: ""
source:
  component: nexus3
type: Warning
```

```
$ kubectl get events         
  LAST SEEN   TYPE      REASON              OBJECT                         MESSAGE
  9m45s       Warning   UpdateFailed        nexus/nexus3                   Failed to update to 3.26.1. Human intervention may be required
# (output omitted)
```
## Custom Configuration

Starting on version 0.6.0, the operator now mounts a [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) with
the contents of the [`nexus.properties`](https://help.sonatype.com/repomanager3/installation/configuring-the-runtime-environment) file
in the path `$NEXUS_DATA/etc/nexus.properties`.

The Nexus Operator mount this file with the contents of the field `Spec.Properties` using [the Java properties format](https://docs.oracle.com/javase/8/docs/api/java/util/Properties.html#load-java.io.Reader-). 
If you change this field, the operator will deploy a new pod _immediately_ to reflect the changes applied in the `ConfigMap`.

**Don't update** the managed `ConfigMap` directly, otherwise the operator will replace its contents with `Spec.Properties` field.
Always use the Nexus CR as the only source of truth. See this [example](examples/nexus3-centos-no-volume-custom-properties.yaml) to
learn how to properly set your properties directly in the CR.

> **Beware!** Since we don't support HA yet, the server will be unavailable until the next pod comes up. Try to update the configuration only 
> when you can afford to have the server unavailable.

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

#### NGINX Ingress troubleshooting

If you've deployed the [NGINX Ingress controller](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/), you might see [`413 ERROR - Entity too large`](https://github.com/kubernetes/ingress-nginx/issues/4825) in uploading the artifacts to the Nexus server.
 
You would need to enter the maximum size allowed for the data packet in the `configMap` for the controller.

If you've deployed the Ingress controller in Minikube it'll be available in the `kube-system` namespace

```
$ kubectl get deploy -n kube-system
                                                              
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
coredns                    1/1     1            1           47h
ingress-nginx-controller   1/1     1            1           47h
``` 
For checking out the name of the `configMap` you can run:

```shell-script
$ kubectl get deploy/ingress-nginx-controller -o yaml -n kube-system | grep "\--configmap" 

- --configmap=$(POD_NAMESPACE)/nginx-load-balancer-conf
```

Now you would need to edit the config map:

`$ kubectl edit configmaps nginx-load-balancer-conf -n kube-system `

In the root of the opened yaml file add:

```yaml
data:
  proxy-body-size: 10m
```

**Note**: If you want to have no limit for the data packet you can specify the `proxy-body-size: 0m`

### Ignoring external changes to Ingress/Route resources

Route and Ingress resources are highly configurable, and often the need to change them arises. For example, further
configuration can be performed by webhooks, but these changes get undone by the Operator as soon as it detects them.

Starting at version 0.6.0 you may specify that the Operator should ignore external changes made to Ingress and Route
resources. This is controlled by the `spec.networking.ignoreUpdates` boolean field in the Nexus resource. It defaults to
`false`, meaning the Operator will change the Ingress/Route specification to match its state as defined by this
resource. Set to `true` in order to prevent the Operator from undoing external changes in the resources' configuration.

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  networking:
    ignoreUpdates: true
```

### TLS/SSL

For details about TLS configuration check out
our [TLS guide](https://github.com/m88i/nexus-operator/tree/main/docs/TLS.md).

### Annotations and Labels

You may provide custom labels and annotations to Route/Ingress resources by setting them
on  `.spec.networking.annotations` and `.spec.networking.labels`. For example:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  networking:
    annotations:
      my-cool-annotation: "even-cooler-value"
      my-other-cool-annotation: "not-as-cool-value"
    labels:
      my-cool-label: "even-cooler-value"
```

## Persistence

### Extra volumes

Starting at version 0.6.0 you may specify extra volumes to be mounted at the pod running Nexus, which comes in handy for
migrating existing blob stores, for example. These volumes are controlled by the `spec.persistence.extraVolumes` field.

For example, if you wanted to mount an AWS EBS volume, some PVC of yours and an EmptyDir volume:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  persistence:
    extraVolumes:
      - name: "my-cool-ebs-vol"
        mountPath: "/path/for/AWS-EBS/"
        # This AWS EBS volume must already exist.
        awsElasticBlockStore:
          volumeID: "<volume id>"
          fsType: ext4
      - name: "my-cool-claim-vol"
        mountPath: "/path/for/persistent-vol-claim/"
        # This PVC must exist on the same namespace
        persistentVolumeClaim:
          claimName: "my-cool-claim"
      - name: "my-cool-empty-dir-vol"
        mountPath: "/path/for/emptyDir/"
        emptyDir: { }
```

Each item of this `extraVolumes` array provides:

- `mountPath`: a string representing the path at which this volume should be mounted
- a Kubernetes `Volume` specification

For more information about Kubernetes Volumes refer to
their [documentation](https://kubernetes.io/docs/concepts/storage/volumes/)
and each specific plugin documentation. For additional details about Persistent Volumes and using Claims as volumes
refer to the [documentation](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#claims-as-volumes).

> **Important**: updating the `spec.persistence.extraVolumes` field may lead to temporary unavailability while the new
> deployment with the new volume configuration rolls out.

### Minikube

On Minikube the dynamic PV [creation might fail](https://github.com/kubernetes/minikube/issues/7218). If this happens in
your environment, **before creating the Nexus server**, create a PV with this
template: [examples/pv-minikube.yaml](examples/pv-minikube.yaml). Then give the correct permissions to the directory in
Minikube VM:

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

## Repositories Auto Creation

From 0.3.0 version, the Operator will try to create an administrator user to be used on internal operations, such as creating community Maven repositories.

The default Nexus user `admin` is used to create the `nexus-operator` user, whose credentials are then stored in a secret with the same name as the Nexus CR.

It's possible to disable the operator user creation by setting `spec.serverOperatons.disableOperatorUserCreation` to `true`. In this case, the `admin` user will be used instead. This configuration is **not recommended**, since you can track all the operations, change the operator user permissions and enable or disable it if you need. By disabling the operator user creation, the Operator will use the default `admin` credentials to perform all server operations, which will fail if you change the default credentials (something that must be done when aiming for a secure environment).

The Operator also will create three Maven repositories by default:

1. [Apache](https://repo.maven.apache.org/maven2/)
2. [JBoss](https://repo.maven.apache.org/maven2/)
3. [Red Hat](https://maven.repository.redhat.com/ga/)

All of these repositories will be also added to the `maven-public` group. This group will gather the vast majority of jars needed by the most common use cases out there. If you won't need them, just disable this behavior by setting the attribute `spec.serverOperatons.disableRepositoryCreation` to `true` in the Nexus CR. 

All of these operations are disabled if the attribute `spec.generateRandomAdminPassword` is set to `true`, since default credentials are needed to create the `nexus-operator` user. You can safely change the default credentials after this user has been created.

## Scaling

For now, the Nexus Operator won't accept a number higher than `1` to the `spec.replicas` attribute.
This is because the Nexus server can't share its mounted persistent volume with other pods. See #191 for more details.

Horizontal scaling will only work once we add [HA support](https://help.sonatype.com/repomanager3/high-availability) to the operator (see #61). 
If you need to scale the server, you should take the vertical approach and increase the numbers of resource limits used
by the Nexus server. For example:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  replicas: 1
  # Set the resources requests and limits for Nexus pods. See: https://help.sonatype.com/repomanager3/system-requirements
  resources:
    limits:
      cpu: "4"
      memory: "8Gi"
    requests:
      cpu: "1"
      memory: "2Gi"
  persistence:
    persistent: true
    volumeSize: 10Gi
```

We are working to support HA in the future.

## Contributing

Please read our [Contribution Guide](CONTRIBUTING.md).
