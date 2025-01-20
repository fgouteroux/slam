package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"github.com/fgouteroux/slam/webserver"
)

var Version = "dev"
var BuildTime string

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	debug := flag.Bool("debug", os.Getenv("DEBUG") == "true", "Enable debug mode")

	slackToken := flag.String("slack-token", os.Getenv("SLACK_TOKEN"), "Slack app token (could be set by SLACK_TOKEN env var)")
	slackMsgLengthLimit := flag.Int("slack-msg-length-limit", 1000, "Slack message length limit before truncate.")

	templateFiles := flag.String("template-files", "", "Template files to load (files identified by the pattern, like *.tmpl)")

	cache := flag.String("cache", "local", "Cache type (local or redis)")

	templateTitleAnnotation := flag.String("template.annotation.title", "summary", "Annotation key name to get for setting slack title message")
	templateTitleLinkAnnotation := flag.String("template.annotation.title-link", "title_link", "Annotation key name to get for setting slack title link message")

	redisDB := flag.Int("cache.redis.db", 0, "Redis DB")
	redisHost := flag.String("cache.redis.host", "localhost:6379", "Redis host")
	redisKeyTTL := flag.Int("cache.redis.key-ttl", 1296000, "Redis key ttl in seconds")

	serverListenAddress := flag.String("server.http-listen-address", "", " Listen address for HTTP server")
	serverListenPort := flag.Int("server.http-listen-port", 8080, " Listen address for HTTP server")
	serverTLSCertPath := flag.String("server.tls-cert-path", "", "TLS certificate path for HTTP server")
	serverTLSKeyPath := flag.String("server.tls-key-path", "", "TLS key path for HTTP server")
	serverReadTimeout := flag.Int("server.http-read-timeout", 30, "Read timeout for HTTP server")
	serverWriteTimeout := flag.Int("server.http-write-timeout", 30, "Write timeout for HTTP server")
	serverIdleTimeout := flag.Int("server.http-idle-timeout", 30, " Idle timeout for HTTP server")
	serverGraceTimeout := flag.Int("server.grace-timeout", 5, " Grace timeout for shutdown HTTP server")

	version := flag.Bool("version", false, "show version")
	required := []string{
		"slack-token",
	}
	flag.Parse()

	if *version {
		fmt.Println("Version:\t", Version)
		fmt.Println("Build Time:\t", BuildTime)
		os.Exit(0)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	seen := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != "" {
			seen[f.Name] = true
		}
	})

	for _, req := range required {
		if !seen[req] {
			log.Errorf("missing required -%s argument/flag", req)
			os.Exit(2)
		}
	}

	slackClient := slack.New(*slackToken)

	authResult, err := slackClient.AuthTest()
	if err != nil {
		log.Error(err)
		return
	}
	if *debug {
		log.Debugf("Using '%s' in '%s' slack workspace (%s)", authResult.User, authResult.Team, authResult.URL)
	}

	ws := webserver.New(slackClient, *cache, *templateTitleAnnotation, *templateTitleLinkAnnotation, *slackMsgLengthLimit)

	listenAddr := fmt.Sprintf("%s:%d", *serverListenAddress, *serverListenPort)
	srv, idleConnectionsClosed := webserver.NewServerWithTimeout(
		listenAddr,
		*serverReadTimeout,
		*serverWriteTimeout,
		*serverIdleTimeout,
		time.Duration(*serverGraceTimeout)*time.Second,
	)
	srv.Handler = ws.Init(*debug, *templateFiles, *redisHost, *redisDB, *redisKeyTTL)

	if *serverTLSCertPath != "" && *serverTLSKeyPath != "" {
		err = srv.ListenAndServeTLS(*serverTLSCertPath, *serverTLSKeyPath)
	} else {
		log.Debugf("Server listen on: %s", listenAddr)
		err = srv.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("Error: Cannot serve requests on '%s' : %v", listenAddr, err)
		os.Exit(1)
	}

	<-idleConnectionsClosed
	log.Info("server exited")

}
