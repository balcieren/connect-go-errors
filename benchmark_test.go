package connectgoerrors

import "testing"

func BenchmarkFormatTemplate(b *testing.B) {
	tpl := "Resource '{{id}}' not found in {{service}} (tenant: {{tenant}})"
	data := M{"id": "user-123", "service": "auth", "tenant": "acme"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatTemplate(tpl, data)
	}
}

func BenchmarkFormatTemplateNoPlaceholders(b *testing.B) {
	tpl := "Internal server error"
	data := M{"id": "123"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatTemplate(tpl, data)
	}
}

func BenchmarkLookup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lookup(NotFound)
	}
}

func BenchmarkLookupParallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Lookup(NotFound)
		}
	})
}

func BenchmarkErr(b *testing.B) {
	data := M{"id": "user-123"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(NotFound, data)
	}
}

func BenchmarkErrParallel(b *testing.B) {
	data := M{"id": "user-123"}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			New(NotFound, data)
		}
	})
}
