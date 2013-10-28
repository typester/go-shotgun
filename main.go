package main

import (
	"flag"
	"fmt"
	"github.com/typester/go-shotgun/shotgun"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	portMap = flag.String("map", "3000:5000", "shotogun to app, port mapping")
	timeout = flag.Uint("timeout", 10, "timeout sec to wait app's reload")

	portMapRule = regexp.MustCompile(`^([1-9][0-9]*?):([1-9][0-9]*?)$`)
)

func main() {
	flag.Parse()

	ports := portMapRule.FindStringSubmatch(*portMap)
	if len(ports) < 3 {
		fmt.Fprintf(os.Stderr, "Invalid map format: %s\n", *portMap)
		os.Exit(1)
	}

	src, err := strconv.Atoi(ports[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid map format: %s\n", *portMap)
		os.Exit(1)
	}
	dest, err := strconv.Atoi(ports[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid map format: %s\n", *portMap)
		os.Exit(1)
	}

	timeout := time.Duration(*timeout) * time.Second

	shotgun, err := shotgun.New(uint(src), uint(dest), flag.Args(), ".")
	shotgun.SetTimeout(timeout)
	shotgun.Run()
}
