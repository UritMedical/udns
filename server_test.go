package udns

func ExampleRegister() {
	register, err := Register(
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
}
