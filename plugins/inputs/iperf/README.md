# Iperf Input Plugin

Uses iperf3 to initiate transfer and check the throughput rate, latency
parameters of the connection.

The plugin uses `method = "exec"` to execute the system's iperf3 utility.

Note, that most of the plugin options are supported in the standard iperf3
utility installed using apt-get. However, there are certain options such as
`--bind-dev` which are supported only in the bleeding edge version of iperf3.
So please check your iperf3 options and if they are supported before using it.

### Configuration:

```toml
[[inputs.ping]]
  ## Hosts to send ping packets to.
  urls = ["example.org"]

  ## Method used for sending pings, can be either "exec" or "native".  When set
  ## to "exec" the systems ping command will be executed.  When set to "native"
  ## the plugin will send pings directly.
  ##
  ## While the default is "exec" for backwards compatibility, new deployments
  ## are encouraged to use the "native" method for improved compatibility and
  ## performance.
  # method = "exec"

  ## Number of ping packets to send per interval.  Corresponds to the "-c"
  ## option of the ping command.
  # count = 1

  ## Time to wait between sending ping packets in seconds.  Operates like the
  ## "-i" option of the ping command.
  # ping_interval = 1.0

  ## If set, the time to wait for a ping response in seconds.  Operates like
  ## the "-W" option of the ping command.
  # timeout = 1.0

  ## If set, the total ping deadline, in seconds.  Operates like the -w option
  ## of the ping command.
  # deadline = 10

  ## Interface or source address to send ping from.  Operates like the -I or -S
  ## option of the ping command.
  # interface = ""

  ## Percentiles to calculate. This only works with the native method.
  # percentiles = [50, 95, 99]

  ## Specify the ping executable binary.
  # binary = "ping"

  ## Arguments for ping command. When arguments is not empty, the command from
  ## the binary option will be used and other options (ping_interval, timeout,
  ## etc) will be ignored.
  # arguments = ["-c", "3"]

  ## Use only IPv6 addresses when resolving a hostname.
  # ipv6 = false
```

#### File Limit

Since this plugin runs the ping command, it may need to open multiple files per
host.  The number of files used is lessened with the `native` option but still
many files are used.  With a large host list you may receive a `too many open
files` error.

To increase this limit on platforms using systemd the recommended method is to
use the "drop-in directory", usually located at
`/etc/systemd/system/telegraf.service.d`.

You can create or edit a drop-in file in the correct location using:
```sh
$ systemctl edit telegraf
```

Increase the number of open files:
```ini
[Service]
LimitNOFILE=8192
```

Restart Telegraf:
```sh
$ systemctl edit telegraf
```

#### Linux Permissions

When using `method = "native"`, Telegraf will attempt to use privileged raw
ICMP sockets.  On most systems, doing so requires `CAP_NET_RAW` capabilities or for Telegraf to be run as root.

With systemd:
```sh
$ systemctl edit telegraf
```
```ini
[Service]
CapabilityBoundingSet=CAP_NET_RAW
AmbientCapabilities=CAP_NET_RAW
```
```sh
$ systemctl restart telegraf
```

Without systemd:
```sh
$ setcap cap_net_raw=eip /usr/bin/telegraf
```

Reference [`man 7 capabilities`][man 7 capabilities] for more information about
setting capabilities.

[man 7 capabilities]: http://man7.org/linux/man-pages/man7/capabilities.7.html

#### Other OS Permissions

When using `method = "native"`, you will need permissions similar to the executable ping program for your OS. 

### Metrics

- ping
  - tags:
    - url
  - fields:
    - packets_transmitted (integer)
    - packets_received (integer)
    - percent_packet_loss (float)
    - ttl (integer, Not available on Windows)
    - average_response_ms (float)
    - minimum_response_ms (float)
    - maximum_response_ms (float)
    - standard_deviation_ms (float, Available on Windows only with method = "native")
    - percentile\<N\>_ms (float, Where `<N>` is the percentile specified in `percentiles`. Available with method = "native" only)
    - errors (float, Windows only)
    - reply_received (integer, Windows with method = "exec" only)
    - percent_reply_loss (float, Windows with method = "exec" only)
    - result_code (int, success = 0, no such host = 1, ping error = 2)

##### reply_received vs packets_received

On Windows systems with `method = "exec"`, the "Destination net unreachable" reply will increment `packets_received` but not `reply_received`*.

##### ttl

There is currently no support for TTL on windows with `"native"`; track
progress at https://github.com/golang/go/issues/7175 and
https://github.com/golang/go/issues/7174


### Example Output

```
ping,url=example.org average_response_ms=23.066,ttl=63,maximum_response_ms=24.64,minimum_response_ms=22.451,packets_received=5i,packets_transmitted=5i,percent_packet_loss=0,result_code=0i,standard_deviation_ms=0.809 1535747258000000000
```
