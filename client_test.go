package udns

import "fmt"

func ExampleNewListener() {
	client := NewListener(
		FindInstance("My App"),
		FindService("http.tcp"),
		FindHost("My-PC"),
		FindKey("my app"),
	)
	defer client.Shutdown()
	go func() {
		for {
			fmt.Println(<-client.Entries)
		}
	}()
	client.Browser()
}
