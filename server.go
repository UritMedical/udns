package udns

import (
	"fmt"
	"github.com/grandcat/zeroconf"
	"net"
	"os"
	"strings"
)

type Server struct {
	server *zeroconf.Server
}

// Register a service by given ServiceOpt. This call will take the system's hostname and lookup IP by that hostname.
func Register(opts ...RegOpt) (server *Server, err error) {
	opt := initRegOpts(opts)
	s := Server{}
	if opt.host != "" {
		s.server, err = zeroconf.RegisterProxy(opt.instance, opt.service, opt.domain, opt.port, opt.host, opt.ips, opt.text, opt.NetInfs)
	} else {
		s.server, err = zeroconf.Register(opt.instance, opt.service, opt.domain, opt.port, opt.text, opt.NetInfs)
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Shutdown closes all udp connections and unregisters the service
func (s *Server) Shutdown() {
	if s.server != nil {
		s.server.Shutdown()
	}
}

//////////////// RegisterOptionsSession ////////////////////////

type (
	registerOptions struct {
		instance string
		service  string
		domain   string
		port     int
		text     []string
		NetInfs  []net.Interface
		host     string
		ips      []string
	}
	RegOpt func(opts *registerOptions)
)

func initRegOpts(opts []RegOpt) *registerOptions {
	options := &registerOptions{
		instance: "Default",
		service:  "_urit",
		domain:   "local",
		port:     8000,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// SetInstance - set instance name
func SetInstance(instance string) RegOpt {
	return func(opts *registerOptions) {
		if instance != "" {
			instance = strings.TrimSpace(instance)
			instance = strings.ReplaceAll(instance, " ", "_")
			opts.instance = instance
		} else {
			opts.instance = fmt.Sprintf("%s-%v", opts.instance, os.Getpid())
		}
	}
}

// SetService - 设置服务类型, 用点分隔
func SetService(service string) RegOpt {
	return func(opts *registerOptions) {
		if service != "" {
			opts.service = fmt.Sprintf("%s.%s", trimDot(opts.service), trimDot(service))
		}
	}
}

func SetPort(port int) RegOpt {
	return func(opts *registerOptions) {
		if port != 0 {
			opts.port = port
		}
	}
}

// SetKey 设置tags
func SetKey(keys ...string) RegOpt {
	return func(opts *registerOptions) {
		opts.text = keys
	}
}

// SetHost set the hostname, if ignore this option register will get from os.Hostname()
func SetHost(host string) RegOpt {
	return func(opts *registerOptions) {
		if host != "" {
			opts.host = host
		}
	}
}

// SetIPs set ip info
func SetIPs(ips ...string) RegOpt {
	return func(opts *registerOptions) {
		opts.ips = ips
	}
}
