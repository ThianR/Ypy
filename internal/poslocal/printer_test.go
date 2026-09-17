package poslocal

import "testing"

func TestPrinterFailureKeepsJobForRetry(t *testing.T) {
	q := NewPrinterQueue()
	job := q.Enqueue("op-1")
	if err := q.Print(job.ID, false); err != ErrPrinterUnavailable {
		t.Fatal("printer failure not reported")
	}
	saved, _ := q.Job(job.ID)
	if saved.Status != "UNKNOWN" {
		t.Fatal("uncertain print was lost")
	}
	if err := q.Print(job.ID, true); err != nil {
		t.Fatal(err)
	}
	saved, _ = q.Job(job.ID)
	if saved.Status != "PRINTED" || saved.Attempts != 2 {
		t.Fatal("retry not recorded")
	}
}
