package iperf

import (
	"errors"
	"fmt"
	// "math"
	// "net"
	"os/exec"
	// "runtime"
	// "sort"
	// "strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HostIperf is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type HostIperf func(binary string, timeout float64, args ...string) (string, error)

// src/iperf3 -p 80 -c 34.93.220.137 -b 5M -t 5 -B 192.168.0.105 --bind-dev enp4s0 --json
type Iperf struct {
	// wg is used to wait for ping with multiple URLs
	wg sync.WaitGroup

	Log telegraf.Logger `toml:"-"`

	// Time Interval to run iperf (iperf -t <INTERVAL>)
	TimeInterval int `toml:"time_interval"`

	// Bytes to transfer (-n, --bytes n[KM] ... number of bytes to transmit)
	BytesToTransfer string `toml:"bytes"`

	// Bandwidth to use
	Bandwidth string `toml:"bandwidth"`

	// Server addresses
	ServerAddrs []string `toml:"server_addrs"`

	// Client Port to use (FUTURE)
	// ClientPort int

	// Parallel client streams to run (FUTURE, need to think of populating output fields in this case)
	// ClientStreams int

	// Reverse mode (server sends, client receives) .. note cwnd and rtt values (FUTURE)
	// will always be zero with this since those are calculated at server end
	// Reverse bool

	// Bind to specific host/interface
	Bind string `toml:"bind"`

	// Bind to device ... this is available only in latest (bleeding edge) ver of iperf3
	BindDev string `toml:"bind-dev"`

	// Protocol to use .. TCP/UDP/SCTP ... (SCTP supported only on FreeBSD and Linux)
	Protocol string `toml:"protocol"`

	// Method defines how to iperf (native or exec) .. currently on exec is supported
	Method string

	// iperf executable binary
	Binary string

	// Arguments for iperf command. When arguments is not empty, system binary will be used and
	// other options (ping_interval, timeout, etc) will be ignored
	Arguments []string

	// host ping function
	iperfHost HostIperf
}

func (*Iperf) Description() string {
	return "Iperf given url(s) and return statistics"
}

const sampleConfig = `
  ## Server addresses to connect (check man iperf3 for detailed help)
  server_addrs = ["34.93.220.137:80"]

  # method = "exec"

  # Time interval for which iperf should run
  time_interval = 10

  # set target bandwidth to n bits/sec (default 1 Mbit/sec for UDP, unlimited for TCP)
  bandwidth = "1M"

  # number of bytes to transmit (instead of time_interval)
  bytes = "1M"

  # Transport protocol to use for transfer (udp/sctp, default tcp)
  # protocol = "udp"

  # number of blocks to transmit (instead of time_interval or bytes)
  # blockcount = "1K"

  ## Specify the ping executable binary.
  # binary = "iperf3"

  ## bind to a specific host
  # bind = "192.168.0.105"

  ## bind to a specific device (supported only in bleeding-edge iperf3, check your iperf3 support before using this)
  # bind-dev = "eth0"

  ## Arguments for iperf command. When arguments is not empty, the command from
  ## the binary option will be used and other options (TimeInterval, timeout,
  ## etc) will be ignored.
  # arguments = ["-c", "3"]

`

func (*Iperf) SampleConfig() string {
	return sampleConfig
}

func (p *Iperf) Gather(acc telegraf.Accumulator) error {
	for _, address := range p.ServerAddrs {
		p.wg.Add(1)
		go func(address string) {
			defer p.wg.Done()

			switch p.Method {
			case "native":
				errors.New("native method not implemented for iperf")
			default:
				p.iperfToURL(address, acc)
			}
		}(address)
	}

	p.wg.Wait()

	return nil
}

// Init ensures the plugin is configured correctly.
func (p *Iperf) Init() error {
	return nil
}

func hostIperfer(binary string, timeout float64, args ...string) (string, error) {
	bin, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	fmt.Println("COMMAND: ", bin, args)
	c := exec.Command(bin, args...)
	out, err := internal.CombinedOutputTimeout(c,
		time.Second*time.Duration(timeout+5))
	return string(out), err
}

func init() {
	inputs.Add("iperf", func() telegraf.Input {
		p := &Iperf{
			iperfHost:    hostIperfer,
			TimeInterval: 10.0,
			Method:       "exec",
			Binary:       "iperf3",
			Arguments:    []string{},
		}
		return p
	})
}
