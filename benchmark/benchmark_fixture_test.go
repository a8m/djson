package benchmark

import (
	"encoding/json"
	"testing"

	"github.com/a8m/djson"
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
