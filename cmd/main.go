package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Urit-Mediacal/udns"
)

var (
	wait, port                       int
	register, search                 bool
	instance, service, host, ip, key string
	keys                             []string
)

func init() {
	flag.IntVar(&wait, "wait", 300, "运行参数, 自动关闭时间, -1 无限制")
	flag.IntVar(&port, "port", 8000, "注册参数, 服务端口")
	flag.BoolVar(&register, "reg", true, "进行注册?")
	flag.BoolVar(&search, "search", true, "进行查询?")

	flag.StringVar(&instance, "name", "", "注册/查询, 服务名称")
	flag.StringVar(&service, "service", "http", "注册/查询, 服务类型")
	flag.StringVar(&host, "host", "", "注册/查询, 主机名(可选).")
	flag.StringVar(&ip, "ip", "", "注册, 主机ip(可选).")
	flag.StringVar(&key, "tags", "", "注册/查询, 服务tag(可选, 非严格模式).")
	flag.Parse()
	keys = strings.Split(key, ",")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	if wait != -1 {
		ctx, cancel = context.WithTimeout(ctx, time.Second*time.Duration(wait))
	}
	defer cancel()

	if register {
		go func() {
			for {
				_, err := udns.Register(
					udns.SetInstance(instance),
					udns.SetService(service),
					udns.SetHost(host),
					udns.SetPort(port),
					udns.SetIPs(ip),
					udns.SetKey(keys...),
				)
				if err != nil {
					log.Fatal(err)
				}
				time.Sleep(time.Second * 10)
			}
		}()
	}

	if search {
		client := udns.NewListener(
			udns.FindInstance(instance),
			udns.FindService(service),
			udns.FindHost(host),
			udns.FindKey(keys...),
		)
		go func() {
			for {
				outputEntry(<-client.Entries)
			}
		}()
		go client.Browser()
	}
	<-ctx.Done()
	fmt.Println("shutdown...")
}

func outputEntry(e *udns.ServiceEntry) {
	title := fmt.Sprintf("%v%v%v\n", strings.Repeat("-", 5), e.Instance, strings.Repeat("-", 5))
	footer := fmt.Sprintf("%v%v%v\n", strings.Repeat("-", 5), strings.Repeat("-", len(e.Instance)), strings.Repeat("-", 5))
	fmt.Printf(title)
	fmt.Printf("name: %v\n", e.Instance)
	fmt.Printf("type: %v\n", e.Service)
	fmt.Printf("host: %v\n", e.HostName)
	fmt.Printf("port: %v\n", e.Port)
	for i := range e.AddrIPv4 {
		fmt.Printf("ip%v:%v\n", i+1, e.AddrIPv4[i].String())
	}
	fmt.Printf(footer)
}
