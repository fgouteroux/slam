# slam
Send alertmanager alerts notifications to slack without creating webhook url for each slack channel.

It update the original slack channel message to avoid searching if an alert is resolved or not.

![](examples/slam.gif)

## usage

```
Usage of slam:
  -cache string
    	Cache type (local or redis) (default "local")
  -cache.redis.db int
    	Redis DB
  -cache.redis.host string
    	Redis host  (default "localhost:6379")
  -debug
    	Enable debug mode
  -server.grace-timeout int
    	 Grace timeout for shutdown HTTP server (default 5)
  -server.http-idle-timeout int
    	 Idle timeout for HTTP server (default 30)
  -server.http-listen-address string
    	 Listen address for HTTP server
  -server.http-listen-port int
    	 Listen address for HTTP server (default 8080)
  -server.http-read-timeout int
    	Read timeout for HTTP server (default 30)
  -server.http-write-timeout int
    	Write timeout for HTTP server (default 30)
  -server.tls-cert-path string
    	TLS certificate path for HTTP server
  -server.tls-key-path string
    	TLS key path for HTTP server
  -slack-token string
    	Slack app token (could be set by SLACK_TOKEN env var)
  -template-files string
    	Template files to load (files identified by the pattern, like *.tmpl)
  -version
    	show version
```

## How it's works

slam use the webhook config from alertmanager: https://prometheus.io/docs/alerting/latest/configuration/#webhook_config

slam use the `groupKey` from webhook json payload, to identify if:
* the slack message has already been sent (for firing status)
* the original slack message should be updated (for resolved status)

This key is stored in local memory or redis.

If the key is not found, it send a new message in the slack channel.


## Template

As using webhook format, there is no templating from alertmanager. So we enable the template feature in slam to allow formating slack message. By default, we apply a simple slack message format (cf gif image)

It's possible to define different template and to choose it in the query url.

Run slam with template location:
```
slam -template-files examples/*.tmpl
```

Send an alert and use slack.tmpl file template:
```
curl "http://localhost:8080/webhook/mychann?template=slack.tmpl" -X POST -H "Content-type: application/json" -d @payload.json
```

## Limitations

If the webhook payload contains several alerts, it will wait that all alerts be resolved before update the original message.
