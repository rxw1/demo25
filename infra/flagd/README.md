**NOTE** The value of `data.flags.json` from `configmap-flags.yaml` is copied
to `flags.json`, which is used by Docker Compose, since the latter can not live
in a Helm chart directory.
