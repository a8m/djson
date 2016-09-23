package benchmark

import (
	"encoding/json"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/a8m/djson"
	"github.com/antonholmquist/jason"
	"github.com/bitly/go-simplejson"
	"github.com/mreiferson/go-ujson"
	"github.com/ugorji/go/codec"
)

func BenchmarkEncodingJsonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make(map[string]interface{})
			json.Unmarshal(smallFixture, &data)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make(map[string]interface{})
			json.Unmarshal(mediumFixture, &data)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make(map[string]interface{})
			json.Unmarshal(largeFixture, &data)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := make(map[string]interface{})
			json.Unmarshal(largeArrayFixture, &data)
		}
	})
}

func BenchmarkUgorjiParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decoder := codec.NewDecoderBytes(smallFixture, new(codec.JsonHandle))
			var v interface{}
			decoder.Decode(&v)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decoder := codec.NewDecoderBytes(mediumFixture, new(codec.JsonHandle))
			var v interface{}
			decoder.Decode(&v)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decoder := codec.NewDecoderBytes(largeFixture, new(codec.JsonHandle))
			var v interface{}
			decoder.Decode(&v)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decoder := codec.NewDecoderBytes(largeArrayFixture, new(codec.JsonHandle))
			var v interface{}
			decoder.Decode(&v)
		}
	})
}

func BenchmarkJasonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jason.NewObjectFromBytes(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jason.NewObjectFromBytes(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jason.NewObjectFromBytes(largeFixture)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jason.NewObjectFromBytes(largeArrayFixture)
		}
	})
}

func BenchmarkSimpleJsonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simplejson.NewJson(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simplejson.NewJson(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simplejson.NewJson(largeFixture)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			simplejson.NewJson(largeArrayFixture)
		}
	})
}

func BenchmarkGabsParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gabs.ParseJSON(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gabs.ParseJSON(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gabs.ParseJSON(largeFixture)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			gabs.ParseJSON(largeArrayFixture)
		}
	})
}

func BenchmarkUJsonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ujson.NewFromBytes(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ujson.NewFromBytes(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ujson.NewFromBytes(largeFixture)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ujson.NewFromBytes(largeArrayFixture)
		}
	})
}

func BenchmarkDJsonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.DecodeObject(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.DecodeObject(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.DecodeObject(largeFixture)
		}
	})

	b.Run("large_array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.DecodeArray(largeArrayFixture)
		}
	})
}

/*
// This is not part of the benchmark test cases;
// Trying to show the preformence while translate the jsonparser's
// result into map[string]interface{}
// import "github.com/buger/jsonparser"
func BenchmarkJsonparserParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]interface{})
			jsonparser.ObjectEach(smallFixture, func(k, v []byte, vt jsonparser.ValueType, o int) error {
				if vt == jsonparser.Number {
					m[string(k)], _ = strconv.ParseFloat(string(v), 64)
				} else {
					m[string(k)] = string(v)
				}
				return nil
			})
		}
	})
}
*/
