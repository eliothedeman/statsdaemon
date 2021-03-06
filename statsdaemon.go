package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eliothedeman/statsdaemon/backend"
	_ "github.com/eliothedeman/statsdaemon/backend/console"
	_ "github.com/eliothedeman/statsdaemon/backend/sqs"
	"github.com/eliothedeman/statsdaemon/config"
	"github.com/eliothedeman/statsdaemon/metric"
)

const (
	MAX_UNPROCESSED_PACKETS = 1000
	MAX_UDP_PACKET_SIZE     = 512
	RECIEVED_PACKET_BUCKET  = "packets_recieved"
)

var signalchan chan os.Signal

type Packet struct {
	Bucket   string
	Value    interface{}
	Modifier string
	Sampling float32
}

type GaugeData struct {
	Relative bool
	Negative bool
	Value    uint64
}

type Percentiles []*Percentile
type Percentile struct {
	float float64
	str   string
}

func (a *Percentiles) Set(s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*a = append(*a, &Percentile{f, strings.Replace(s, ".", "_", -1)})
	return nil
}
func (p *Percentile) String() string {
	return p.str
}
func (a *Percentiles) String() string {
	return fmt.Sprintf("%v", *a)
}

var (
	percentThreshold = Percentiles{}
	configFileName   = flag.String("conf", "conf.json", "path to config file")
)

func init() {
	flag.Var(&percentThreshold, "percent-threshold",
		"percentile calculation for timers (0-100, may be given multiple times)")
	metrics.Add(RECIEVED_PACKET_BUCKET, metric.NewCounter(RECIEVED_PACKET_BUCKET))
}

var (
	In       = make(chan *Packet, MAX_UNPROCESSED_PACKETS)
	metrics  = metric.NewContainer(metric.ExpireTime(time.Minute * 30))
	backends []backend.Backend
)

func monitor(conf *config.Config) {
	ticker := time.NewTicker(conf.FlushInterval)
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			if err := submit(time.Now().Add(conf.FlushInterval)); err != nil {
				log.Printf("ERROR: %s", err)
			}
			return
		case <-ticker.C:
			go func() {
				start := time.Now()
				if err := submit(time.Now().Add(conf.FlushInterval)); err != nil {
					log.Printf("ERROR: %s", err)
				}
				log.Printf("submitting took %s", time.Now().Sub(start))
			}()
		case s := <-In:
			packetHandler(s)
		}
	}
}

func packetHandler(s *Packet) {
	metrics.Update(RECIEVED_PACKET_BUCKET, int64(1), 1)

	metricToUpdate := metrics.Get(s.Bucket)
	if metricToUpdate == nil {
		switch s.Modifier {
		case "ms":
			metricToUpdate = metric.NewTimer(s.Bucket)
		case "g":
			metricToUpdate = metric.NewGauge(s.Bucket)
		case "c":
			metricToUpdate = metric.NewCounter(s.Bucket)
		case "s":
			metricToUpdate = metric.NewSet(s.Bucket)
		}
		metrics.Add(s.Bucket, metricToUpdate)
	}
	metricToUpdate.Update(s.Value, s.Sampling)
}

func submit(deadline time.Time) error {
	all := processMetrics()
	for i := range backends {
		err := backends[i].Submit(all, deadline)
		if err != nil {
			return err
		}
		log.Printf("Submitted %d metrics to %s", len(all), reflect.TypeOf(backends[i]).Elem().Name())
	}
	return nil
}

func processMetrics() []metric.Metric {
	metrics.Reap()
	return metrics.List()
}

func parseMessage(data []byte, conf *config.Config) []*Packet {
	var (
		output []*Packet
		input  []byte
	)

	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		input = line

		index := bytes.IndexByte(input, ':')
		if index < 0 || index == len(input)-1 {
			if conf.Debug {
				log.Printf("ERROR: failed to parse line: %s\n", string(line))
			}
			continue
		}

		name := input[:index]

		index++
		input = input[index:]

		index = bytes.IndexByte(input, '|')
		if index < 0 || index == len(input)-1 {
			if conf.Debug {
				log.Printf("ERROR: failed to parse line: %s\n", string(line))
			}
			continue
		}

		val := input[:index]
		index++

		var mtypeStr string

		if input[index] == 'm' {
			index++
			if index >= len(input) || input[index] != 's' {
				if conf.Debug {
					log.Printf("ERROR: failed to parse line: %s\n", string(line))
				}
				continue
			}
			mtypeStr = "ms"
		} else {
			mtypeStr = string(input[index])
		}

		index++
		input = input[index:]

		var (
			value interface{}
			err   error
		)

		if mtypeStr[0] == 'c' {
			value, err = strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				log.Printf("ERROR: failed to ParseInt %s - %s", string(val), err)
				continue
			}
		} else if mtypeStr[0] == 'g' {
			var relative, negative bool
			var stringToParse string

			switch val[0] {
			case '+', '-':
				relative = true
				negative = val[0] == '-'
				stringToParse = string(val[1:])
			default:
				relative = false
				negative = false
				stringToParse = string(val)
			}

			gaugeValue, err := strconv.ParseFloat(stringToParse, 64)
			if err != nil {
				log.Printf("ERROR: failed to ParseFloat %s - %s", string(val), err)
				continue
			}

			value = metric.GaugeData{relative, negative, gaugeValue}
		} else if mtypeStr[0] == 's' {
			value = string(val)
		} else {
			value, err = strconv.ParseFloat(string(val), 64)
			if err != nil {
				log.Printf("ERROR: failed to ParseUint %s - %s", string(val), err)
				continue
			}
		}

		var sampleRate float32 = 1

		if len(input) > 0 && bytes.HasPrefix(input, []byte("|@")) {
			input = input[2:]
			rate, err := strconv.ParseFloat(string(input), 32)
			if err == nil {
				sampleRate = float32(rate)
			}
		}

		packet := &Packet{
			Bucket:   conf.Prefix + string(name),
			Value:    value,
			Modifier: mtypeStr,
			Sampling: sampleRate,
		}
		output = append(output, packet)
	}
	return output
}

func udpListener(conf *config.Config) {
	address, _ := net.ResolveUDPAddr("udp", conf.ServiceAddress)
	log.Printf("listening on %s", address)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	defer listener.Close()

	message := make([]byte, MAX_UDP_PACKET_SIZE)
	for {
		n, remaddr, err := listener.ReadFromUDP(message)
		if err != nil {
			log.Printf("ERROR: reading UDP packet from %+v - %s", remaddr, err)
			continue
		}

		for _, p := range parseMessage(message[:n], conf) {
			In <- p
		}
	}
}

func main() {
	flag.Parse()

	signalchan = make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGTERM)
	conf, err := config.LoadConfigFromFile(*configFileName)
	if err != nil {
		log.Fatal(err)
	}
	backends, err = conf.InitBackends()
	if err != nil {
		log.Fatal(err)
	}

	go udpListener(conf)
	monitor(conf)
}
