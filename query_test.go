

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
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	want := "10.0.0.2"
	got := query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100)
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}

}

func TestNoSuchHost(t *testing.T) {

	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(
		200).BodyString(`{"count":0,"next":null,"previous":null,"results":[]}`)

	want := ""
	got := query("https://example.org/api/ipam/ip-addresses", "mytoken", "NoSuchHost", time.Millisecond*100)
	if got != want {
		t.Fatalf("Expected empty string but got %s", got)
	}

}

func TestLocalCache(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	ip_address := ""

	got := query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100)

	item, err := localCache.Get("my_host")
	if err == nil {
		ip_address = item.Value().(string)
	}

	assert.Equal(t, got, ip_address, "local cache item didn't match")

}

func TestLocalCacheExpiration(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	query("https://example.org/api/ipam/ip-addresses", "mytoken", "my_host", time.Millisecond*100)
	<-time.After(101 * time.Millisecond)
	item, err := localCache.Get("my_host")
	if err != nil {
		t.Fatalf("Expected errors, but got: %v", item)
	}
}
