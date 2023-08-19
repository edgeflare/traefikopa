## Plugin development environment

Create a working directory, say, traefik-opa-plugin.

```sh
mkdir traefik-opa-plugin && cd traefik-opa-plugin
```

Inside working dir, create Traefik static config, `traefik.yml` with below content

```yaml
log:
  level: INFO

entryPoints:
  web:
    address: ":80"

providers:
  file:
    filename: dynamic_conf.yml  # Config file for dynamic configuration
    watch: true  # watch file changes

api:
  insecure: true  # Enables the dashboard on http://localhost/dashboard/
  dashboard: true

experimental:
  localPlugins:
    opa: # custom name for the plugin, used in the middleware
      moduleName: github.com/edgeflare/traefikopa
```

And treafik dynamic config `dynamic_conf.yml` with

```yaml
http:
  routers:
    my-router:
      rule: host(`opa.develop.local`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - traefik-opa

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:8080 # downstream backend server
  
  middlewares:
    traefik-opa: # custom name for the middleware
      plugin:
        opa: # custom name for the plugin
          url: http://localhost:8181/v1/data/httpapi/authz
```

Clone this repo into a nested directory like

```sh
mkdir -p plugins-local/src/github.com/edgeflare
git clone git@github.com:edgeflare/traefikopa.git plugins-local/src/github.com/edgeflare/traefikopa
```

For testing have your OPA server and backend server ready. Now start Traefik

```sh
traefik
```

And get going! Any contribution is more than welcome :)