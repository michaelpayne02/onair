# On Air

This tiny application acts as an extension to vMix. It repeatedly attempts to connect to vMix's TCP API and subscribes to the Recording, Streaming, External, and MultiCorder activators. It then toggles the On-Air light vi MQTT based on the state of each instance's activators.

