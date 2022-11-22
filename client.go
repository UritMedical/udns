package udns

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"strings"
)

type (
	Client struct {
		opt      *clientOptions
		resolver *zeroconf.Resolver
		Entries  chan *ServiceEntry
		entries  chan *zeroconf.ServiceEntry
	}
	ServiceEntry struct {
		zeroconf.ServiceEntry
	}
)

func NewListener(opts ...ClientOpt) *Client {
	opt := initClientOpts(opts)
	return &Client{
		opt:     opt,
		Entries: make(chan *ServiceEntry),
		entries: make(chan *zeroconf.ServiceEntry),
	}
}

func (c *Client) Browser() {
	// 启动回传协程
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Printf("get %v\n", entry.Instance)
			// 预处理
			entry.HostName = trimDot(strings.ReplaceAll((entry.HostName), entry.Domain, ""))
			entry.Service = trimDot(strings.Trim(entry.Service, "_urit"))

			// 过滤
			if c.opt.instance != "" {
				if entry.Instance != c.opt.instance {
					continue
				}
			}
			if c.opt.host != "" {
				if entry.HostName != c.opt.host {
					continue
				}
			}
			if c.opt.text != nil {
				var found bool
				for i := range entry.Text {
					for j := range c.opt.text {
						if entry.Text[i] == c.opt.text[j] {
							found = true
							break
						}
					}
				}
				if !found {
					continue
				}
			}

			c.Entries <- &ServiceEntry{*entry}
		}
	}(c.entries)

	// 开始查询
	resolver, err := zeroconf.NewResolver(zeroconf.SelectIPTraffic(zeroconf.IPv4))
	if err == nil {
		err = resolver.Browse(context.Background(), c.opt.service, c.opt.domain, c.entries)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Client) Shutdown() {
	close(c.entries)
	close(c.Entries)
}

//////////////// ClientOptionsSession ////////////////////////

type (
	clientOptions struct {
		instance string
		service  string
		domain   string
		text     []string
		host     string
	}
	ClientOpt func(opts *clientOptions)
)

func initClientOpts(opts []ClientOpt) *clientOptions {
	options := &clientOptions{
		instance: "",
		service:  "_urit",
		domain:   "local",
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func FindInstance(instance string) ClientOpt {
	return func(opts *clientOptions) {
		if instance != "" {
			instance = strings.TrimSpace(instance)
			instance = strings.ReplaceAll(instance, " ", "_")
			opts.instance = instance
		}
	}
}

func FindService(service string) ClientOpt {
	return func(opts *clientOptions) {
		if service != "" {
			opts.service = fmt.Sprintf("%s.%s", trimDot(opts.service), trimDot(service))
		}
	}
}

func FindKey(keys ...string) ClientOpt {
	return func(opts *clientOptions) {
		if keys != nil {
			opts.text = keys
		}
	}
}

func FindHost(host string) ClientOpt {
	return func(opts *clientOptions) {
		if host != "" {
			opts.host = host
		}
	}
}
