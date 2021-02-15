

package httpfetch

import (
	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n  }`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `httpfetch {}`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}


func TestSetupWithUrl(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n }`)
	httpFetch, _ := newHttpFetch(c)
	assert.Equal(t, "example.org", httpFetch.ReqUrl, "Url not set properly")
	assert.Equal(t, "GET", httpFetch.ReqMethod, "Http method did not default to GET")
	assert.Equal(t, "dns_name={{ .DnsName }}", httpFetch.ReqQueryTemplate, "Http method did not default to GET")
}

func TestSetupWithQueryTemplate(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n query "domain={{ .DnsName }}"\n }`)
	httpFetch, _ := newHttpFetch(c)
	assert.Equal(t, "domain={{ .DnsName }}", httpFetch.ReqQueryTemplate, "Query template not set properly")
}


// see caddyfile syntax at https://caddyserver.com/docs/caddyfile/concepts
// it's strange that we can't use backtick ` here
func TestSetupWithParameterEscaping(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch {    httpfetch {
      url https://httpfetch.example.org/
      method POST
      query "dns_name={{ .DnsName }}"
      body "{{ (dict \"dns_name\" .DnsName) | toJson }}"
      header "Authorization: Bearer XXX"
      header "Content-Type: application/json"
      
      analyze_ip "{{ (.Body | fromJson).ip_address  }}"
      analyze_ttl "{{ (.Body | fromJson).ttl  }}"
   }`)

	httpFetch, _ := newHttpFetch(c)
	assert.Equal(t, "https://httpfetch.example.org/", httpFetch.ReqUrl, "Url not set properly")
	assert.Equal(t, "POST", httpFetch.ReqMethod, "Http method did not default to GET")

assert.Equal(t, `dns_name={{ .DnsName }}`, httpFetch.ReqQueryTemplate, "Query template was not processed correctly")
	assert.Equal(t, `{{ (dict "dns_name" .DnsName) | toJson }}`, httpFetch.ReqBodyTemplate, "Body template was not processed correctly")
	assert.Equal(t, 2, len(httpFetch.ReqHeaders), "Request header not processed correctly")
	assert.Equal(t, `Authorization: Bearer XXX`, httpFetch.ReqHeaders[0], "Request header not processed correctly")
	assert.Equal(t, `{{ (.Body | fromJson).ip_address  }}`, httpFetch.ResIPExtractor, "Request header not processed correctly")
}
