# Cirrus Webhooks Server

Examples of the webhook event processors from the Cirrus CI.

## Processors

## DataDog

This processor receives, enriches and streams Cirrus CI webhook events to DataDog.

The simplest way to try this processor is to use Docker and [ngrok](https://ngrok.com/).

First, you'll need to run the DataDog agent locally. The easiest way to do that would probably be [a Docker container](https://docs.datadoghq.com/containers/docker/?tab=standard), just don't forget to expose the DogStatsD port by adding the following argument to `docker run`:

```
-p 8125:8125/udp
```

Then, run the DataDog processor:

```sh
docker run -it --rm -p 8080:8080 ghcr.io/cirruslabs/cirrus-webhooks-server:latest datadog --dogstatsd-addr=host.docker.internal:8125
```

Note the `--dogstatsd-addr=host.docker.internal:8125`, it shows the DataDog processor where to find the DogStatsD daemon. In this case, it's on the Docker's host machine.

Finally, [install](https://ngrok.com/download) and run `ngrok` to expose our DataDog processor's HTTP server to the internet:

```sh
ngrok http 8080
```

This will open the following TUI window:

![](docs/ngrok-http-8080.png)

You'll need to copy the forwarding address and set it in your organization's settings in the Cirrus CI app:

![](docs/cirrus-ci-webhook-settings.png)

Now you can run some tasks, and the corresponding audit events will appear shortly in your "Events Explorer":

![](docs/datadog-webhook-event.png)
