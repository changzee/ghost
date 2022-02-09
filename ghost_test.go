package ghost

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGhost(t *testing.T) {
	hosts, err := ioutil.TempFile("", "*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hosts.Close() }()
	defer func() { _ = os.Remove(hosts.Name()) }()

	ghost, err := NewGhost(hosts.Name())
	if err != nil {
		t.Fatal(err)
	}

	_, err = ghost.Lookup("www.baidu.com")
	if err == nil {
		t.Fatal("ghost.Lookup(\"www.baidu.com\") expect not exist error")
	}

	err = ghost.WriteHosts("127.0.0.1", "www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)
	ip, err := ghost.Lookup("www.baidu.com")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "127.0.0.1", ip)

	host := ghost.ReverseLookup("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, host, []string{"www.baidu.com"})

}
