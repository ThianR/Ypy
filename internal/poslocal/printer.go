package poslocal

import "errors"

var ErrPrinterUnavailable = errors.New("printer unavailable")

type PrintJob struct {
	ID, OperationID, Status string
	Attempts                int
}
type PrinterQueue struct{ jobs map[string]PrintJob }

func NewPrinterQueue() *PrinterQueue { return &PrinterQueue{jobs: map[string]PrintJob{}} }
func (q *PrinterQueue) Enqueue(operationID string) PrintJob {
	id := "ticket-" + operationID
	job := PrintJob{ID: id, OperationID: operationID, Status: "PENDING"}
	q.jobs[id] = job
	return job
}
func (q *PrinterQueue) Print(id string, available bool) error {
	job, ok := q.jobs[id]
	if !ok {
		return errors.New("print job not found")
	}
	job.Attempts++
	if !available {
		job.Status = "UNKNOWN"
		q.jobs[id] = job
		return ErrPrinterUnavailable
	}
	job.Status = "PRINTED"
	q.jobs[id] = job
	return nil
}
func (q *PrinterQueue) Job(id string) (PrintJob, bool) { job, ok := q.jobs[id]; return job, ok }
