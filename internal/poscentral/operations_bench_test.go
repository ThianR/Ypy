package poscentral

import "testing"

func BenchmarkMemoryStoreApplyIdempotent(b *testing.B) {
	s := NewStore()
	op := Operation{ID: "00000000-0000-0000-0000-000000000001", EmpresaID: "4", TerminalID: "2", Payload: `{"sequence":1,"total":"10","currency":"PYG"}`}
	if _, err := s.Apply(op); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Apply(op); err != nil {
			b.Fatal(err)
		}
	}
}
