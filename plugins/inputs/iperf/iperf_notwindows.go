// +build !windows

package iperf

import (
	"encoding/json"
	"errors"
	"fmt"
	// "os/exec"
	// "regexp"
	"runtime"
	"strconv"
	"strings"
	// "syscall"

	"github.com/influxdata/telegraf"
)

func (p *Iperf) iperfToURL(address string, acc telegraf.Accumulator) {
	tags := map[string]string{"address": address}
	fields := map[string]interface{}{"result_code": 0}

	out, err := p.iperfHost(p.Binary, 60.0, p.args(address, runtime.GOOS)...)
	if err != nil {
		acc.AddError(fmt.Errorf("host %s: %s", address, err))
		fields["result_code"] = 2
		acc.AddFields("iperf", fields, tags)
		return
	}
	// fmt.Println(out)
	err = processIperfOutput(out, fields)
	if err != nil {
		// fatal error
		acc.AddError(fmt.Errorf("%s: %s", err, address))
		fields["result_code"] = 2
		acc.AddFields("iperf", fields, tags)
		return
	}

	acc.AddFields("iperf", fields, tags)
}

func (p *Iperf) args(address string, system string) []string {
	srvPort := "5201"
	str := strings.Split(address, ":")
	srvAddr := str[0]
	if len(str) > 1 {
		srvPort = str[1]
	}

	// build the iperf command args based on toml config
	args := []string{"-p", srvPort, "-c", srvAddr, "--json"}
	if len(p.Arguments) > 0 {
		return append(p.Arguments, address)
	}
	if p.Bandwidth != "" {
		args = append(args, "--bandwidth", p.Bandwidth)
	}
	if p.Bind != "" {
		args = append(args, "--bind", p.Bind)
	}
	if p.BindDev != "" {
		args = append(args, "--bind-dev", p.BindDev)
	}

	// Either number of bytes or the time-interval is used
	if p.BytesToTransfer != "" {
		args = append(args, "--bytes", p.BytesToTransfer)
	} else {
		args = append(args, "--time", strconv.Itoa(p.TimeInterval))
	}

	if p.Protocol != "" {
		switch p.Protocol {
		case "udp", "UDP":
			args = append(args, "--udp")
		case "sctp", "SCTP":
			args = append(args, "--sctp")
		}
	}
	return args
}

// processIperfOutput takes in a json string from the iperf command
func processIperfOutput(jsonStr string, fields map[string]interface{}) error {
	var jd map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &jd)
	if err != nil {
		return err
	}
	endData := jd["end"].(map[string]interface{})
	if endData == nil {
		return errors.New("could not find 'end' in the json string")
	}
	streams := endData["streams"].([]interface{})
	streamsData := streams[0].(map[string]interface{})
	sender := streamsData["sender"].(map[string]interface{})

	fields["bytes_sent"] = sender["bytes"].(float64)
	fields["retransmits"] = sender["retransmits"].(float64)
	fields["bits_per_second"] = sender["bits_per_second"].(float64)
	fields["max_snd_cwnd"] = sender["max_snd_cwnd"].(float64)
	fields["max_rtt"] = sender["max_rtt"].(float64) / 1000
	fields["min_rtt"] = sender["min_rtt"].(float64) / 1000
	fields["mean_rtt"] = sender["mean_rtt"].(float64) / 1000

	return nil
}
