package sync

import (
	//"path"
	"sync"

	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
)

type Job struct {
	SourcePath string
	DestPath   string
	Type       ftp.EntryType
}

type Scheduler struct {
	toDo     chan Job
	fetchers []*fetcher
	wg       sync.WaitGroup
}

type fetcher struct {
	id     int
	source *ftp.ServerConn
	dest   *ftp.ServerConn
}

func (f *fetcher) run(s *Scheduler) {
	s.wg.Add(1)
	defer s.wg.Done()
	for job := range s.toDo {
		log.WithFields(log.Fields{
			"thread-id": f.id,
			"from":      job.SourcePath,
			"to":        job.DestPath,
		}).Debug("Syncing new file")

		err := f.fetchOne(s, job)
		if err != nil {
			log.WithFields(log.Fields{
				"thread-id": f.id,
				"job":       job,
				"error":     err,
			}).Error("Error downloading file")
		}
	}
}

func (f *fetcher) fetchOne(s *Scheduler, job Job) error {

	switch job.Type {
	case ftp.EntryTypeFile:
		r, err := f.source.Retr(job.SourcePath)
		if err != nil {
			return err
		}
		defer r.Close()

		err = f.dest.Stor(job.DestPath, r)
		if err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}
func NewScheduler(toDo chan Job) *Scheduler {
	return &Scheduler{
		toDo: toDo,
		wg:   sync.WaitGroup{},
	}
}

func (s *Scheduler) AddFetcher(source, dest *ftp.ServerConn) {
	fetcher := &fetcher{
		id:     len(s.fetchers),
		source: source,
		dest:   dest,
	}

	s.fetchers = append(s.fetchers, fetcher)
	go fetcher.run(s)
}

func (s *Scheduler) Wait() {
	s.wg.Wait()
}
