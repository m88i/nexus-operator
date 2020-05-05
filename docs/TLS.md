# Nexus Operator TLS guide

While TLS termination itself is automatically supported by the cluster in both Openshift and Kubernetes via self-signed certificates, the Nexus Operator defines a new resource "Nexus" that offers the following TLS configurations:

  - `spec.networking.tls.mandatory` (*boolean*): When exposing via Route, set to `true` to only allow encrypted traffic using TLS (disables HTTP in favor HTTPS). Defaults to false.
  - `spec.networking.tls.secretName` (*string*): When exposing via Ingress, inform the name of the TLS secret containing certificate and private key for TLS encryption. It must be present in the same namespace as the Operator.

This configuration is meant for testing purposes only and does not seek to address all requirements faced in a production environment. If more complex configuration is required, set `spec.networking.expose` to `false` in order to configure the desired network resource (e.g., Ingress) directly.

## Constraints

When using an Openshift [Route](https://docs.openshift.com/container-platform/4.3/networking/routes/route-configuration.html) only `spec.networking.tls.mandatory` is available. This is due to Routes containing the TLS information directly instead of sourcing it via a Secret.

When using a Kubernetes [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) only `spec.networking.tls.secretName` is available. This is due to HTTP->HTTPS redirection being managed by the [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/), not by the Ingress itself.

> **Note**: There is a gap between TLS features supported by various Ingress controllers. Please refer to documentation on NGINX, GCE, or any other platform specific Ingress controller to understand how TLS works in your environment

When using NodePort none of them are available.

`spec.networking.tls.secretName` must point to a valid "kubernetes.io/tls" Secret, such as:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: testsecret-tls
  namespace: default
data:
  tls.crt: <base64 encoded cert>
  tls.key: <base64 encoded key>
type: kubernetes.io/tls
```

## Examples

In this section we'll have a look at some examples on how to use these features. The CRs in use can be found at `examples/`. It is assumed you have a cluster with a functioning Nexus Operator deployment (if you don't, check out our [README quick install guide](https://github.com/m88i/nexus-operator#quick-install)).

### Kubernetes Ingress

In this example we'll use the `spec.networking.tls.secretName` directive:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: testsecret-tls
  namespace: nexus
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURtakNDQW9LZ0F3SUJBZ0lKQUlQMFdHVEMxMnA5TUEwR0NTcUdTSWIzRFFFQkN3VUFNR0l4Q3pBSkJnTlYKQkFZVEFsaFlNUlV3RXdZRFZRUUhEQXhFWldaaGRXeDBJRU5wZEhreEhEQWFCZ05WQkFvTUUwUmxabUYxYkhRZwpRMjl0Y0dGdWVTQk1kR1F4SGpBY0JnTlZCQU1NRlc1bGVIVnpMWFJzY3k1bGVHRnRjR3hsTG1OdmJUQWVGdzB5Ck1EQTFNREl5TVRRd016ZGFGdzB5TVRBMU1ESXlNVFF3TXpkYU1HSXhDekFKQmdOVkJBWVRBbGhZTVJVd0V3WUQKVlFRSERBeEVaV1poZFd4MElFTnBkSGt4SERBYUJnTlZCQW9NRTBSbFptRjFiSFFnUTI5dGNHRnVlU0JNZEdReApIakFjQmdOVkJBTU1GVzVsZUhWekxYUnNjeTVsZUdGdGNHeGxMbU52YlRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCCkJRQURnZ0VQQURDQ0FRb0NnZ0VCQU13RzF2ODJ0VjJnUHhjcmFGWXk3cmxLcHJVQnRVUHpuQ2Y3YzdlTUVsTEkKczRHWUJ6Z3Jyb3YrTnM3OXFPaHFsWmk3WEtjajYzT2RNYzg3RmZ3VHJZQ3NNUFZvbU9XdVgzTEJDVVZOSmJIOAowaUZwRmdlNGk0S05YVStWUTBoVmUyN2dkNzRGOWpBcUlYVWNPUUVISy9Ec0gwWnQyZW5rSVNLS2owdnlKYXJtCmIyeUd3MWRZUXozdVY1ak5XWEhOTjhEQTQzbFJBRlJzUW5kVXNTNzBnS0V1Qk8ycHkrYURoYU9rWUhxZ0pqVXkKTmVOcmxoYXltL1YyZGt6TkFVbnhFTFlQUkhicEJ2dnBubmkzQTdRcEtYQzJpOUFmTlYzWDVCajlkUVJIbEUxdQpWWWNMTmE2aXh0WGtjM1h5cTlqMmhUSElLdHJOUGxoMEMySDRucG5xK3o4Q0F3RUFBYU5UTUZFd0hRWURWUjBPCkJCWUVGRUJzV0xMTFV0K3ZOSlppLzU3dExlMWZEclhCTUI4R0ExVWRJd1FZTUJhQUZFQnNXTExMVXQrdk5KWmkKLzU3dExlMWZEclhCTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUN0ZQo2ekd5WHc2d3I4cFZvL0hMK25lMVNYc0lIRjVqWDRuRy9WOWdQNXlyMCtraDFuNkR6RzZYMTJuTnhJM08vZFJRClhuMzZWZk9TbFpjUFE3Rk1WeVRxdVNVTHJscFd2MnllTmFYTDRMMGMvbC80dWFKblphTytjZnk4b0lNeTJjaGQKWG1tS2l0eWFhVDNPZ1RGdHpjem1vWmZKbmJQM04zRjRxZWpta2hCOGRXb2RuOVVkNEdIOG9hamhZYkNvdDRoYQpCZ1BLbHEydFcvd2JUbUxHcGRJb3NjYWdhR1hVMG1wbTFUNExWL01oYWxIRmMvbXNtTGs3RHBzbkZVKytQeHFXCnFab0tDVnZYY05LMGpCSnNSVS9aaTFXS2o3dE1qRG10UE8wSFJCd0JJWnAwU1hDMnN3M2dnbWVOTmZ0TTZlZmQKVkU3L2JsTUFtbjRzRnpYWFdndz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2d0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktrd2dnU2xBZ0VBQW9JQkFRRE1CdGIvTnJWZG9EOFgKSzJoV011NjVTcWExQWJWRDg1d24rM08zakJKU3lMT0JtQWM0SzY2TC9qYk8vYWpvYXBXWXUxeW5JK3R6blRIUApPeFg4RTYyQXJERDFhSmpscmw5eXdRbEZUU1d4L05JaGFSWUh1SXVDalYxUGxVTklWWHR1NEhlK0JmWXdLaUYxCkhEa0JCeXZ3N0I5R2JkbnA1Q0VpaW85TDhpV3E1bTlzaHNOWFdFTTk3bGVZelZseHpUZkF3T041VVFCVWJFSjMKVkxFdTlJQ2hMZ1R0cWN2bWc0V2pwR0I2b0NZMU1qWGphNVlXc3B2MWRuWk16UUZKOFJDMkQwUjI2UWI3Nlo1NAp0d08wS1Nsd3RvdlFIelZkMStRWS9YVUVSNVJOYmxXSEN6V3Vvc2JWNUhOMThxdlk5b1V4eUNyYXpUNVlkQXRoCitKNlo2dnMvQWdNQkFBRUNnZ0VCQUtYcktneUNRUnZTcFh3Q0lPZDRwQXFyV3NiNGpLQ21DK2UzNVhMdXhqZnMKWDQ2VVE4cTZPMGc5Zy9Udzg5dU9VZm5TNUpaSDc4SWxvOHdURU4zZVlzcXhQVjlWb0lHb1BKYmx4TlJHMk5EVgorVXlTV1FnT1crWGZjSEszdisrakZLNG1mVHBiUHNvTWVRanoycWVyWFdDZnBROXhaalYvQURORzJ5RzMya29QCmg3VHFGb1ZWMnRuYk83MnNHbkprNmlZZzlsV2dweE1FOCtDYTdmT3B3aXdQK2F6LzBrNTM5QXJHbVNzQnN1MmoKNis5Sm5DK2JrY0Z0M0haK3FNUWdOc1QrNUdZQnRCOEFqWlpRZlJTczhTSWdybkZRcjJIeDNIcnRMcUdKb1NtOQpmcXdYWjRLMTFzT0lNMHhtaXBuYTk2aXM0TjFRRVBLM0F5RzNSVW1uTG9FQ2dZRUE3VFp5b0l3WTJFc2h2YWprCkFLeTBnSEVqY25EWW9ydHUvK0VtTkNjandFc2xScC9nN1dIdDNwTXl6MmlCR0FPaVlPQXdTb2lvSzBvbUR5MGwKNTFFYTZOd3NPOG1YMzY1b3U5VGRLL2VzN1ZERzdSZStiZG94OERMNkNSUVlYNU9oTUM5bVNXbVIzaDJML2hzQgpkU2preVArV3NkS2FwcFZsNGdpdVhhVkZkbHNDZ1lFQTNDK0pseWRDVmxwZWNXUjJCdTJHVTRCd3puU2xrazA3CkJIWkxBaEk3QXhnVVpjaGx6dGtURU9ENVJZUUxGVmdjQzFVMVB0Um16OVFnRW1qcTk0YzBHSFNRL0Z0Y2psdzgKcDlCYnlVTlF5Tms0eURZeElHdmsrdndHSkhob0VKYWIxNWpkS2FDY2xYRkN3TG9LTy8wZ2U5b1kxWEF0SlFGLwp6Tm9NQzUxYmkrMENnWUVBMWx2b2srcG1IVGN2dzJzV3R3RmFqK010akNJcnNrcThrT0NPSEh6dUVyd1ZjRS9UCis5QU9HNTliUUJDSThBR0F3QmgzcXpMNC9UMmhUUCtZakFNLzFRUDV1UUdBaS9MWTFEd1VyY0hBZENnVnkzVTAKY2FMR2svQU5BUjAydFUvOFRrWFhJaW9UVmV2UGNRNUliKzVIYU5lRy95UjQrbVp4VGlSWUpXblBicGNDZ1lBcAo0Q0NhenF5Zkg3QzJnQlN5WEpvZloxNE01Y0pYZ0xpb3NKYXpYaVE3QW1sZXNpNHFtTDQzaDVIZzFxd0U4eXppCk91SlZnSy9NOXRyaXBYR0tnZncyYW5Ub2liZWdtNG90b0VMVWxDalpDZmJ5bk52YS9xb2QwYkNaWHd6cm1ya28KMTdtNElRT21xRk81czZnZW9KVjgrSTJnaWlVTDFLMHBtSTZSNXV0eVFRS0JnUUNIdDBvb0NyaHdnRStiVEkwZApNbmxDNFZzREhxc3NrMElzZDZRaHExV1FwUUk2WjNOdXZDLy9hbzZjeHZsUGRLVkt5eXZBc3RyWUx5KzM2TDJBCkpmaldSRVplRjVENGl1K2ZsMWY4Uzl1MnBzM01pN3IrL1VOWjlxNGp4TVhDdy9lbkpUYkNwRSt6ZlQ2OEZJOXYKaThXcWt1OUxWbk1SbkN3L251Y3NCcUh0d0E9PQotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg==
type: kubernetes.io/tls
---
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  # Number of Nexus pod replicas (can't be increased after creation)
  replicas: 1
  # Here you can specify the image version to fulfill your needs. Defaults to docker.io/sonatype/nexus3:latest if useRedHatImage is set to false
  #image: "docker.io/sonatype/nexus3:latest"
  # let's use the centOS image since we do not have access to Red Hat Catalog
  useRedHatImage: false
  # Set the resources requests and limits for Nexus pods. See: https://help.sonatype.com/repomanager3/system-requirements
  resources:
    limits:
      cpu: "2"
      memory: "2Gi"
    requests:
      cpu: "1"
      memory: "2Gi"
  # Data persistence details
  persistence:
    # Should we persist Nexus data? (turn this to false only if you're evaluating this resource)
    persistent: false
  # details regarding networking
  networking:
    # expose please
    expose: true
    # How do you want to expose the Nexus server? In this case, we're using Ingress. Only available on Kubernetes
    exposeAs: "Ingress"
    # The host is required when using Ingress, needs to resolve to your cluster
    host: "nexus-tls.example.com"
    # Configuration related to TLS
    tls:
      # Use the TLS secret we defined above
      secretName: "testsecret-tls"
```

Note that a TLS Secret is also defined in the first few lines. It contains a self-signed x.509 certificate and its private key encoded in base64.

> **Note**: Base64 encoding is not an encryption method and is considered the same as plain text. Access control to the Secret is still necessary.

**Important**: the certificate used when generating the secret must have the Common Name field set to exactly the same hostname your Ingress is using or the Ingress Controller may reject your certificate.

In this example the secret CR has already been defined in the same file as the nexus CR for simplicity, but you may also create your secret from existing files using `kubectl`:

```bash
$ kubectl create secret tls secret-name --cert /path/to/certificate --key /path/to/private/key
```

Let's create these resources:

```
$ kubectl create -n nexus -f examples/nexus3-centos-tls-ingress.yaml
secret/testsecret-tls created
nexus.apps.m88i.io/nexus3 created
```

Now we should have an Ingress with TLS configuration:

```
$ kubectl get -n nexus ingress/nexus3
NAME     CLASS    HOSTS                   ADDRESS         PORTS     AGE
nexus3   <none>   nexus-tls.example.com   192.168.39.29   80, 443   84s
$ kubectl get -n nexus ingress/nexus3 -o yaml
apiVersion: extensions/v1beta1
kind: Ingress
# (output omitted)
spec:
  rules:
  - host: nexus-tls.example.com
    http:
      paths:
      - backend:
          serviceName: nexus3
          servicePort: 8081
        path: /
        pathType: ImplementationSpecific
  tls:
  - hosts:
    - nexus-tls.example.com
    secretName: testsecret-tls
status:
  loadBalancer:
    ingress:
    - ip: 192.168.39.29
```

There it is:

```yaml
  tls:
  - hosts:
    - nexus-tls.example.com
    secretName: testsecret-tls
```

Using openssl's `s_client` module we can actually connect and inspect the TLS handshake outcome (note the "-servername" flag which should enable the use of [SNI](https://en.wikipedia.org/wiki/Server_Name_Indication)):

```
openssl s_client -connect 192.168.39.29:443 -servername nexus-tls.example.com < /dev/null
CONNECTED(00000003)
depth=0 C = XX, L = Default City, O = Default Company Ltd, CN = nexus-tls.example.com
verify error:num=18:self signed certificate
verify return:1
depth=0 C = XX, L = Default City, O = Default Company Ltd, CN = nexus-tls.example.com
verify return:1
---
Certificate chain
 0 s:/C=XX/L=Default City/O=Default Company Ltd/CN=nexus-tls.example.com
   i:/C=XX/L=Default City/O=Default Company Ltd/CN=nexus-tls.example.com
---
Server certificate
-----BEGIN CERTIFICATE-----
MIIDmjCCAoKgAwIBAgIJAIP0WGTC12p9MA0GCSqGSIb3DQEBCwUAMGIxCzAJBgNV
BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQg
Q29tcGFueSBMdGQxHjAcBgNVBAMMFW5leHVzLXRscy5leGFtcGxlLmNvbTAeFw0y
MDA1MDIyMTQwMzdaFw0yMTA1MDIyMTQwMzdaMGIxCzAJBgNVBAYTAlhYMRUwEwYD
VQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQx
HjAcBgNVBAMMFW5leHVzLXRscy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAMwG1v82tV2gPxcraFYy7rlKprUBtUPznCf7c7eMElLI
s4GYBzgrrov+Ns79qOhqlZi7XKcj63OdMc87FfwTrYCsMPVomOWuX3LBCUVNJbH8
0iFpFge4i4KNXU+VQ0hVe27gd74F9jAqIXUcOQEHK/DsH0Zt2enkISKKj0vyJarm
b2yGw1dYQz3uV5jNWXHNN8DA43lRAFRsQndUsS70gKEuBO2py+aDhaOkYHqgJjUy
NeNrlhaym/V2dkzNAUnxELYPRHbpBvvpnni3A7QpKXC2i9AfNV3X5Bj9dQRHlE1u
VYcLNa6ixtXkc3Xyq9j2hTHIKtrNPlh0C2H4npnq+z8CAwEAAaNTMFEwHQYDVR0O
BBYEFEBsWLLLUt+vNJZi/57tLe1fDrXBMB8GA1UdIwQYMBaAFEBsWLLLUt+vNJZi
/57tLe1fDrXBMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBACte
6zGyXw6wr8pVo/HL+ne1SXsIHF5jX4nG/V9gP5yr0+kh1n6DzG6X12nNxI3O/dRQ
Xn36VfOSlZcPQ7FMVyTquSULrlpWv2yeNaXL4L0c/l/4uaJnZaO+cfy8oIMy2chd
XmmKityaaT3OgTFtzczmoZfJnbP3N3F4qejmkhB8dWodn9Ud4GH8oajhYbCot4ha
BgPKlq2tW/wbTmLGpdIoscagaGXU0mpm1T4LV/MhalHFc/msmLk7DpsnFU++PxqW
qZoKCVvXcNK0jBJsRU/Zi1WKj7tMjDmtPO0HRBwBIZp0SXC2sw3ggmeNNftM6efd
VE7/blMAmn4sFzXXWgw=
-----END CERTIFICATE-----
subject=/C=XX/L=Default City/O=Default Company Ltd/CN=nexus-tls.example.com
issuer=/C=XX/L=Default City/O=Default Company Ltd/CN=nexus-tls.example.com
---
No client certificate CA names sent
Peer signing digest: SHA256
Server Temp Key: X25519, 253 bits
---
SSL handshake has read 1603 bytes and written 377 bytes
Verification error: self signed certificate
---
New, TLSv1.2, Cipher is ECDHE-RSA-AES256-GCM-SHA384
Server public key is 2048 bit
Secure Renegotiation IS supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : ECDHE-RSA-AES256-GCM-SHA384
    Session-ID: 79A15F639F3896A0C9BC12FF46549069773B64BFEC5763BCD4C997ECFF54E8C2
    Session-ID-ctx:
    Master-Key: CC4749CA60E60F0A1F568EE0592FA422598030A0C31755568F69A1C226E58FCBCDCA13A29D44864F2EC4EDFE611B1586
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    TLS session ticket lifetime hint: 600 (seconds)
    TLS session ticket:
    0000 - 3c bd 0f f1 ba 97 ac 7d-8f 0c 19 49 f2 f2 80 a1   <......}...I....
    0010 - a2 66 0b 98 2e cb 44 97-01 de cd 03 ce 8e cb 4c   .f....D........L
    0020 - 7f 7b 1c 2f 67 35 8a e6-8b a4 f1 f7 e2 12 73 d1   .{./g5........s.
    0030 - 33 cd 0f 5c a1 51 ae c4-f9 64 c3 96 27 1f f9 8e   3..\.Q...d..'...
    0040 - 84 92 1e 7d 42 89 4d bd-54 35 1c 57 d2 e6 36 66   ...}B.M.T5.W..6f
    0050 - ab b6 77 65 ea af f9 81-d7 61 93 72 45 64 1c fb   ..we.....a.rEd..
    0060 - 43 d7 7c a1 75 a5 62 9f-a9 7c c9 b6 d2 4c c0 41   C.|.u.b..|...L.A
    0070 - 75 4a 9a 71 f7 d0 63 70-9f 58 7a 1f 4d bc 6f 1e   uJ.q..cp.Xz.M.o.
    0080 - fc 7f 37 a7 e2 01 bf 8c-bd 36 3d 5d 49 99 97 6b   ..7......6=]I..k
    0090 - f9 f6 4e 27 50 c0 e3 58-9b 29 0f 08 ac b8 1d 65   ..N'P..X.).....e
    00a0 - d4 ee c6 ca 40 23 ad c0-41 a2 89 c8 86 ab fb f5   ....@#..A.......
    00b0 - 22 d6 6f cf 13 46 ea 77-f5 19 68 ec a7 a8 4c d1   ".o..F.w..h...L.
    00c0 - 74 c8 7b aa 50 72 10 a0-d6 26 d7 38 7a b9 19 8d   t.{.Pr...&.8z...

    Start Time: 1588457259
    Timeout   : 7200 (sec)
    Verify return code: 18 (self signed certificate)
    Extended master secret: yes
---
DONE
```

### Openshift Route

In this example we'll make use of the `spec.networking.tls.mandatory` directive:

```yaml
apiVersion: apps.m88i.io/v1alpha1
kind: Nexus
metadata:
  name: nexus3
spec:
  # Number of Nexus pod replicas (can't be increased after creation)
  replicas: 1
  # Here you can specify the image version to fulfill your needs. Defaults to docker.io/sonatype/nexus3:latest if useRedHatImage is set to false
  #image: "docker.io/sonatype/nexus3:latest"
  # let's use the centOS image since we do not have access to Red Hat Catalog
  useRedHatImage: false
  # Set the resources requests and limits for Nexus pods. See: https://help.sonatype.com/repomanager3/system-requirements
  resources:
    limits:
      cpu: "2"
      memory: "2Gi"
    requests:
      cpu: "1"
      memory: "2Gi"
  # Data persistence details
  persistence:
    # Should we persist Nexus data? (turn this to false only if you're evaluating this resource)
    persistent: false
  # details regarding networking
  networking:
    # expose please
    expose: true
    # How do you want to expose the Nexus server? In this case, we're using Route. Only available on Openshift
    exposeAs: "Route"
    # Configuration related to TLS
    tls:
      # Causes all insecure traffic (HTTP) to be redirected to a secured port (HTTPS)
      mandatory: true
```

Let's create this resource:

```
$ oc apply -f examples/nexus3-centos-tls-route.yaml
nexus.apps.m88i.io/nexus3 created
```

Now we should have a Route with edge TLS termination and a redirection policy:

```
$ oc get route/nexus3                               
NAME      HOST/PORT                            PATH      SERVICES   PORT      TERMINATION     WILDCARD
nexus3    nexus3-nexus.192.168.42.189.nip.io             nexus3     8081      edge/Redirect   None
```

To be sure, let's take a look at the YAML as well:

```
$ oc get route/nexus3 -o yaml
apiVersion: route.openshift.io/v1
kind: Route
# (output omitted)
spec:
  host: nexus3-nexus.192.168.42.189.nip.io
  port:
    targetPort: 8081
  tls:
    insecureEdgeTerminationPolicy: Redirect
    termination: edge
  to:
    kind: Service
    name: nexus3
    weight: 100
  wildcardPolicy: None
status:
  ingress:
  - conditions:
    - lastTransitionTime: 2020-05-03T02:52:13Z
      status: "True"
      type: Admitted
    host: nexus3-nexus.192.168.42.189.nip.io
    routerName: router
    wildcardPolicy: None
```

There it is:

```yaml
tls:
  insecureEdgeTerminationPolicy: Redirect
  termination: edge
```

We can use `curl` to test it out. First let's gather some verbose output when connecting using HTTP as our protocol:

```
$ curl http://nexus3-nexus.192.168.42.189.nip.io -v                      
* Rebuilt URL to: http://nexus3-nexus.192.168.42.189.nip.io/
*   Trying 192.168.42.189...
* TCP_NODELAY set
* Connected to nexus3-nexus.192.168.42.189.nip.io (192.168.42.189) port 80 (#0)
> GET / HTTP/1.1
> Host: nexus3-nexus.192.168.42.189.nip.io
> User-Agent: curl/7.59.0
> Accept: */*
>
< HTTP/1.1 302 Found
< Cache-Control: no-cache
< Content-length: 0
< Location: https://nexus3-nexus.192.168.42.189.nip.io/
<
* Connection #0 to host nexus3-nexus.192.168.42.189.nip.io left intact
```

See that "Location" header? That's responsible for the redirect, but we have to run it again with the "-L" flag. We're also adding the "-k" option to ignore TLS verification failure as we're using a self-signed certificate (for more information consult the curl(1) man page):

```
$ curl -vLk http://nexus3-nexus.192.168.42.189.nip.io
* Rebuilt URL to: http://nexus3-nexus.192.168.42.189.nip.io/
*   Trying 192.168.42.189...
* TCP_NODELAY set
* Connected to nexus3-nexus.192.168.42.189.nip.io (192.168.42.189) port 80 (#0)
> GET / HTTP/1.1
> Host: nexus3-nexus.192.168.42.189.nip.io
> User-Agent: curl/7.59.0
> Accept: */*
>
< HTTP/1.1 302 Found
< Cache-Control: no-cache
< Content-length: 0
< Location: https://nexus3-nexus.192.168.42.189.nip.io/
<
* Connection #0 to host nexus3-nexus.192.168.42.189.nip.io left intact
* Issue another request to this URL: 'https://nexus3-nexus.192.168.42.189.nip.io/'
*   Trying 192.168.42.189...
* TCP_NODELAY set
* Connected to nexus3-nexus.192.168.42.189.nip.io (192.168.42.189) port 443 (#1)
# (output omitted)
```
