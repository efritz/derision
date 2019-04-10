package main

import (
	"github.com/efritz/derision/internal/server"
	"github.com/efritz/nacelle"
)

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterProcess(
		server.NewServer(),
		nacelle.WithProcessName("server"),
	)

	return nil
}

func main() {
	nacelle.NewBootstrapper("derision", setup).BootAndExit()
}
