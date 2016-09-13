package benchmark

import (
	"encoding/json"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/a8m/djson"
	"github.com/antonholmquist/jason"
	"github.com/bitly/go-simplejson"
	"github.com/mreiferson/go-ujson"
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
}

func BenchmarkDJsonParser(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.Decode(smallFixture)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.Decode(mediumFixture)
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			djson.Decode(largeFixture)
		}
	})
}
