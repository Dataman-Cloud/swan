package handler

import (
	"net/http"

	"github.com/Dataman-Cloud/swan-janitor/src/config"
	"github.com/Dataman-Cloud/swan-janitor/src/upstream"
)

const HANDLER_FACTORY_KEY = "handler_factory"

type Factory struct {
	HttpHandlerCfg config.HttpHandler
	ListenerCfg    config.Listener
	UpstreamLoader upstream.UpstreamLoader
}

func NewFactory(cfg config.HttpHandler, listenerCfg config.Listener) *Factory {
	return &Factory{
		HttpHandlerCfg: cfg,
		ListenerCfg:    listenerCfg,
	}
}

func (factory *Factory) HttpHandler(us *upstream.Upstream) http.Handler {
	return NewHTTPProxy(&http.Transport{}, factory.HttpHandlerCfg, factory.ListenerCfg, us, factory.UpstreamLoader)
}
