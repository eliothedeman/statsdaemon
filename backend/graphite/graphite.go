package graphite

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func submit(deadline time.Time) error {
	var buffer bytes.Buffer
	var num int64

	now := time.Now().Unix()

	if *graphiteAddress == "-" {
		return nil
	}

	client, err := net.Dial("tcp", *graphiteAddress)
	if err != nil {
		if *debug {
			log.Printf("WARNING: resetting counters when in debug mode")
			processCounters(&buffer, now)
			processGauges(&buffer, now)
			processTimers(&buffer, now, percentThreshold)
			processSets(&buffer, now)
		}
		errmsg := fmt.Sprintf("dialing %s failed - %s", *graphiteAddress, err)
		return errors.New(errmsg)
	}
	defer client.Close()

	err = client.SetDeadline(deadline)
	if err != nil {
		errmsg := fmt.Sprintf("could not set deadline:", err)
		return errors.New(errmsg)
	}

	num += processCounters(&buffer, now)
	num += processGauges(&buffer, now)
	num += processTimers(&buffer, now, percentThreshold)
	num += processSets(&buffer, now)
	if num == 0 {
		return nil
	}

	if *debug {
		for _, line := range bytes.Split(buffer.Bytes(), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			log.Printf("DEBUG: %s", line)
		}
	}

	_, err = client.Write(buffer.Bytes())
	if err != nil {
		errmsg := fmt.Sprintf("failed to write stats - %s", err)
		return errors.New(errmsg)
	}

	log.Printf("sent %d stats to %s", num, *graphiteAddress)

	return nil
}
