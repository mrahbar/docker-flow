package main

// TODO: Test

import (
	"fmt"
	"log"
	"strings"
)

func init() {
	log.SetPrefix(">> Docker Flow: ")
	log.SetFlags(0)
}

var logFatal = log.Fatal
var logPrintln = log.Println

func main() {
	flow := getFlow()
	stats := getStats()

	opts, err := GetOpts()
	if err != nil {
		logFatal(err)
	}
	dc := getDockerCompose()

	for _, step := range opts.Flow {
		switch strings.ToLower(step) {
		case STATS_SERVICES:
			resp, err := stats.Services(opts)
			if err != nil {
				logFatal(err)
			} else {
				logPrintln(resp)
			}
		case STATS_NODES:
			resp, err := stats.Nodes(opts, opts.Target)
			if err != nil {
				logFatal(err)
			} else {
				logPrintln(resp)
			}
		case FLOW_DEPLOY:
			if err := flow.Deploy(opts, dc); err != nil {
				logFatal(err)
			}
		case FLOW_SCALE:
			if !deployed {
				logPrintln(fmt.Sprintf("Scaling (%s)...", opts.CurrentTarget))
				fmt.Println(opts.Flow)
				if err := flow.Scale(opts, dc, opts.CurrentTarget, true); err != nil {
					logFatal(err)
				}
			}
		case FLOW_STOP_OLD:
			if err := flow.StopOld(opts, dc); err != nil {
				logFatal(err)
			}
		case FLOW_STOP_APP:
			if err := flow.StopApp(opts, dc); err != nil {
				logFatal(err)
			}
		case FLOW_STOP_COLOR:
			if err := flow.StopColor(opts, dc); err != nil {
				logFatal(err)
			}
		case FLOW_PROXY:
			if err := flow.Proxy(opts, haProxy); err != nil {
				logFatal(err)
			}
		}
	}
}
