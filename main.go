package main

import (
	"fmt"
	"hub.ahiho.com/ahiho/squirrel-srv/internal/vpn"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/version"
	"log"
	"os"
)

func main() {
	log.Printf(
		"Starting the service...\ncommit: %s, build time: %s, release: %s",
		version.Commit, version.BuildTime, version.Release,
	)

	if err := vpn.RunServer(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
