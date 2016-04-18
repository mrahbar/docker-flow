package main

import (
)

type Statsable interface {
	Services(opts Opts) (string, error)
	Nodes(opts Opts, target string) (string, error)
}

const STATS_SERVICES = "services"
const STATS_NODES = "nodes"

type Stats struct{}

var stats Statsable = Stats{}
func getStats() Statsable {
	return stats
}

func (m Stats) Services(opts Opts) (string, error) {
	sc := getServiceDiscovery()

	services, err := sc.GetServices(opts.ServiceDiscoveryAddress)
	if err != nil {
		return "", err
	}

	return services, err
}

func (m Stats) Nodes(opts Opts, target string) (string, error) {
	sc := getServiceDiscovery()

	nodes, err := sc.GetNodes(opts.ServiceDiscoveryAddress, target)
	if err != nil {
		return "", err
	}

	return nodes, err
}
