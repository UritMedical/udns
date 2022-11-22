package udns

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/grandcat/zeroconf"
)

type (
	Resolver struct {
		opt     *clientOptions
		Entries chan *ServiceEntry
		entries chan *zeroconf.ServiceEntry
	}
	ServiceEntry struct {
		Instance string   `json:"name"`   // Instance name (e.g. "My web page")
		Service  string   `json:"type"`   // Service name (e.g. _http._tcp.)
		Domain   string   `json:"domain"` // If blank, assumes "local"
		HostName string   `json:"hostname"`
		TTL      uint32   `json:"ttl"`
		AddrIPv4 []net.IP `json:"ips"`
	}
)

func NewResolver(name string, opts ...ClientOpt) *Resolver {
	r := Resolver{
		opt:     initClientOpts(opts),
		Entries: make(chan *ServiceEntry),
		entries: make(chan *zeroconf.ServiceEntry),
	}
	if name != "" {
		r.opt.instance = name
	}
	return &r
}

func (r *Resolver) Browser() {
	// 启动回传协程
	go func(results <-chan *zeroconf.ServiceEntry) {
		for zEntry := range results {
			zEntry.HostName = trimDot(strings.ReplaceAll(zEntry.HostName, zEntry.Domain, ""))
			zEntry.Service = trimDot(strings.Trim(zEntry.Service, "_urit"))

			entry := &ServiceEntry{
				Instance: zEntry.Instance,
				Service:  zEntry.Service,
				Domain:   zEntry.Domain,
				HostName: zEntry.HostName,
				TTL:      zEntry.TTL,
			}

			// 过滤
			if r.opt.instance != "" {
				if !strings.EqualFold(zEntry.Instance, r.opt.instance) {
					continue
				}
			}
			if r.opt.host != "" {
				if !strings.EqualFold(zEntry.HostName, r.opt.host) {
					continue
				}
			}

			for _, ipv4 := range zEntry.AddrIPv4 {
				conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", ipv4.String(), zEntry.Port))
				if err != nil {
					continue
				}
				buf := [512]byte{}
				n, err := conn.Read(buf[:])
				if err != nil {
					continue
				}

				if strings.EqualFold(string(buf[:n]), TCP_MSG) {
					entry.AddrIPv4 = append(entry.AddrIPv4, ipv4)
				}
			}
			r.Entries <- entry
		}
	}(r.entries)

	// 开始查询
	resolver, err := zeroconf.NewResolver(zeroconf.SelectIPTraffic(zeroconf.IPv4))
	if err == nil {
		err = resolver.Browse(context.Background(), r.opt.service, r.opt.domain, r.entries)
	}
	if err != nil {
		fmt.Println("browse service failed, err: ", err)
	}
}

func (r *Resolver) Shutdown() {
	close(r.entries)
	close(r.Entries)
}

//////////////// ClientOptionsSession ////////////////////////

type (
	clientOptions struct {
		instance string
		service  string
		domain   string
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

func FindService(service string) ClientOpt {
	return func(opts *clientOptions) {
		if service != "" {
			opts.service = fmt.Sprintf("%s.%s", trimDot(opts.service), trimDot(service))
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
