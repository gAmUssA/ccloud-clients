package main

import (
	"ccloud"
	"encoding/binary"
	"fmt"
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v0/kafka"
)

func main() {

	props := make(map[string]string)
	ccloud.LoadProperties(props)

	schemaRegistryClient := ccloud.CreateSchemaRegistryClient(
		props["schema.registry.url"], props["schema.registry.basic.auth.username"],
		props["schema.registry.basic.auth.password"])

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       props["bootstrap.servers"],
		"broker.version.fallback": "0.10.0.0",
		"api.version.fallback.ms": 0,
		"sasl.mechanisms":         props["sasl.mechanism"],
		"security.protocol":       props["security.protocol"],
		"sasl.username":           props["sasl.username"],
		"sasl.password":           props["sasl.password"],
		"session.timeout.ms":      6000,
		"group.id":                "golang-consumer",
		"auto.offset.reset":       "latest"})
	if err != nil {
		panic(fmt.Sprintf("Failed to create consumer: %s", err))
	}

	consumer.SubscribeTopics([]string{"orders"}, nil)

	for {
		msg, err := consumer.ReadMessage(100 * time.Millisecond)
		if err == nil {
			schemaID := binary.BigEndian.Uint32(msg.Value[1:5])
			codec, err := schemaRegistryClient.GetSchema(int(schemaID))
			if err != nil {
				continue
			}
			native, _, _ := codec.NativeFromBinary(msg.Value[5:])
			order, _ := codec.TextualFromNative(nil, native)
			fmt.Println(string(order))
		}

	}

	consumer.Close()

}