package udns

import (
	"context"
	"fmt"
	detect "github.com/allanpk716/go-protocol-detector/pkg"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

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
		for serviceEntry := range results {
			serviceEntry.HostName = trimDot(strings.ReplaceAll(serviceEntry.HostName, serviceEntry.Domain, ""))
			serviceEntry.Service = trimDot(strings.Trim(serviceEntry.Service, "_urit"))

			entry := &ServiceEntry{
				Instance: serviceEntry.Instance,
				Service:  serviceEntry.Service,
				Domain:   serviceEntry.Domain,
				HostName: serviceEntry.HostName,
				TTL:      serviceEntry.TTL,
			}

			// 过滤
			if r.opt.instance != "" {
				if !strings.EqualFold(serviceEntry.Instance, r.opt.instance) {
					continue
				}
			}
			if r.opt.host != "" {
				if !strings.EqualFold(serviceEntry.HostName, r.opt.host) {
					continue
				}
			}

			ipFilterStart := time.Now()
			if r.opt.tcpCheck {
				entry.AddrIPv4 = r.filterIPs(serviceEntry.AddrIPv4, serviceEntry.Port)
			} else {
				entry.AddrIPv4 = r.filterIPsByNetSegment(serviceEntry.AddrIPv4)
			}

			log.Printf("filtering %s ips with %dms", serviceEntry.Instance, time.Since(ipFilterStart).Milliseconds())

			r.Entries <- entry
		}
	}(r.entries)

	// 开始查询
	resolver, err := zeroconf.NewResolver(zeroconf.SelectIPTraffic(zeroconf.IPv4))
	if err == nil {
		err = resolver.Browse(context.Background(), r.opt.service, r.opt.domain, r.entries)
		log.Println("start browse service...")
	}
	if err != nil {
		log.Println("browse service failed, err: ", err)
	}
}

// 通过网段判断
func (r *Resolver) filterIPsByNetSegment(rmtIPs []net.IP) (resp []net.IP) {
	locAddrs, _ := net.InterfaceAddrs()
	for _, rmtIP := range rmtIPs {
		for _, addr := range locAddrs {
			if locIP, ok := addr.(*net.IPNet); ok {
				if locIP.Contains(rmtIP) {
					resp = append(resp, rmtIP)
				}
			}
		}
	}
	return
}

// 通过tcp服务判断
func (r *Resolver) filterIPs(rmtIPs []net.IP, port int) (resp []net.IP) {
	var ips []string
	for i := range rmtIPs {
		ips = append(ips, rmtIPs[i].String())
	}

	resp = []net.IP{}
	output, err := detect.NewScanTools(3, r.opt.tcpTimeout).Scan(detect.Common, detect.InputInfo{
		Host: strings.Join(ips, ","),
		Port: strconv.Itoa(port),
	}, false)
	if err != nil {
		log.Println(err)
	}
	for k := range output.SuccessMapString {
		if ip := net.ParseIP(k); ip != nil {
			resp = append(resp, ip)
		}
	}
	return resp
}

func (r *Resolver) Shutdown() {
	close(r.entries)
	close(r.Entries)
}

//////////////// ClientOptionsSession ////////////////////////

type (
	clientOptions struct {
		instance   string
		service    string
		domain     string
		host       string
		tcpCheck   bool
		tcpTimeout time.Duration
	}
	ClientOpt func(opts *clientOptions)
)

func initClientOpts(opts []ClientOpt) *clientOptions {
	options := &clientOptions{
		instance:   "",
		service:    "_urit",
		domain:     "local",
		tcpCheck:   false,
		tcpTimeout: 5 * time.Second,
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

func TCPCheck(ok bool, timeout int) ClientOpt {
	return func(opts *clientOptions) {
		opts.tcpCheck = ok
		if timeout != 0 {
			opts.tcpTimeout = time.Duration(timeout) * time.Second
		}
	}
}
