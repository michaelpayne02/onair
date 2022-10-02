package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	dotenv "github.com/joho/godotenv"
)

// Create options and client
var opts *mqtt.ClientOptions = mqtt.NewClientOptions()
var mqttClient mqtt.Client = mqtt.NewClient(opts)

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

	// Set options
	opts.AddBroker(getEnv("MQTT_HOST"))
	opts.SetUsername(getEnv("MQTT_USERNAME")).SetPassword(getEnv("MQTT_HOST"))

	// Attempt to connect to MQTT broker
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to", getEnv("MQTT_HOST"))

	// Start routines for each host and hang indefinitely
	hosts := getEnv("VMIX_HOSTS")
	topic := getEnv("MQTT_TOPIC")
	var wg sync.WaitGroup
	for _, host := range strings.Split(hosts, ",") {
		wg.Add(1)
		// Create & connect vMix instance
		vmix := Vmix{Host: host}
		go vmix.connect(topic)
	}
	wg.Wait()
}
