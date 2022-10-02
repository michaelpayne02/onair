package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var opts *mqtt.ClientOptions = mqtt.NewClientOptions()
var mqttClient mqtt.Client

var hosts []string = strings.Split(os.Getenv("VMIX_HOSTS"), ",")

type Vmix struct {
	Host                                        string
	Recording, Streaming, External, MultiCorder bool
}

func (v Vmix) isActive() bool {
	return (v.Recording || v.Streaming || v.External || v.MultiCorder)
}

func (v Vmix) update() *Vmix {
	// Create message
	payload := "OFF"
	if v.isActive() {
		payload = "ON"
	}

	// Publish commands depending on active state
	mqttClient.Publish(os.Getenv("MQTT_TOPIC"), 2, false, payload).Wait()
	return &v
}

func main() {
	// Set MQTT options and create the client
	fmt.Println(opts.ClientID)
	mqtt.ERROR = log.New(os.Stdout, "MQTT:", 0)
	opts.AddBroker(os.Getenv("MQTT_HOST"))
	opts.SetUsername(os.Getenv("MQTT_USERNAME")).SetPassword(os.Getenv("MQTT_PASSWORD"))
	mqttClient = mqtt.NewClient(opts)

	// Attempt to connect to MQTT broker
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to", os.Getenv("MQTT_HOST"))

	// Start routines for each host and hang indefinitely
	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)
		go connect(host)
	}
	wg.Wait()
}

// Regular expression to match incoming messages
var r, _ = regexp.Compile(`ACTS OK (Recording|Streaming|External|MultiCorder) (\d)`)

func connect(host string) {
	for {
		// Instantiate vMix instance
		vmix := Vmix{Host: host}

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

		// Event loop
		for scanner.Scan() {

			// Match only messages pertaining to Activators
			args := r.FindStringSubmatch(scanner.Text())
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
		fmt.Println(host, "Disconnected")
	}
}
