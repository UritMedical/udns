package udns

import "fmt"

func ExampleNewListener() {
	client := NewResolver(
		"My App",
		FindService("http.tcp"),
		FindHost("My-PC"),
	)
	defer client.Shutdown()
	go func() {
		for {
			fmt.Println(<-client.Entries)
		}
	}()
	client.Browser()
}
