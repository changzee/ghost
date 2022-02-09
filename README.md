# ghost
A  cross-platform hosts file parsing and resolver package for golang.

## Install

    go get -u github.com/changezee/ghost

## Usage

```go
package main

import (
    "fmt"

    "github.com/changezee/ghost"
)

func main() {
	hosts, err := ghost.NewGhost()
	if err != nil {
		panic(err)
	}

	hostfile := hosts.GetHostsFile()
    fmt.Printf("the system host file is: %s\n", hostfile)

	err = ghost.WriteHosts("127.0.0.1", "www.google.com")
	if err != nil {
		panic(err)
	}
	
	ip, err := hosts.Lookup("www.google.com")
    if err != nil {
		panic(err)
    }
	fmt.Printf("www.google.com's ip is: %+v\n", ip)

	hosts := hosts.ReverseLookup("127.0.0.1")
	fmt.Printf("hosts belongs to 127.0.0.1 are: %+v\n", hosts)
}
```

Output:

    the system host file is: /etc/hosts
    www.google.com's ip is: 127.0.0.1
    hosts belongs to 127.0.0.1 are: []string{"127.0.0.1"}

## Supported Operating Systems

Tested and verified working on:

* Mac OS X

## Unit-tests

Running the unit-tests is straightforward and standard:

    go test


## License

Permissive MIT license.
