

package httpfetch

import (
	"context"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"io"
	"net"
	"os"
	"strings"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("httpfetch")

// method: GET/POST/PUT/...
// URL: ends with /
// queryTemplate: does not contains /
// header: may have multiple headers...

type Httpfetch struct {
	ReqMethod        string
	ReqUrl           string
	ReqHeaders       []string
	ReqQueryTemplate string
	ReqBodyTemplate  string

	ResIPExtractor  string // json(), .response
	ResTTLExtractor string

	Next          plugin.Handler
}

func (httpFetch Httpfetch) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	answers := []dns.RR{}
	state := request.Request{W: w, Req: r}

	ipAddress, err := query(&httpFetch, strings.TrimRight(state.QName(), "."))

	if err != nil {
		log.Warning("Error fetching dns from upstream: %v", err)
		return plugin.NextOrFailure(httpFetch.Name(), httpFetch.Next, ctx, w, r)
	}
	if len(ipAddress) == 0 {
		return plugin.NextOrFailure(httpFetch.Name(), httpFetch.Next, ctx, w, r)
	}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
	rec.A = net.ParseIP(ipAddress)
	answers = append(answers, rec)
	m := new(dns.Msg)
	m.Answer = answers
	m.SetReply(r)
	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (httpFetch Httpfetch) Name() string { return "httpfetch" }

// Make out a reference to os.Stdout so we can easily overwrite it for testing.
var out io.Writer = os.Stdout
