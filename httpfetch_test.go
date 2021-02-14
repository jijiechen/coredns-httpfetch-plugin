

package httpfetch

import (
	"context"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestHttpfetch(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(200).BodyString(`10.0.0.2`)

	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name={{ .DnsName }}"}

	if fetcher.Name() != "httpfetch" {
		t.Errorf("expected plugin name: %s, got %s", "httpfetch", fetcher.Name())
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	r := new(dns.Msg)
	r.SetQuestion("my_host.", dns.TypeA)

	rcode, err := fetcher.ServeDNS(context.Background(), rec, r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rcode != 0 {
		t.Errorf("Expected rcode %v, got %v", 0, rcode)
	}
	IP := rec.Msg.Answer[0].(*dns.A).A.String()

	if IP != "10.0.0.2" {
		t.Errorf("Expected %v, got %v", "10.0.0.2", IP)
	}

}
