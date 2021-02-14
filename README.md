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
      url https://httpfetch.example.org/
      method POST
      query "dns_name={{ .DnsName }}"
      body "{{ (dict \"dns_name\" .DnsName) | toJson }}"
      header Authorization: Bearer XXX
      header Content-Type: application/json
      
      analyze_ip "{{ (.ResponseText | fromJson).ip_address  }}"
      analyze_ttl "{{ (.ResponseText | fromJson).ttl  }}"
   }
}
```

Only the `url` config parameter is mandatory.

The following fields support the Go Template syntax and functions provided by the Sprig library are available. See more from its website at http://masterminds.github.io/sprig/
* query
* body
* analyze_ip
* analyze_ttl

Trouble escaping your configuration? you may also refer to the Caddyfile documentation at https://caddyserver.com/docs/caddyfile This is because CoreDNS uses Caddy as its underlying engine. 

## Developing locally

You can test the plugin functionality with CoreDNS by adding the following to
`go.mod` in the source code directory of coredns.

```
replace github.com/jijiechen/coredns-httpfetch-plugin => <path-to-you-local-copy>/coredns-httpfetch-plugin
```


