// Package giam_traefik_plugin a plugin to integrate Traefik with Giam.
package giam_traefik_plugin

import (
	"context"
	"net/http"

	authorizationhandler "github.com/usegiam/giam-traefik-plugin/internal/authorization/handler"
	authorizationservice "github.com/usegiam/giam-traefik-plugin/internal/authorization/service"
	lokihandler "github.com/usegiam/giam-traefik-plugin/internal/datasource/loki/handler"
	lokiservice "github.com/usegiam/giam-traefik-plugin/internal/datasource/loki/service"
	prometheushandler "github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus/handler"
	prometheusservice "github.com/usegiam/giam-traefik-plugin/internal/datasource/prometheus/service"
	"github.com/usegiam/giam-traefik-plugin/internal/handler"
	"github.com/usegiam/giam-traefik-plugin/pkg/grafana"
	"github.com/usegiam/giam-traefik-plugin/pkg/hash"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

type Config struct {
	APIKey     string `yaml:"APIKey"`
	APIUrl     string `yaml:"APIUrl"`
	GrafanaUrl string `yaml:"GrafanaUrl"`
	LogLevel   string `yaml:"LogLevel"`
}

func CreateConfig() *Config {
	return &Config{}
}

type Plugin struct {
	next     http.Handler
	name     string
	config   *Config
	handlers []handler.Handler
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	hashSvc := hash.NewService()
	logger := log.New(config.LogLevel)
	lokiSvc := lokiservice.New(&lokiservice.Deps{
		APIUrl: config.APIUrl,
		APIKey: config.APIKey,
		Logger: logger,
	})
	prometheusSvc := prometheusservice.New(&prometheusservice.Deps{
		APIUrl: config.APIUrl,
		APIKey: config.APIKey,
		Logger: logger,
	})
	authorizationSvc := authorizationservice.NewService(&authorizationservice.Deps{
		APIUrl: config.APIUrl,
		APIKey: config.APIKey,
		Logger: logger,
	})
	grafanaRepo := grafana.NewRepo(config.GrafanaUrl, logger)
	handlers := []handler.Handler{
		authorizationhandler.NewDatasourceHandler(&authorizationhandler.DatasourceHandlerDeps{
			Logger:      logger,
			GrafanaRepo: grafanaRepo,
			Service:     authorizationSvc,
		}),
		lokihandler.NewQueryHandler(&lokihandler.QueryHandlerDeps{
			Service:     lokiSvc,
			GrafanaRepo: grafanaRepo,
			Logger:      logger,
		}),
		lokihandler.NewSeriesHandler(&lokihandler.SeriesHandlerDeps{
			Service:     lokiSvc,
			GrafanaRepo: grafanaRepo,
			Logger:      logger,
		}),
		lokihandler.NewLabelValueHandler(&lokihandler.LabelValuesHandlerDeps{
			Service:     lokiSvc,
			GrafanaRepo: grafanaRepo,
			Logger:      logger,
		}),
		prometheushandler.NewQueryHandler(&prometheushandler.QueryHandlerDeps{
			Logger:        logger,
			GrafanaRepo:   grafanaRepo,
			HashSvc:       hashSvc,
			PrometheusSvc: prometheusSvc,
		}),
		prometheushandler.NewSeriesHandler(&prometheushandler.SeriesHandlerDeps{
			Logger:      logger,
			GrafanaRepo: grafanaRepo,
			Service:     prometheusSvc,
		}),
	}

	finalHandler := handler.ChainHandlers(next, handlers...)

	return &Plugin{
		next:     finalHandler,
		name:     name,
		config:   config,
		handlers: handlers,
	}, nil
}

func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.next.ServeHTTP(rw, req)
}
