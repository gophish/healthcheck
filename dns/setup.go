package dns

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/gophish/healthcheck/config"
	"github.com/gophish/healthcheck/db"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("healthcheck", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next() // healthcheck

	err := config.LoadConfig("./config.json")
	if err != nil {
		return err
	}

	err = db.Setup()
	if err != nil {
		return err
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return HealthCheckPlugin{
			Next: next,
		}
	})

	return nil
}
