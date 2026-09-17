package poslocal

import (
	"errors"
	"gorm.io/gorm"
)

type PersistentPrinter struct{ DB *gorm.DB }

var ErrInvalidPrinterStore = errors.New("invalid printer store")

func (p PersistentPrinter) Pending() ([]PrintJob, error) {
	if p.DB == nil {
		return nil, ErrInvalidPrinterStore
	}
	var jobs []PrintJob
	result := p.DB.Raw(`SELECT job_id AS id,operation_id,status,attempts FROM pos_print_jobs WHERE status IN ('PENDING','UNKNOWN') ORDER BY rowid`).Scan(&jobs)
	return jobs, result.Error
}

func (p PersistentPrinter) Enqueue(operationID string) (PrintJob, error) {
	if p.DB == nil || operationID == "" {
		return PrintJob{}, errors.New("invalid print job")
	}
	job := PrintJob{ID: "ticket-" + operationID, OperationID: operationID, Status: "PENDING"}
	if result := p.DB.Exec(`INSERT INTO pos_print_jobs(job_id,operation_id,status) VALUES(?,?,?) ON CONFLICT(operation_id) DO NOTHING`, job.ID, job.OperationID, job.Status); result.Error != nil {
		return PrintJob{}, result.Error
	}
	return p.Job(job.ID)
}

func (p PersistentPrinter) Print(id string, available bool) error {
	if p.DB == nil {
		return ErrInvalidPrinterStore
	}
	status := "PRINTED"
	if !available {
		status = "UNKNOWN"
	}
	result := p.DB.Exec(`UPDATE pos_print_jobs SET status=?, attempts=attempts+1 WHERE job_id=?`, status, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("print job not found")
	}
	if !available {
		return ErrPrinterUnavailable
	}
	return nil
}

func (p PersistentPrinter) Job(id string) (PrintJob, error) {
	if p.DB == nil {
		return PrintJob{}, ErrInvalidPrinterStore
	}
	var job PrintJob
	result := p.DB.Raw(`SELECT job_id AS id,operation_id,status,attempts FROM pos_print_jobs WHERE job_id=?`, id).Scan(&job)
	return job, result.Error
}
