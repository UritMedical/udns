package udns

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

type Register struct {
	opt    *registerOptions
	server *zeroconf.Server
}

// NewRegister a service by given ServiceOpt. This call will take the system's hostname and lookup IP by that hostname.
func NewRegister(name string, opts ...RegisterOption) (server *Register, err error) {
	r := Register{
		opt: initRegOpts(opts),
	}
	if name != "" {
		r.opt.instance = strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
	}
	if r.opt.tcp {
		go r.tcpGoroutine()
	}
	time.Sleep(time.Millisecond * 1000)
	go r.register()
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *Register) register() {
	for {
		var err error
		if r.opt.host != "" {
			r.server, err = zeroconf.RegisterProxy(r.opt.instance, r.opt.service, r.opt.domain, r.opt.port, r.opt.host, r.opt.ips, r.opt.text, r.opt.netIfaces)
		} else {
			r.server, err = zeroconf.Register(r.opt.instance, r.opt.service, r.opt.domain, r.opt.port, r.opt.text, r.opt.netIfaces)
		}
		time.Sleep(time.Millisecond * 5000)
		if err != nil {
			log.Println("register service failed, err: ", err)
		}
	}
}

// Shutdown closes all udp connections and unregisters the service
func (r *Register) Shutdown() {
	if r.server != nil {
		r.server.Shutdown()
	}
}

// tcpGoroutine start a tcp server for client check connections
func (r *Register) tcpGoroutine() {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", r.opt.port))
	if err != nil {
		log.Println("start tcp failed, err: ", err)
		return
	}
	r.opt.port = listen.Addr().(*net.TCPAddr).Port
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("accept tcp failed, err: ", err)
			continue
		}
		go func(net.Conn) {
			defer conn.Close()
			conn.Write([]byte(TCP_MSG))
		}(conn)
	}
}

//////////////// RegisterOptionsSession ////////////////////////

type (
	registerOptions struct {
		instance  string
		service   string
		domain    string
		port      int
		text      []string
		netIfaces []net.Interface
		host      string
		ips       []string
		tcp       bool
	}
	RegisterOption func(opts *registerOptions)
)

func initRegOpts(opts []RegisterOption) *registerOptions {
	options := &registerOptions{
		instance: fmt.Sprintf("default-%v", os.Getpid()),
		service:  "_urit",
		domain:   "local",
		tcp:      false,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// SetService - 设置服务类型, 用点分隔
func SetService(service string) RegisterOption {
	return func(opts *registerOptions) {
		if service != "" {
			opts.service = fmt.Sprintf("%s.%s", trimDot(opts.service), trimDot(service))
		}
	}
}

func SetPort(port int) RegisterOption {
	return func(opts *registerOptions) {
		if port != 0 {
			opts.port = port
		}
	}
}

// SetKey 设置tags
func SetKey(keys ...string) RegisterOption {
	return func(opts *registerOptions) {
		opts.text = keys
	}
}

// SetHost set the hostname, if ignore this option register will get from os.Hostname()
func SetHost(host string) RegisterOption {
	return func(opts *registerOptions) {
		if host != "" {
			opts.host = host
		}
	}
}

// SetIPs set ip info
func SetIPs(ips ...string) RegisterOption {
	return func(opts *registerOptions) {
		opts.ips = ips
	}
}

// SetTCP setup tcp service
func SetTCP(ok bool) RegisterOption {
	return func(opts *registerOptions) {
		opts.tcp = ok
	}
}
