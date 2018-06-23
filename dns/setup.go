package dns

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/go-redis/redis"
	"github.com/gophish/healthcheck/config"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("healthcheck", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
	config.LoadConfig("./config.json")
}

func setup(c *caddy.Controller) error {
	c.Next() // healthcheck

	client := redis.NewClient(&redis.Options{
		Addr: config.Config.RedisAddr,
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return HealthCheckPlugin{
			Redis: client,
			Next:  next,
		}
	})

	return nil
}
