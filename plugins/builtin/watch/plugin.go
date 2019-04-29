package watch

import (
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"sync"
	"time"

	"github.com/go-gilbert/gilbert/runner/job"
	"github.com/rjeczalik/notify"
)

// Plugin implements plugins.Plugin interface
type Plugin struct {
	params
	scope  sdk.ScopeAccessor
	log    sdk.Logger
	done   chan bool
	events chan notify.EventInfo
	dead   *sync.Mutex
}

func newPlugin(s sdk.ScopeAccessor, p params, l sdk.Logger) (*Plugin, error) {
	return &Plugin{
		params: p,
		scope:  s,
		log:    l,
		done:   make(chan bool),
	}, nil
}

// Call starts watch plugin
func (p *Plugin) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) error {
	p.events = make(chan notify.EventInfo, 1)
	if err := notify.Watch(p.Path, p.events, notify.All); err != nil {
		return fmt.Errorf("failed to initialize watcher for '%s': %s", p.Path, err)
	}

	p.dead = &sync.Mutex{}
	childCtx := ctx.ChildContext()
	defer func() {
		notify.Stop(p.events)
		childCtx.Cancel()
		p.log.Debug("watcher removed")
	}()

	// Start file watcher
	go func() {
		interval := p.DebounceTime.ToDuration()
		t := time.NewTimer(interval) // Debounce timer

		for {
			select {
			case event, ok := <-p.events:
				if !ok {
					return
				}
				fPath := event.Path()
				ignored, err := p.pathIgnored(fPath)
				if err != nil {
					p.log.Errorf("path ignore check failed: %s", err)
					continue
				}

				if !ignored {
					p.log.Debugf("event: %v %s", event.Event(), fPath)
					t.Reset(interval)
				}
			case <-t.C:
				// Re-start job when timer ends.
				p.log.Debug("timer ended")

				if childCtx.IsAlive() {
					childCtx.Cancel()
				}
				childCtx = ctx.ChildContext()
				go p.invokeJob(childCtx, r)
			}
		}
	}()

	p.log.Infof("watcher is watching for changes in '%s'", p.Path)
	<-p.done
	return nil
}

func (p *Plugin) invokeJob(ctx sdk.JobContextAccessor, r sdk.JobRunner) {
	p.log.Debug("wait until previous process stops")
	p.dead.Lock()
	// override errors channel
	jctx := ctx.(*job.RunContext)
	jctx.Error = make(chan error, 1)

	description := p.Job.FormatDescription()
	p.log.Infof("- Starting '%s'", description)
	r.RunJob(*p.Job, ctx)
	select {
	case err := <-ctx.Errors():
		p.dead.Unlock()
		if err != nil {
			p.log.Errorf("- '%s' failed: %s", description, err)
			return
		}

		p.log.Successf("- '%s' finished", description)
	}
}

// Cancel stops watch plugin
func (p *Plugin) Cancel(ctx sdk.JobContextAccessor) error {
	p.done <- true
	notify.Stop(p.events)
	p.log.Debug("watcher removed")
	return nil
}
