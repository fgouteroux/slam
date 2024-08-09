package webserver

import (
	"strings"
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
	debug   bool
	msgTmpl *template.Template

	msgFailedSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "slam",
			Name:      "message_sent_failed_total",
			Help:      "The total number of failed messages sent.",
		},
		[]string{"channel"},
	)

	msgSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "slam",
			Name:      "message_sent_total",
			Help:      "The total number of successfully messages sent.",
		},
		[]string{"channel"},
	)

	localCache    *memcache.MemCache
	redisCache    *redigo.Pool
	redisCacheTTL int
)

func init() {
	prometheus.MustRegister(msgFailedSent, msgSent)
}

type webserver struct {
	Cache                       string
	Slack                       *slack.Client
	Prometheus                  *ginprometheus.Prometheus
	TemplateTitleAnnotation     string
	TemplateTitleLinkAnnotation string
}

// New webserver
func New(s *slack.Client, cache, templateTitleAnnotation, templateTitleLinkAnnotation string) *webserver {
	return &webserver{
		Slack:                       s,
		Cache:                       cache,
		TemplateTitleAnnotation:     templateTitleAnnotation,
		TemplateTitleLinkAnnotation: templateTitleLinkAnnotation,
	}
}

// Init a webserver with Gin
func (ws *webserver) Init(debugEnabled bool, templateFiles, redisHost string, redisDB, redisKeyTTL int) *gin.Engine {
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

		redisCacheTTL = redisKeyTTL
	}

	debug = debugEnabled

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	ws.Prometheus = ginprometheus.NewPrometheus("gin")
	ws.Prometheus.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, p := range c.Params {
			if p.Key == "channel" {
				url = strings.Replace(url, p.Value, ":channel", 1)
				break
			}
		}
		return url
	}

	ws.Prometheus.Use(router)

	router.POST("/webhook/:channel", checkErr(ws.handleWebhook))

	router.GET("/health", ws.healthHandler)
	router.GET("/ready", ws.readyHandler)
	return router
}

func checkErr(f func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := f(c)
		if err != nil {
			log.Error(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
	}
}
