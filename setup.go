

package httpfetch

import (
	"errors"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"time"

	"github.com/caddyserver/caddy"
)

// init registers this plugin.
func init() { plugin.Register("httpfetch", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {

	httpfetchPlugin, err := newHttpFetch(c)
	if err != nil {
		return plugin.Error("httpfetch", err)
	}

	// Add a startup function that will -- after all plugins have been loaded -- check if the
	// prometheus plugin has been used - if so we will export metrics. We can only register
	// this metric once, hence the "once.Do".
	c.OnStartup(func() error {
		once.Do(func() { metrics.MustRegister(c, requestCount) })
		return nil
	})

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return httpfetchPlugin
	})

	// All OK, return a nil error.
	return nil
}

func newHttpFetch(c *caddy.Controller) (Httpfetch, error) {

	url := ""
	token := ""
	localCacheDuration := ""
	duration := time.Second
	var err error

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "url":
					if !c.NextArg() {
						c.ArgErr()
					}
					url = c.Val()

				case "token":
					if !c.NextArg() {
						c.ArgErr()
					}
					token = c.Val()

				case "localCacheDuration":
					if !c.NextArg() {
						c.ArgErr()
					}
					localCacheDuration = c.Val()
					duration, err = time.ParseDuration(localCacheDuration)
					if err != nil {
						localCacheDuration = ""
					}
				}

				if !c.Next() {
					break
				}
			}
		}

	}

	if url == "" || token == "" || localCacheDuration == "" {
		return Httpfetch{}, errors.New("Could not parse httpfetch config")
	}

	return Httpfetch{ReqUrl: url, Token: token, CacheDuration: duration}, nil

}
