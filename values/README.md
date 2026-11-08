# Kubernetes manifests (Helm values)

This service keeps **no Helm templates of its own**. Manifests are rendered by
the centralized golden chart
[`InventiumOrg/platform-microservice`](https://github.com/InventiumOrg/platform-microservice),
and this directory holds only the values that drive it — the "values only"
pattern described in that chart's `examples/README.md`.

```
values/
  base.yaml    environment-independent settings (always included first)
  local.yaml   local cluster overlay
  dev.yaml     dev cluster overlay
  prod.yaml    production overlay
```

`base.yaml` is never used alone; it is always combined with exactly one
environment overlay, overlay last so it wins.

## Chart version

| | |
|---|---|
| Chart | `oci://596484148719.dkr.ecr.ap-southeast-1.amazonaws.com/inventium/charts/platform-microservice` |
| Pinned version | `2.0.0` |

Pin the version explicitly on every command so a chart release never changes a
deploy implicitly.

## Render (no cluster needed)

```bash
CHART=oci://596484148719.dkr.ecr.ap-southeast-1.amazonaws.com/inventium/charts/platform-microservice

helm template inventory-service "$CHART" --version 2.0.0 \
  -f values/base.yaml -f values/dev.yaml
```

Against a local checkout of the chart instead:

```bash
helm template inventory-service ../platform-microservice \
  -f values/base.yaml -f values/dev.yaml
```

## Diff against a live cluster

```bash
helm template inventory-service "$CHART" --version 2.0.0 \
  -f values/base.yaml -f values/prod.yaml | kubectl diff -f -
```

## Install / upgrade

```bash
helm upgrade --install inventory-service "$CHART" --version 2.0.0 \
  --namespace inventory-service --create-namespace \
  -f values/base.yaml -f values/dev.yaml \
  --set image.tag="$BUILD_TAG"
```

CI supplies the image tag with `--set image.tag=$BUILD_TAG`, which overrides the
placeholder tag in the overlay.

## Prerequisites per namespace

Neither secret is stored in git. Create both once per namespace before the first
deploy:

```bash
# Application credentials: DB_SOURCE, CLERK_KEY, OTEL_EXPORTER_OTLP_HEADERS,
# LOKI_PASSWORD / TEMPO_PASSWORD / PROMETHEUS_PASSWORD
kubectl create secret generic inventory-service-env \
  --from-env-file=app.env -n inventory-service

# Kafka CA bundle, mounted at /etc/secrets/ca.pem
kubectl create secret generic inventory-service-kafka-ca \
  --from-file=ca.pem=ca.pem -n inventory-service
```

## What renders

| Object | local | dev | prod |
|---|:--:|:--:|:--:|
| Deployment (app + OTel sidecar) | ✅ | ✅ | ✅ |
| Service | ✅ | ✅ | ✅ |
| ConfigMap (app config + OTel config) | ✅ | ✅ | ✅ |
| HTTPRoute (Kong gateway) | ✅ | ✅ | ✅ |
| HorizontalPodAutoscaler | — | — | ✅ |
| PodDisruptionBudget | — | — | ✅ |

## Conventions worth keeping

- **Probes target the port by name** (`port: http`), so the port number lives in
  `service.port` only and cannot drift out of sync.
- **`config:` is a ConfigMap — never put credentials in it.** Secret values go
  in the `inventory-service-env` secret, listed under `existingSecrets`.
- **Hostnames and image tags belong in overlays**, never in `base.yaml`.
