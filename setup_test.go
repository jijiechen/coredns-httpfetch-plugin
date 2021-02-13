

package httpfetch

import (
	//"fmt"
	"testing"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/stretchr/testify/assert"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n token foobar\n localCacheDuration 10s }`)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}

	c = caddy.NewTestController("dns", `httpfetch {}`)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}

func TestSetupWithWrongDuration(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n token foobar\n localCacheDuration Wrong }`)
	_, err := newHttpFetch(c)
	assert.Error(t, err, "Expected error")
}

func TestSetupWithDuration(t *testing.T) {
	c := caddy.NewTestController("dns", `httpfetch { url example.org\n token foobar\n localCacheDuration 10s }`)
	nb, _ := newHttpFetch(c)
	assert.Equal(t, nb.CacheDuration, time.Second*10, "Duration not set properly")
}
