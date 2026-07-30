package ingest

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScanJob is an async agent run (single source or bulk).
type ScanJob struct {
	ID        string             `json:"id"`
	Kind      string             `json:"kind"` // "single" | "bulk"
	Handle    string             `json:"handle,omitempty"`
	Status    string             `json:"status"` // queued | running | done | failed
	Step      string             `json:"step"`
	Progress  int                `json:"progress"` // 0-100
	Message   string             `json:"message,omitempty"`
	Result    *SourceScanResult  `json:"result,omitempty"`
	Results   []SourceScanResult `json:"results,omitempty"` // bulk
	Error     string             `json:"error,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type jobStore struct {
	mu   sync.RWMutex
	jobs map[string]*ScanJob
}

func newJobStore() *jobStore {
	return &jobStore{jobs: map[string]*ScanJob{}}
}

func (j *jobStore) put(job *ScanJob) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.jobs[job.ID] = job
}

func (j *jobStore) get(id string) (*ScanJob, bool) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	job, ok := j.jobs[id]
	if !ok {
		return nil, false
	}
	// copy
	cp := *job
	return &cp, true
}

func (j *jobStore) update(id string, fn func(*ScanJob)) {
	j.mu.Lock()
	defer j.mu.Unlock()
	job, ok := j.jobs[id]
	if !ok {
		return
	}
	fn(job)
	job.UpdatedAt = time.Now().UTC()
}

// StartScanJob runs ScanSource in the background and returns a job id immediately.
func (w *Worker) StartScanJob(handle string, limit int) string {
	if w.jobs == nil {
		w.jobs = newJobStore()
	}
	id := "job_" + uuid.NewString()
	job := &ScanJob{
		ID: id, Kind: "single", Handle: handle, Status: "queued",
		Step: "queued", Progress: 0, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	w.jobs.put(job)
	go func() {
		w.jobs.update(id, func(j *ScanJob) {
			j.Status = "running"
			j.Step = "profile"
			j.Progress = 5
			j.Message = "Refreshing profile & location…"
		})
		// Use background context so client disconnect doesn't cancel the job.
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
		defer cancel()
		// Progress hooks via wrapping — ScanSource updates step at key points if we pass a callback.
		// For simplicity we stage around ScanSource.
		w.jobs.update(id, func(j *ScanJob) {
			j.Step = "posts"
			j.Progress = 25
			j.Message = "Fetching posts via Bright Data…"
		})
		res, err := w.ScanSource(ctx, handle, limit)
		if err != nil {
			w.jobs.update(id, func(j *ScanJob) {
				j.Status = "failed"
				j.Step = "failed"
				j.Progress = 100
				j.Error = err.Error()
				j.Message = err.Error()
			})
			return
		}
		_ = w.store.TouchSourceScan(handle, len(res.Couples), res.ActionsCreated)
		w.jobs.update(id, func(j *ScanJob) {
			j.Status = "done"
			j.Step = "done"
			j.Progress = 100
			j.Message = "Scan complete"
			cp := res
			j.Result = &cp
		})
	}()
	return id
}

// StartBulkScanJob scans many sources sequentially (stale photographers by default).
func (w *Worker) StartBulkScanJob(handles []string, limit int) string {
	if w.jobs == nil {
		w.jobs = newJobStore()
	}
	if limit <= 0 {
		limit = 12
	}
	id := "job_" + uuid.NewString()
	job := &ScanJob{
		ID: id, Kind: "bulk", Status: "queued", Step: "queued", Progress: 0,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		Results: []SourceScanResult{},
	}
	w.jobs.put(job)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
		defer cancel()
		w.jobs.update(id, func(j *ScanJob) {
			j.Status = "running"
			j.Step = "bulk"
			j.Message = "Starting bulk scan…"
		})
		n := len(handles)
		if n == 0 {
			w.jobs.update(id, func(j *ScanJob) {
				j.Status = "done"
				j.Progress = 100
				j.Step = "done"
				j.Message = "No sources to scan"
			})
			return
		}
		var results []SourceScanResult
		for i, h := range handles {
			if ctx.Err() != nil {
				break
			}
			pct := int(float64(i) / float64(n) * 100)
			w.jobs.update(id, func(j *ScanJob) {
				j.Handle = h
				j.Progress = pct
				j.Step = "scanning"
				j.Message = "Scanning @" + h + "…"
			})
			res, err := w.ScanSource(ctx, h, limit)
			if err != nil {
				res = SourceScanResult{Handle: h, Errors: []string{err.Error()}, Couples: []ScannedCouple{}}
			}
			_ = w.store.TouchSourceScan(h, len(res.Couples), res.ActionsCreated)
			results = append(results, res)
			w.jobs.update(id, func(j *ScanJob) {
				j.Results = append([]SourceScanResult{}, results...)
			})
		}
		w.jobs.update(id, func(j *ScanJob) {
			j.Status = "done"
			j.Step = "done"
			j.Progress = 100
			j.Handle = ""
			j.Message = "Bulk scan complete"
			j.Results = results
		})
	}()
	return id
}

// GetScanJob returns a copy of the job state for polling.
func (w *Worker) GetScanJob(id string) (*ScanJob, bool) {
	if w.jobs == nil {
		return nil, false
	}
	return w.jobs.get(id)
}
