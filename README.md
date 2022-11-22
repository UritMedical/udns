## MDNS In Pack

## Installation

    go get github.com/Urit-Mediacal/udns

## Usage
### Register a service
```go
register, err := udns.Register(
    SetInstance("My App"),
    SetService("http.tcp"),
    SetHost("My-PC"),
    SetPort(8080),
    SetKey("my app"),
    SetIPs("192.168.1.5"),
)
if err != nil {
    // process
}
defer register.Shutdown()
```
### Discover services on the network
```go
client := NewListener(
    FindInstance("My App"),
    FindService("http.tcp"),
    FindHost("My-PC"),
    FindKey("my app"),
)
go func() {
    for {
        fmt.Println(<-client.Entries)
    }
}()
client.Browser()
defer client.Shutdown()
```
## Build
After download the source code, execute in the source code directory: 

    go build ./cmd/main.go
