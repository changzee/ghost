package ghost

import (
	"bufio"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/fsnotify/fsnotify"
)

type Ghost struct {
	hostsMap *sync.Map

	hosts   string
	watcher *fsnotify.Watcher
}

func parseHosts(hosts string) (*sync.Map, error) {
	contents, err := ioutil.ReadFile(hosts)
	if err != nil {
		return nil, err
	}

	var hostsMap = sync.Map{}
	lines := strings.Split(strings.Trim(string(contents), " \t\r\n"), "\n")
	for _, line := range lines {
		line = strings.Replace(strings.Trim(line, " \t"), "\t", " ", -1)
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}
		pieces := strings.SplitN(line, " ", 2)
		if len(pieces) > 1 && len(pieces[0]) > 0 {
			if names := strings.Fields(pieces[1]); len(names) > 0 {
				for _, name := range names {
					hostsMap.Store(name, pieces[0])
				}
			}
		}
	}
	return &hostsMap, nil
}

// WriteHosts write a ip-host kv to host file.
func (g *Ghost) WriteHosts(ip string, hostname string) error {
	file, err := os.OpenFile(g.hosts, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("\n" + ip + "_" + hostname)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// Lookup takes a host and returns a ip address.
func (g *Ghost) Lookup(hostname string) (string, error) {
	unsafePointer := unsafe.Pointer(g.hostsMap)
	hostsMap := (*sync.Map)(atomic.LoadPointer(&unsafePointer))

	value, ok := hostsMap.Load(hostname)
	if !ok {
		return "", errors.New("not exist")
	}
	return value.(string), nil
}

// ReverseLookup takes an IP address and returns a slice of matching hosts file
// entries.
func (g *Ghost) ReverseLookup(ip string) []string {
	var hosts = make([]string, 0, 32)
	g.hostsMap.Range(func(key, value interface{}) bool {
		if value.(string) == ip {
			hosts = append(hosts, key.(string))
		}
		return true
	})
	return hosts
}

func (g *Ghost) watchChange() {
	// init fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	// add host file to watcher
	err = watcher.Add(g.hosts)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name == g.hosts {
						hostsMap, err := parseHosts(g.hosts)
						if err != nil {
							log.Println("error:", err)
							continue
						}
						unsafePointer := unsafe.Pointer(g.hostsMap)
						atomic.StorePointer(&unsafePointer, unsafe.Pointer(hostsMap))
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
}

func NewGhost(host ...string) (*Ghost, error) {
	if len(host) == 0 {
		host = []string{HostsFile}
	}

	hostsMap, err := parseHosts(host[0])
	if err != nil {
		return nil, err
	}
	var ghost = Ghost{
		hosts:    host[0],
		hostsMap: hostsMap,
		watcher:  nil,
	}
	go ghost.watchChange()
	return &ghost, nil
}
