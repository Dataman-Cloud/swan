package nameserver

import (
	"strings"
	"time"

	"github.com/urfave/cli"
)

type Config struct {
	Domain          string        `json:"domain"`
	Listener        string        `json:"ip"`
	Port            int           `json:"port"`
	Resolvers       []string      `json:"resolvers"`
	ExchangeTimeout time.Duration `json:"exchange_timeout"`

	SOAname    string `json:"soaname"`
	SOARname   string `json:"soarname"`
	SOAMname   string `json:"soamname"`
	SOASerial  uint32 `json:"soaserial"`
	SOARefresh uint32 `json:"soarefresh"`
	SOARetry   uint32 `json:"soaretry"`
	SOAExpire  uint32 `json:"soaexpire"`

	RecurseOn bool `json:"recurse_on"`

	TTL int `json:"ttl"`
}

func NewConfig(ctx *cli.Context) *Config {
	config := &Config{}

	config.Domain = ctx.String("domain")
	config.Listener = ctx.String("listener")
	config.Port = ctx.Int("port")
	config.ExchangeTimeout = ctx.Duration("exchange-timeout")

	if len(strings.Replace(ctx.String("resolvers"), " ", "", -1)) > 0 && strings.Contains(ctx.String("resolvers"), ",") {
		config.Resolvers = strings.Split(ctx.String("resolvers"), ",")
	} else {
		config.Resolvers = []string{"114.114.114.114"}
	}

	if ctx.IsSet("soarname") {
		config.SOARname = ctx.String("soarname")
	} else {
		config.SOARname = ""
	}

	if ctx.IsSet("soamname") {
		config.SOAMname = ctx.String("soamname")
	}

	if ctx.IsSet("soaserial") {
		config.SOASerial = uint32(ctx.Int("soaserial"))
	}

	if ctx.IsSet("soarefresh") {
		config.SOARefresh = uint32(ctx.Int("soarefresh"))
	}

	if ctx.IsSet("soaretry") {
		config.SOARetry = uint32(ctx.Int("soaretry"))
	}

	if ctx.IsSet("soaexpire") {
		config.SOAExpire = uint32(ctx.Int("soaexpire"))
	}

	if ctx.IsSet("recurseon") {
		config.RecurseOn = ctx.Bool("recurseon")
	}

	if ctx.IsSet("ttl") {
		config.TTL = ctx.Int("ttl")
	}

	return config
}
