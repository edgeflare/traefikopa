# Open Policy Agent (OPA) Authorization middleware for Traefik

> ### This plugin is useful if the full request context is needed for evaluating OPA policy decision. Traefik forwardAuth middleware doesn't preserve the request entirely, stripping off, for example, the `body`, before forwarding to the authz server. If you can NOT modify Traefik installation, you might checkout the simpler [traefik-opa-proxy](https://github.com/edgeflare/traefik-opa-proxy) which has some limitations, though.

## Installtion

### Using Helm

```yaml
apiVersion: helm.cattle.io/v1
kind: HelmChart # or HelmChartConfig
metadata:
  name: traefik
  namespace: kube-system
spec:
  valuesContent: |-
    additionalArguments:
      - "--experimental.plugins.opa.moduleName=github.com/edgeflare/traefikopa"
      - "--experimental.plugins.opa.version=v0.0.1"
#     - others-additional-arguments
```

### Using command line arguments

```sh
traefik \
  --experimental.plugins.opa.moduleName=github.com/edgeflare/traefikopa \
  --experimental.plugins.opa.version=v0.0.1
```

## Usage in Kubernetes

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: opa-authz
  namespace: kube-system
spec:
  plugin:
    opa:
      URL: http://opa.kube-system:8181/v1/data/httpapi/authz
      # Assuming OPA is installed in kube-system namespace
      # and exposed via a service named opa on port 8181
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: yourapp.example.com
  namespace: demo
spec:
  entryPoints:
  - web
  - websecure
  routes:
  - match: Host(`yourapp.example.com`) 
    kind: Rule
    services:
    - name: yourapp-service
      port: 80
    middlewares:
    - name: opa-authz
  tls: # optional
    secretName: yourapp.example.com-tls
---
# Use either IngressRoute, or Ingress
kind: Ingress
metadata:
  name: yourapp.example.com
  namespace: demo
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.ingress.kubernetes.io/router.middlewares: kube-system-opa-authz@kubernetescrd
spec:
  rules:
  - host: yourapp.example.com
    http:
      paths:
      - backend:
          service:
            name: yourapp-service
            port:
              number: 80
        path: /
```

See [example](https://github.com/edgeflare/traefik-opa-proxy/tree/master/example) for Kubernetes deployment manifests.