# Cirrus Webhooks Server

Examples of the webhook event processors from the Cirrus CI.

## Datadog processor

This processor receives, enriches and streams Cirrus CI webhook events to Datadog.

### Usage

```
docker run -it --rm ghcr.io/cirruslabs/cirrus-webhooks-server:latest datadog
```

The following command-line arguments are supported:

* `--api-key` (`string`) — enables sending events via the Datadog API using the specified API key
* `--api-site` (`string`) — specifies the [Datadog site](https://docs.datadoghq.com/getting_started/site/) to use when sending events via the Datadog API (defaults to `datadoghq.com`)
* `--dogstatsd-addr` — enables sending events via the DogStatsD protocol to the specified address (for example, `--dogstatsd-addr=127.0.0.1:8125`)
* `--event-types` (`string`) — comma-separated list of the event types to limit processing to (for example, --event-types=audit_event or --event-types=build,task
* `--http-addr` (`string`) — address on which the HTTP server will listen on (defaults to `:8080`)
* `--http-path` (`string`) — HTTP path on which the webhook events will be expected (defaults to `/`)
* `--secret-token` (`string`) — if specified, this value will be used as a HMAC SHA-256 secret to verify the webhook events

### Example

The simplest way to try this processor is to use Docker and [ngrok](https://ngrok.com/).

First, obtain the API key from the Datadog's `Organization Settings` → `API Keys`.

Then, run the Datadog processor:

```sh
docker run -it --rm -p 8080:8080 ghcr.io/cirruslabs/cirrus-webhooks-server:latest datadog --api-key=$DD_API_KEY
```

Finally, [install](https://ngrok.com/download) and run `ngrok` to expose our Datadog processor's HTTP server to the internet:

```sh
ngrok http 8080
```

This will open the following TUI window:

![](docs/ngrok-http-8080.png)

You'll need to copy the forwarding address and set it in your organization's settings in the Cirrus CI app:

![](docs/cirrus-ci-webhook-settings.png)

Now you can run some tasks, and the corresponding audit events will appear shortly in your "Events Explorer":

![](docs/datadog-webhook-event.png)
