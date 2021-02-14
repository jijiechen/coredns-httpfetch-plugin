

package httpfetch

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(`10.0.0.2`)

	want := "10.0.0.2"
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name=%s"}
	got, _ := query(fetcher,  "my_host")
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}
}

func TestNoSuchHost(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(200).BodyString(``)

	want := ""
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name=%s"}
	got, _ := query(fetcher, "NoSuchHost")
	if got != want {
		t.Fatalf("Expected empty string but got %s", got)
	}

}

func TestLocalCache(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(`10.0.0.2`)

	ipAddress := ""
	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name=%s"}
	got, err := query(fetcher,  "my_host_with_ttl")

	item, err := localCache.Get("my_host_with_ttl")
	if err == nil {
		ipAddress = item.Value().(string)
	}

	assert.Equal(t, ipAddress, got, "local cache item didn't match")
}

func TestLocalCacheExpiration(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_expired_host"}).Reply(
		200).BodyString(`10.0.0.25`)

	fetcher := HttpFetch{ReqUrl: "https://example.org/api/ipam/ip-addresses/", ReqQueryTemplate: "dns_name=%s"}
	query(fetcher, "my_expired_host")
	<-time.After(61 * time.Millisecond)
	item, err := localCache.Get("my_expired_host")
	if err != nil {
		t.Fatalf("Expected errors, but got: %v", item)
	}
}

func TestQueryWithHeader(t *testing.T) {
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


func TestQueryWithIPExtractor(t *testing.T) {
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

