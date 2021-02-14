# coredns-httpfetch-plugin

This plugin gets an A record from an HTTP upstream server.


## Usage

To activate the plugin you need to compile CoreDNS with the plugin added
to `plugin.cfg`

```
https://github.com/jijiechen/coredns-httpfetch-plugin
```

Then add it to Corefile:

```
{
   httpfetch {
      # the base URL of the upstream HTTP endpoint 
      url https://httpfetch.example.org/

      # the HTTP method to use when sending query requests
      method POST

      # template for building URL query string when sending query requests
      query "dns_name={{ .DnsName }}"

      # template for building request body when sending query requests
      body "{{ (dict \"dns_name\" .DnsName) | toJson }}"

      # Headers to add when sending query requests
      header "Authorization: Bearer XXX"
      header "Content-Type: application/json"
      
      # Template to extract resolved IP address from the response
      # If omiited, the whole response body will be used as the IP address
      analyze_ip "{{ (.Body | fromJson).ip_address  }}"

      # Template to extract ttl (time to last in seconds) from the response
      # ttl defaults to 60 seconds 
      analyze_ttl "{{ (.Body | fromJson).ttl  }}"
   }
}
```

Only the `url` config parameter is mandatory.

The following fields support the Go Template syntax and functions provided by the Sprig library are available. See more from its website at http://masterminds.github.io/sprig/

* `query`
* `body`
* `analyze_ip`
* `analyze_ttl`

Available fields for `query` and `body`:
* `DnsName` the currently querying dns name

Available fields for `analyze_ip` and `analyze_ttl`:
* `Body` response body returned from upstream HTTP server  
* `Header` response headers returned from upstream HTTP server 

Trouble escaping your configuration? You may also refer to the Caddyfile documentation at https://caddyserver.com/docs/caddyfile This is because CoreDNS uses Caddy as its underlying engine. 

## Developing locally

You can test the plugin functionality with CoreDNS by adding the following to
`go.mod` in the source code directory of coredns.

```
replace github.com/jijiechen/coredns-httpfetch-plugin => <path-to-you-local-copy>/coredns-httpfetch-plugin
```


