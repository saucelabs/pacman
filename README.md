# pacman

This package provides a pure Go [pac](https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_(PAC)_file) parser based on [goja](https://github.com/dop251/goja)

## Usage

### Get package

```bash
go get github.com/saucelabs/pacman
```

### Import package

```go
package main

import (
 "fmt"
 "log"

 "github.com/saucelabs/pacman"
)

// Example of PAC.
var scripts = `
  function FindProxyForURL(url, host) {
    if (isPlainHostName(host)) return DIRECT;
    else return "PROXY 127.0.0.1:8080; PROXY 127.0.0.1:8081; DIRECT";
  }
`

func main() {
 	pac, err := pacman.New(scripts)
	if err != nil {
		log.Fatall(err)
	}

 	r, err := pac.FindProxyForURL("http://www.example.com/")
	 if err != nil {
		log.Fatall(err)
	}

 	fmt.Println(r) // returns PROXY 127.0.0.1:8080; PROXY 127.0.0.1:8081; DIRECT

 	// Get issues request via a list of proxies and returns at the first request that succeeds
 	resp, err := pac.Get("http://www.example.com/")
	 if err != nil {
		log.Fatall(err)
	}

 	fmt.Println(resp.Status)
}
```
