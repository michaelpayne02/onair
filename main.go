package main

import (
	"log"
	"os"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	dotenv "github.com/joho/godotenv"
)

var mqttClient mqtt.Client

// Create mqtt client and state
var instances []*Vmix

// Helper function that loads environment variables.
func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	log.Fatalf("The environment variable %s must be defined. Please set it or update the .env file.", key)
	return ""
}

func main() {
	// Load environment variables
	if err := dotenv.Load(); err != nil {
		log.Print("Error loading .env file. System environment variables will still be used.")
	}

	// Set options and create client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(getEnv("MQTT_HOST"))
	opts.SetUsername(getEnv("MQTT_USERNAME")).SetPassword(getEnv("MQTT_PASSWORD"))
	mqttClient = mqtt.NewClient(opts)
	topic := getEnv("MQTT_TOPIC")

	// Attempt to connect to MQTT broker
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Connected to", getEnv("MQTT_HOST"))

	// Start routines for each host and hang indefinitely
	hosts, exists := os.LookupEnv("VMIX_HOSTS")
	if !exists {
		hosts = "127.0.0.1:8099"
	}

	var wg sync.WaitGroup
	wg.Add(1)
	// Create & connect vMix instances using goroutines
	for _, host := range strings.Split(hosts, ",") {
		vmix := Vmix{Host: host}

		instances = append(instances, &vmix)
		go vmix.connect(topic, instances)
	}
	wg.Wait()
}
