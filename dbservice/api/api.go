package api

import (
	"context"
	"database-service/auth"
	"database-service/cache"
	"database-service/dbservice/api/config"
	"database-service/dbservice/api/controller"
	"database-service/dbservice/api/logger"
	"database-service/dbservice/api/services"
	"database-service/domain"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	cachepkg "github.com/go-redis/cache/v9"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const MetricPath = "/metrics"

type Api struct {
	config     *config.Values
	router     *gin.Engine
	redis      *redis.Client
	db         *sql.DB
	logger     logger.Logger
	metricsReg *prometheus.Registry
	requestsM  *prometheus.HistogramVec
}

func NewApi() (api Api, err error) {
	api.config, err = config.ReadConfig()
	if err != nil {
		return
	}

	l, err := zap.NewProduction()
	if err != nil {
		return
	}
	api.logger = l.Sugar()

	if err = api.initConnections(); err != nil {
		return
	}

	if err = api.initMetrics(); err != nil {
		return
	}

	api.router = gin.New()

	api.initMiddleware()

	api.InitRoutes()

	return api, nil

}

func (api *Api) initConnections() (err error) {
	api.db, err = NewDB(api.config)
	if err != nil {
		api.logger.Error(err)
		return
	}
	api.redis, err = NewRedis(api.config)
	if err != nil {
		api.logger.Error(err)
		return
	}
	return nil
}

func (api *Api) initMetrics() (err error) {
	api.metricsReg = prometheus.NewRegistry()
	api.requestsM = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "database_service_requests",
		Help:    "database-service requests duration seconds",
		Buckets: []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10},
	}, []string{"route", "status_code"})
	return api.metricsReg.Register(api.requestsM)
}

func (api *Api) initMiddleware() {
	api.router.Use(api.GetMetricsHandler())
}

func (api *Api) Run() {
	err := api.router.Run()
	if err != nil {
		panic(err)
	}
}

func (api *Api) InitRoutes() {
	cacheClient := cachepkg.New(&cachepkg.Options{
		Redis:      api.redis,
		LocalCache: cachepkg.NewTinyLFU(1000, api.config.GetCacheTTL()),
	})
	cacheAdapter := cache.NewCacheAdapter(api.config.GetCacheTTL(), cacheClient)
	dbDataRepo := domain.NewDbDataRepository(api.db)
	cachedRepo := domain.NewCachedDataRepository(dbDataRepo, cacheAdapter, api.config.Cache.KeyPrefix)
	usersRepo := domain.NewUserRepository(api.db, api.logger)
	authService := auth.NewJwtService(api.config.JwtCodePhrase)
	dataService := services.NewDataService(cachedRepo, usersRepo)

	dataController := controller.NewDataController(authService, dataService, api.logger)

	dataController.AddRoutes(api.router)

	api.router.GET(MetricPath, gin.WrapH(promhttp.HandlerFor(api.metricsReg, promhttp.HandlerOpts{})))
}

func (api *Api) GetMetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == MetricPath {
			c.Next()
			return
		}
		startTime := time.Now().UTC()
		c.Next()
		dur := time.Now().UTC().Sub(startTime)
		route := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		status := strconv.Itoa(c.Writer.Status())
		api.requestsM.WithLabelValues(route, status).Observe(dur.Seconds())
	}
}

func NewDB(config *config.Values) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.GetDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	err = db.PingContext(timeout)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewRedis(config *config.Values) (*redis.Client, error) {
	rcl := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		Username: config.Redis.Username,
		DB:       config.Redis.Db,
	})
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	_, err := rcl.Ping(ctx).Result()
	return rcl, err
}
