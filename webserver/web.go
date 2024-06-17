package webserver

import (
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/zsais/go-gin-prometheus"

	"github.com/fgouteroux/slam/memcache"
	"github.com/fgouteroux/slam/redis"
	redigo "github.com/gomodule/redigo/redis"
)

var (
	debug         bool
	msgTmpl       *template.Template
	errorsSending = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slam_message_sent_errors",
		Help: "Number of errors posting message to channels.",
	})
	localCache *memcache.MemCache
	redisCache *redigo.Pool
)

func init() {
	prometheus.MustRegister(errorsSending)
}

type webserver struct {
	Cache      string
	Slack      *slack.Client
	Prometheus *ginprometheus.Prometheus
}

// New webserver
func New(s *slack.Client, cache string) *webserver {
	return &webserver{
		Slack: s,
		Cache: cache,
	}
}

// Init a webserver with Gin
func (ws *webserver) Init(debugEnabled bool, templateFiles, redisHost string, redisDB int) *gin.Engine {
	if templateFiles != "" {
		msgTmpl = template.Must(template.ParseGlob(templateFiles))
	}

	if ws.Cache == "local" {
		localCache = memcache.NewLocalCache()
	} else if ws.Cache == "redis" {
		redisCache = redis.Connect(redisHost, redisDB)

		redisReady := redis.Ping(redisCache)
		if redisReady != nil {
			log.Fatal("Redis not ready. ", redisReady)
		} else {
			log.Info("Redis is ready: ", redisHost)
		}
	}

	debug = debugEnabled

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	if ws.Prometheus != nil {
		ws.Prometheus.Use(router)
	}

	router.POST("/webhook/:channel", checkErr(ws.handleWebhook))

	router.GET("/health", ws.healthHandler)
	return router
}

func checkErr(f func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			errorsSending.Inc()
			log.Error(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
	}
}
