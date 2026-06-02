// go:build goexperiment.jsonv2

package jsonbench

import (
	jsonv2 "encoding/json/v2"
	"testing"
)

func BenchmarkMarshal_StdV2(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := jsonv2.Marshal(&testOrder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_StdV2(b *testing.B) {
	data, _ := jsonv2.Marshal(&testOrder)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var o Order
		if err := jsonv2.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}
