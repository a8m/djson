package djson_test

import (
	"fmt"
	"log"

	"github.com/a8m/djson"
)

func ExampleDecode() {
	var data = []byte(`[
		{"Name": "Platypus", "Order": "Monotremata"},
		{"Name": "Quoll",    "Order": "Dasyuromorphia"}
	]`)

	val, err := djson.Decode(data)
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Printf("%+v", val)

	// Output:
	// [map[Name:Platypus Order:Monotremata] map[Name:Quoll Order:Dasyuromorphia]]
}

func ExampleDecodeArray() {
	var data = []byte(`[
		"John",
		"Dan",
		"Kory",
		"Ariel"
	]`)

	users, err := djson.DecodeArray(data)
	if err != nil {
		log.Fatal("error:", err)
	}
	for i, user := range users {
		fmt.Printf("[%d]: %v\n", i, user)
	}
}

// Example that demonstrate the basic transformation I do on each incoming
// event.
// `lowerKeys` and `§fixEncoding` are two generic methods, and they don't care
// about the schema.
// The three others(`maxMindGeo`, `dateFormat`and `refererURL`) process and
// extend the events dynamically based on the "APP_ID" field.
func ExampleDecodeObject() {
	var data = []byte(`{
		"ID": 76523,
		"IP": "69.89.31.226"
		"APP_ID": "BD311",
		"Name": "Ariel",
		"Username": "a8m",
		"Score": 99,
		"Date": 1475332371532,
		"Image": {
			"Src": "images/67.png",
			"Height": 450,
			"Width":  370,
			"Alignment": "center"
		},
		"RefererURL": "http://..."
	}`)
	event, err := djson.DecodeObject(data)
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Printf("Value: %v", event)

	// Process the event
	//
	// lowerKeys(event)
	// fixEncoding(event)
	// dateFormat(event)
	// maxMindGeo(event)
	// refererURL(event)
	//
	// pipeline.Pipe(event)
}

func ExampleDecoder_AllocString() {
	var data = []byte(`{"event_type":"click","count":"93","userid":"4234A"}`)
	dec := djson.NewDecoder(data)
	dec.AllocString()

	val, err := dec.DecodeObject()
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Printf("Value: %+v", val)

	// Output:
	// map[count:93 userid:4234A event_type:click]
}
