package ghost

import (
	"bufio"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Ghost struct {
	hostsMap atomic.Value

	hosts   string
	watcher *fsnotify.Watcher
}

func parseHosts(hosts string) (map[string]string, error) {
	contents, err := ioutil.ReadFile(hosts)
	if err != nil {
		return nil, err
	}

	var hostsMap = make(map[string]string, 64)
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
					hostsMap[name] = pieces[0]
				}
			}
		}
	}
	return hostsMap, nil
}

// GetHostsFile return the host file that parsing.
func (g *Ghost) GetHostsFile() string {
	return g.hosts
}

// WriteHosts write a ip-host kv to host file.
func (g *Ghost) WriteHosts(ip string, hostname string) error {
	file, err := os.OpenFile(g.hosts, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("\n" + ip + " " + hostname)
	if err != nil {
		return err
	}
	return writer.Flush()
}

// Lookup takes a host and returns a ip address.
func (g *Ghost) Lookup(hostname string) (string, error) {
	hostsMap := g.hostsMap.Load().(map[string]string)
	value, ok := hostsMap[hostname]
	if !ok {
		return "", errors.New("not exist")
	}
	return value, nil
}

// ReverseLookup takes an IP address and returns a slice of matching hosts file
// entries.
func (g *Ghost) ReverseLookup(ip string) []string {
	var hosts = make([]string, 0, 32)
	for key, value := range g.hostsMap.Load().(map[string]string) {
		if value == ip {
			hosts = append(hosts, key)
		}
	}
	return hosts
}

func (g *Ghost) watchChange() {
	// init fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// add host file to watcher
	err = watcher.Add(g.hosts)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer func() { _ = watcher.Close() }()

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

						g.hostsMap.Store(hostsMap)
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

	var ghost = Ghost{
		hosts:   host[0],
		watcher: nil,
	}

	hostsMap, err := parseHosts(host[0])
	if err != nil {
		return nil, err
	}
	ghost.hostsMap.Store(hostsMap)
	ghost.watchChange()
	return &ghost, nil
}
