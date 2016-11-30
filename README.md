# k8s-ns-meddler
[![wercker status](https://app.wercker.com/status/437468793eefe633f79a7e2fe535aaf0/s/master
"wercker status")](https://app.wercker.com/project/byKey/437468793eefe633f79a7e2fe535aaf0)

k8s-ns-meddler will watch Kubernetes API for new namespaces and create new
secret (based on some existing secret) in each.

This tool was written to simplify working with namespaces and secured Ingress.
In order to use TLS on your Ingress resource, you need to specify `secretName`
in the Ingress spec, and this Secret needs to be in the same namespace as
the Ingress itself.

```
Usage:
  k8s-ns-meddler [OPTIONS]

Application Options:
  -d, --debug         Log debug messages
  -p, --port=         Port to listen on for /health (default: 8080)
  -s, --sourcesecret= Source secret name to copy data from
  -t, --targetsecret= Targed secret name to create
  -v, --version       Show version

Help Options:
  -h, --help          Show this help message
```
