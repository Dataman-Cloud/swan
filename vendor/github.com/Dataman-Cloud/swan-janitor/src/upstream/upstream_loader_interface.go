package upstream

import (
	"strings"

	"github.com/Dataman-Cloud/swan-janitor/src/config"

	"golang.org/x/net/context"
)

var UpstreamLoaderKey string

type UpstreamLoader interface {
	Poll()
	List() []*Upstream
	Get(serviceName string) *Upstream
	Remove(upstream *Upstream)
	ChangeNotify() <-chan bool
}

func InitAndStartUpstreamLoader(ctx context.Context, Config config.Config) (UpstreamLoader, error) {
	var upstreamLoader UpstreamLoader
	var err error
	switch strings.ToLower(Config.Upstream.SourceType) {
	case "swan":
		UpstreamLoaderKey = SWAN_UPSTREAM_LOADER_KEY
		upstreamLoader, err = InitSwanUpstreamLoader(Config.Listener.IP, Config.Listener.DefaultPort, Config.Listener.DefaultProto)
		if err != nil {
			return nil, err
		}
	}

	return upstreamLoader, nil
}
