

package httpfetch

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
	"text/template"
	"time"
)

func TestQuery(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(`10.0.0.2`)

	want := "10.0.0.2"
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name={{ .DnsName }}"}
	got, _ := query(fetcher,  "my_host")
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}
}

func TestNoSuchHost(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(200).BodyString(``)

	want := ""
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name={{ .DnsName }}"}
	got, _ := query(fetcher, "NoSuchHost")
	if got != want {
		t.Fatalf("Expected empty string but got %s", got)
	}

}

func TestLocalCache(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(`10.0.0.2`)

	ipAddress := ""
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name={{ .DnsName }}"}
	got, err := query(fetcher,  "my_host_with_ttl")

	item, err := localCache.Get("my_host_with_ttl")
	if err == nil {
		ipAddress = item.Value().(string)
	}

	assert.Equal(t, ipAddress, got, "local cache item didn't match")
}

func TestLocalCacheExpiration(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_expired_host"}).Reply(
		200).BodyString(`10.0.0.25`)

	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name={{ .DnsName }}"}
	query(fetcher, "my_expired_host")
	<-time.After(61 * time.Millisecond)
	item, err := localCache.Get("my_expired_host")
	if err != nil {
		t.Fatalf("Expected errors, but got: %v", item)
	}
}

func TestQueryWithHeader(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() //
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchHeader("X-Token", "xyz").Reply(
		200).BodyString(`10.0.0.2`)

	want := "10.0.0.2"
	fetcher := HttpFetch{
		ReqUrl: "https://example.org/api/ipam/ip-addresses/",
		ReqHeaders: []string{"X-Token: xyz"}}
	got, _ := query(fetcher,  "my_host_with_header")
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}
}

func TestQueryWithBody(t *testing.T) {
	resetTemplateCache()
	defer gock.Off()

	var requestedBodyString string
	gock.New("https://example.org/api/ipam/ip-addresses/").AddMatcher(func(httpReq *http.Request, req *gock.Request) (bool, error) {
		body, _ := ioutil.ReadAll(httpReq.Body)
		requestedBodyString = string(body)
		return len(requestedBodyString) > 0, nil
	}).Reply(200).BodyString(`10.0.0.2`)

	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqBodyTemplate: "{{ (dict \"dns_name\" .DnsName) | toJson }}"}

	got, _ := query(fetcher,  "my_host_with_body")
	assert.Equal(t, "10.0.0.2", got, "The IP did not match expected")
	assert.Equal(t, "{\"dns_name\":\"my_host_with_body\"}",requestedBodyString, "The request body template did not work as expected")
}


func TestQueryWithIPExtractor(t *testing.T) {
	resetTemplateCache()
	defer gock.Off() //
	gock.New("https://example.org/api/ipam/ip-addresses/").Reply(
		200).BodyString(`10.0.0.2/32`)

	want := "10.0.0.2"
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ResIPExtractor: "{{ (.ResponseText | split `/`)._0 }}"  }
	got, _ := query(fetcher,  "my_host_with_ip_extractor")
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}
}


func TestQueryWithTTLExtractor(t *testing.T) {
	resetTemplateCache()
	defer gock.Off()
	gock.New("https://example.org/ip-addresses-with-ttl/").Reply(
		200).BodyString(`{"ip_address": "10.0.0.5", "ttl": 3600}`)

	expectedIP := "10.0.0.5"
	expectedTTL := 3600
	fetcher := HttpFetch{
		ReqUrl: "https://example.org/ip-addresses-with-ttl/",
		ResIPExtractor: "{{ (.ResponseText | fromJson).ip_address  }}",
		ResTTLExtractor: "{{ (.ResponseText | fromJson).ttl }}",
	}

	got, _ := query(fetcher,  "the_cool_host")
	if got != expectedIP {
		t.Fatalf("Expected %s but got %s", expectedIP, got)
	}

	item, err := localCache.Get("the_cool_host")
	if err != nil {
		t.Fatalf("Expected item was not stored correctly in TTL cache: %v", err)
	}

	if item.TTL() < time.Duration(expectedTTL - 3/*allowed skew*/) * time.Second {
		t.Fatalf("Expected TTL %d, but was not correctly set.", expectedTTL)
	}
}

func resetTemplateCache(){
	templateCache = make(map[string]*template.Template)
}