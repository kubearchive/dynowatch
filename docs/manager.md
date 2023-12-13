# Dynowatch Manager

`make run` launches the Dynowatch _manager_, which starts the controllers that emit CloudEvents.
The `dynowatch.yaml` file contains the configuration for the controllers and other global settings.
This file should be placed in one of the following directories:

- `/etc/dynowatch`
- `$HOME/.dynowatch`
- `./config/dynowatch`

## dynowatch.yaml Schema

Example file:

```yaml
metrics:
  bind-address: ":9000"
healthz:
  bind-address: ":9001"
leader-election: true
cloud-events:
  source-uri: https://github.com/kubarchive/dynowatch
  target-address: https://splunk.mycorp.com/events
watches:
  - name: deployments
    group: apps
    version: v1
    kind: Deployment
  - name: jobs
    group: batch
    version: v1
    kind: Job
```

| Field | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `metrics.bind-addres` | `string` | `:8080` | Port that the controller metrics endpoint binds to |
| `healthz.bind-address` | `string` | `:8081` | Port that the controller's health endpoint binds to |
| `leader-election` | `bool` | `false` | If true, enable leader election for high availability |
| `cloudevents.source-uri` | `string` | `localhost` | URI that identifies the source of the events |
| `cloudevents.target-address` | `string` | `http://localhost:8082` | Address to send CloudEvents to |
| `watches.[*]` | `array` | Empty | List of objects to watch with a controller. Each watch must have a `name`, `group`, `version`, and `kind`. |
