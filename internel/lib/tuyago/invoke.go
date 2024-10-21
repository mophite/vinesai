package tuyago

import (
	"fmt"
	"sync"
	"vinesai/internel/ava"
	"vinesai/internel/x"
)

type Handler interface {
	Call(c *ava.Context)
}

var dispatch = &dispatcher{
	mux:      new(sync.RWMutex),
	handlers: make(map[string]Handler),
}

type dispatcher struct {
	mux      *sync.RWMutex
	handlers map[string]Handler
}

func (d *dispatcher) getHandler(name string) (Handler, error) {
	d.mux.RLock()
	h, ok := d.handlers[name]
	if !ok {
		d.mux.RUnlock()
		return nil, fmt.Errorf("handler %q not found", name)
	}

	d.mux.RUnlock()

	return h, nil
}

func Register(name string, h Handler) {
	dispatch.mux.Lock()
	_, ok := dispatch.handlers[name]
	if ok {
		dispatch.mux.Unlock()
		panic("handler " + name + "already exist")
	}
	dispatch.handlers[name] = h
	dispatch.mux.Unlock()
}

func invoke(c *ava.Context, bizCode string, data []byte) {
	//调用函数业务逻辑处理
	h, err := dispatch.getHandler(bizCode)
	if err != nil {
		ava.Error(err)
		return
	}

	err = x.MustUnmarshal(data, h)
	if err != nil {
		ava.Error(err)
		return
	}

	h.Call(c)
}
