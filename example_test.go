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
	fmt.Printf("%#v", val)
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

func ExampleDecodeObject() {
	var data = []byte(`{
		"id": 76523,
		"name": "Ariel",
		"username": "a8m",
		"score": 99,
		"image": {
			"src": "images/67.png",
			"height": 450,
			"width":  370,
			"alignment": "center"
		}
	}`)
	user, err := djson.DecodeObject(data)
	if err != nil {
		log.Fatal("error:", err)
	}
	fmt.Printf("User: %v", user)
}
