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
      query dns_name=%s
      body '{"dns_name":"%s"}'
      header Authorization: Bearer XXX
      header Content-Type: application/json
      
      analyze_ip '.responseText | json | access ".obj[0].ip"'
      analyze_ttl '.responseText | json | access ".obj[0].ttl"'
   }
}
```

Only the `url` config parameter is mandatory.

## Developing locally

You can test the plugin functionality with CoreDNS by adding the following to
`go.mod` in the source code directory of coredns.

```
replace github.com/jijiechen/coredns-httpfetch-plugin => <path-to-you-local-copy>/coredns-httpfetch-plugin
```


