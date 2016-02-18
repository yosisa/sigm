package sigm

import (
	"os"
	"os/signal"
	"sync"
)

var defaultDispatcher = new(dispatcher)

func Handle(sig os.Signal, f func()) {
	defaultDispatcher.handle(sig, f)
}

type Register struct {
	sigs []os.Signal
}

func (r *Register) Handle(f func()) {
	for _, s := range r.sigs {
		Handle(s, f)
	}
}

func For(sigs ...os.Signal) *Register {
	return &Register{sigs}
}

type dispatcher struct {
	c     chan os.Signal
	funcs map[os.Signal][]func()
	m     sync.RWMutex
}

func (d *dispatcher) init() {
	d.c = make(chan os.Signal, 10)
	d.funcs = make(map[os.Signal][]func())
	go d.dispatch()
}

func (d *dispatcher) dispatch() {
	for sig := range d.c {
		d.m.RLock()
		funcs, _ := d.funcs[sig]
		d.m.RUnlock()
		for _, f := range funcs {
			f()
		}
	}
}

func (d *dispatcher) handle(sig os.Signal, f func()) {
	d.m.Lock()
	defer d.m.Unlock()
	if d.funcs == nil {
		d.init()
	}
	if _, ok := d.funcs[sig]; !ok {
		signal.Notify(d.c, sig)
	}
	d.funcs[sig] = append(d.funcs[sig], f)
}
