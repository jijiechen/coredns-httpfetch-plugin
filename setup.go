

package httpfetch

import (
	"errors"
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"strings"
)

// init registers this plugin.
func init() { plugin.Register("httpfetch", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {

	httpFetchPlugin, err := newHttpFetch(c)
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
		return httpFetchPlugin
	})

	// All OK, return a nil error.
	return nil
}

func newHttpFetch(c *caddy.Controller) (HttpFetch, error) {
	url := ""
	method := "GET"
	query := "dns_name=%s"
	body := ""
	ipAnalyzer := ""
	ttlAnalyzer := ""
	var headers []string

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
					case "url":
						if !c.NextArg() {
							c.ArgErr()
						}
						url = strings.TrimRight(c.Val(), `\n`)

					case "method":
						if !c.NextArg() {
							c.ArgErr()
						}
						method = strings.TrimRight(c.Val(), `\n`)

					case "query":
						if !c.NextArg() {
							c.ArgErr()
						}
						query = strings.TrimRight(c.Val(), `\n`)

					case "body":
						if !c.NextArg() {
							c.ArgErr()
						}
						body = strings.TrimRight(c.Val(), `\n`)

					case "header":
						if !c.NextArg() {
							c.ArgErr()
						}
						headers = append(headers, strings.TrimRight(c.Val(), `\n`))

					case "analyze_ip":
						if !c.NextArg() {
							c.ArgErr()
						}
						ipAnalyzer = strings.TrimRight(c.Val(), `\n`)

					case "analyze_ttl":
						if !c.NextArg() {
							c.ArgErr()
						}
						ttlAnalyzer = strings.TrimRight(c.Val(), `\n`)
				}

				if !c.Next() {
					break
				}
			}
		}

	}

	if url == "" {
		return HttpFetch{}, errors.New("Error parsing httpfetch config: parameter url is missing.")
	}

	return HttpFetch{
		ReqUrl: url,
		ReqMethod: method,
		ReqQueryTemplate: query,
		ReqBodyTemplate: body,
		ReqHeaders: headers,
		ResIPExtractor: ipAnalyzer,
		ResTTLExtractor: ttlAnalyzer,
	}, nil
}
