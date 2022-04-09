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
	url := "http://www.example.com/"
	pac, err := pacman.New(scripts)
	if err != nil {
		log.Fatal(err)
	}

	r, err := pac.FindProxyForURL(url) // returns PROXY 127.0.0.1:8080; PROXY 127.0.0.1:8081; DIRECT
	if err != nil {
		log.Fatal(err)
	}

	proxies, err := pacman.ParseProxy(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found proxies for %s:\n", url)

	for _, proxy := range proxies {
		fmt.Println(proxy.String())
	}
}
```
