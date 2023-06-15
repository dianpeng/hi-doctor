package trigger

// A thing wrapper around our internal trigger system

import (
	"github.com/robfig/cron/v3"
	"sync"
)

type trigger struct {
	cron  *cron.Cron
	nowWg sync.WaitGroup
}

func newTrigger() *trigger {
	return &trigger{
		cron: cron.New(),
	}
}

var theTrigger *trigger

func init() {
	theTrigger = newTrigger()
}

type CronId int

func Cron(expr string, cb func()) (CronId, error) {
	a, b := theTrigger.cron.AddFunc(expr, cb)
	return CronId(a), b
}

func Remove(id CronId) {
	theTrigger.cron.Remove(cron.EntryID(id))
}

func Now(cb func()) error {
	theTrigger.nowWg.Add(1)
	go func() {
		defer func() {
			theTrigger.nowWg.Done()
		}()
		cb()
	}()
	return nil
}

func Start() {
	theTrigger.cron.Start()
}

func Stop() {
	theTrigger.cron.Stop() // just abort the cron
}

func StopSafely() {
	theTrigger.nowWg.Wait() // wait all the immeidate job to be done
	theTrigger.cron.Stop()  // stop the cron
}
