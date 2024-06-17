# slam
Send alertmanager alerts notifications to slack without creating webhook url for each slack channel.

It update the original slack channel message to avoid searching if an alert is resolved or not.

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