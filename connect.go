package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"time"
)

type Vmix struct {
	Host                                        string
	Recording, Streaming, External, MultiCorder bool
}

func (vmix Vmix) isActive() bool {
	return (vmix.Recording || vmix.Streaming || vmix.External || vmix.MultiCorder)
}

func (vmix Vmix) update(topic string) *Vmix {
	// Create message
	payload := "OFF"
	if vmix.isActive() {
		payload = "ON"
	}

	// Publish commands depending on active state
	mqttClient.Publish(topic, 2, false, payload).Wait()
	return &vmix
}

// Regular expression to match incoming messages
var activators, _ = regexp.Compile(`ACTS OK (Recording|Streaming|External|MultiCorder) (\d)`)

func (vmix Vmix) connect(topic string) *Vmix {
	for {
		// Attempt to connect to the tcp socket every 15 seconds
		conn, err := net.DialTimeout("tcp", vmix.Host, 5*time.Second)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		fmt.Println(vmix.Host, "Connected")

		// Init: subscribe to activators and query the state of them
		vmix.update()
		fmt.Fprint(conn, "SUBSCRIBE ACTS\r\nACTS Recording\r\nACTS Streaming\r\n"+
			"ACTS External\r\nACTS MultiCorder\r\n")

		// Create a scanner that iterates over lines
		scanner := bufio.NewScanner(conn)

		// Loop through all lines
		for scanner.Scan() {

			// Match only messages pertaining to Activators
			args := activators.FindStringSubmatch(scanner.Text())
			if len(args) == 0 {
				continue
			}

			var act, state = args[1], args[2] == "1"
			prevState := vmix.isActive()

			// Update the state of vMix
			switch act {
			case "Recording":
				vmix.Recording = state
			case "Streaming":
				vmix.Streaming = state
			case "External":
				vmix.External = state
			case "MultiCorder":
				vmix.MultiCorder = state
			default:
			}

			// Only publish messages if active state has changed
			if vmix.isActive() != prevState {
				fmt.Println(vmix.Host, act, state)
				vmix.update()
			}
		}
		Vmix{}.update(topic)
		log.Println(vmix.Host, "Disconnected")
	}
}
