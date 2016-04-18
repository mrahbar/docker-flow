package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"os"
	"strconv"
	"strings"
)

const dockerFlowPath = "docker-flow.yml"
const dockerComposePath = "docker-compose.yml"

var getWd = os.Getwd
var parseYml = ParseYml
var parseEnvVars = ParseEnvVars
var parseArgs = ParseArgs
var processOpts = ProcessOpts

type Opts struct {
	Host                    string   `short:"H" long:"host" description:"Docker daemon socket to connect to. If not specified, DOCKER_HOST environment variable will be used instead."`
	CertPath                string   `long:"cert-path" description:"Docker certification path. If not specified, DOCKER_CERT_PATH environment variable will be used instead." yaml:"cert_path" envconfig:"cert_path"`
	ComposePath             string   `short:"f" long:"compose-path" value-name:"docker-compose.yml" description:"Path to the Docker Compose configuration file." yaml:"compose_path" envconfig:"compose_path"`
	BlueGreen               bool     `short:"b" long:"blue-green" description:"Perform blue-green deployment." yaml:"blue_green" envconfig:"blue_green"`
	Color               	string   `short:"C" long:"color" description:"Specify either blue or green directly."`
	Target                  string   `short:"t" long:"target" description:"Docker Compose target."`
	SideTargets             []string `short:"T" long:"side-target" description:"Side or auxiliary Docker Compose targets. Multiple values are allowed." yaml:"side_targets"`
	PullSideTargets         bool     `short:"S" long:"pull-side-targets" description:"Pull side or auxiliary targets." yaml:"pull_side_targets" envconfig:"pull_side_targets"`
	Project                 string   `short:"p" long:"project" description:"Docker Compose project. If not specified, the current directory will be used instead."`
	ServiceDiscoveryAddress string   `short:"c" long:"consul-address" description:"The address of the Consul server." yaml:"consul_address" envconfig:"consul_address"`
	Scale                   string   `short:"s" long:"scale" description:"Number of instances to deploy. If the value starts with the plus sign (+), the number of instances will be increased by the given number. If the value begins with the minus sign (-), the number of instances will be decreased by the given number." yaml:"scale" envconfig:"scale"`
	Flow                    []string `short:"F" long:"flow" description:"The actions that should be performed as the flow. Multiple values are allowed.\ndeploy: Deploys a new release\nscale: Scales currently running release\nstop-old: Stops the old release\nproxy: Reconfigures the proxy\n" yaml:"flow" envconfig:"flow"`
	ProxyHost               string   `long:"proxy-host" description:"The host of the proxy. Visitors should request services from this domain. Docker Flow uses it to request reconfiguration when a new service is deployed or an existing one is scaled. This argument is required only if the proxy flow step is used." yaml:"proxy_host" envconfig:"proxy_host"`
	ProxyDockerHost         string   `long:"proxy-docker-host" description:"Docker daemon socket of the proxy host. This argument is required only if the proxy flow step is used." yaml:"proxy_docker_host" envconfig:"proxy_docker_host"`
	ProxyDockerCertPath     string   `long:"proxy-docker-cert-path" description:"Docker certification path for the proxy host." yaml:"proxy_docker_cert_path" envconfig:"proxy_docker_cert_path"`
	ProxyReconfPort         string   `long:"proxy-reconf-port" description:"The port used by the proxy to reconfigure its configuration" yaml:"proxy_reconf_port" envconfig:"proxy_reconf_port"`
	ServicePath             []string `long:"service-path" description:"Path that should be configured in the proxy (e.g. /api/v1/my-service). This argument is required only if the proxy flow step is used." yaml:"service_path"`
	ServiceName             string
	CurrentColor            string
	NextColor               string
	CurrentTarget           string
	NextTarget              string
}

var GetOpts = func() (Opts, error) {
	opts := Opts{
		ComposePath: dockerComposePath,
		Flow:        []string{"deploy"},
	}
	if err := parseYml(&opts); err != nil {
		return opts, err
	}
	if err := parseEnvVars(&opts); err != nil {
		return opts, err
	}
	if err := parseArgs(&opts); err != nil {
		return opts, err
	}
	if err := processOpts(&opts); err != nil {
		return opts, err
	}
	return opts, nil
}

func ParseYml(opts *Opts) error {
	data, err := readFile(dockerFlowPath)
	if err != nil {
		return nil
	}
	if err = yaml.Unmarshal([]byte(data), opts); err != nil {
		return fmt.Errorf("Could not parse the Docker Flow file %s\n%v", dockerFlowPath, err)
	}
	return nil
}

func ParseArgs(opts *Opts) error {
	if _, err := flags.ParseArgs(opts, os.Args[1:]); err != nil {
		return fmt.Errorf("Could not parse command line arguments\n%v", err)
	}
	return nil
}

func ParseEnvVars(opts *Opts) error {
	if err := envconfig.Process("flow", opts); err != nil {
		return fmt.Errorf("Could not retrieve environment variables\n%v", err)
	}
	data := []struct {
		key   string
		value *[]string
	}{
		{"FLOW_SIDE_TARGETS", &opts.SideTargets},
		{"FLOW", &opts.Flow},
		{"FLOW_SERVICE_PATH", &opts.ServicePath},
	}
	for _, d := range data {
		value := strings.Trim(os.Getenv(d.key), " ")
		if len(value) > 0 {
			*d.value = strings.Split(value, ",")
		}
	}
	return nil
}

func ProcessOpts(opts *Opts) (err error) {
	sc := getServiceDiscovery()
	if len(opts.Project) == 0 {
		dir, _ := getWd()
		opts.Project = dir[strings.LastIndex(dir, string(os.PathSeparator))+1:]
	}
	if len(opts.Target) == 0 {
		return fmt.Errorf("target argument is required")
	}
	if len(opts.ServiceDiscoveryAddress) == 0 {
		return fmt.Errorf("consul-address argument is required")
	}
	if len(opts.Scale) != 0 {
		if _, err := strconv.Atoi(opts.Scale); err != nil {
			return fmt.Errorf("scale must be a number or empty")
		}
	}
	if len(opts.Flow) == 0 {
		opts.Flow = []string{"deploy"}
	}
	if len(opts.ServiceName) == 0 {
		opts.ServiceName = fmt.Sprintf("%s-%s", opts.Project, opts.Target)
		logPrintln(fmt.Sprintf("ServiceName (%s)", opts.ServiceName))
	}
	if len(opts.ProxyReconfPort) == 0 {
		opts.ProxyReconfPort = strconv.Itoa(ProxyReconfigureDefaultPort)
	}
	if opts.CurrentColor, err = sc.GetColor(opts.ServiceDiscoveryAddress, opts.ServiceName); err != nil {
		return err
	}
	if len(opts.Host) == 0 {
		opts.Host = os.Getenv("DOCKER_HOST")
	}
	if len(opts.CertPath) == 0 {
		opts.CertPath = os.Getenv("DOCKER_CERT_PATH")
	}
	opts.NextColor = sc.GetNextColor(opts.CurrentColor)
	if opts.BlueGreen {
		opts.NextTarget = fmt.Sprintf("%s-%s", opts.Target, opts.NextColor)
		opts.CurrentTarget = fmt.Sprintf("%s-%s", opts.Target, opts.CurrentColor)
	} else {
		opts.NextTarget = opts.Target
		opts.CurrentTarget = opts.Target
	}
	return nil
}
