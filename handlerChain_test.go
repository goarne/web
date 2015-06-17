package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestOn6ChainedHandlers(t *testing.T) {
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/testblog", nil)

	h := HandlerChain{}
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})

	h.ServeHTTP(resp, req)

	if h.numExc != 6 {
		t.Errorf("HandlerChain only executed %v times", h.numExc)
	}
}

func TestStoppedStateChainedHandlers(t *testing.T) {
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/testblog", nil)

	h := HandlerChain{}
	h.Add(MockHandlerReady{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerStopped{})
	h.Add(MockHandlerActive{})
	h.Add(MockHandlerActive{})

	h.ServeHTTP(resp, req)

	if h.numExc != 3 {
		t.Errorf("HandlerChain only executed %v times", h.numExc)
	}
}

type MockHandlerReady struct{}

func (MockHandlerReady) ServeReq(c *ChainedReq) {
	c.State = READY
}

type MockHandlerActive struct{}

func (MockHandlerActive) ServeReq(c *ChainedReq) {
	c.State = ACTIVE
}

type MockHandlerStopped struct{}

func (MockHandlerStopped) ServeReq(c *ChainedReq) {
	c.State = STOPPED
}

type MockHandlerFinnished struct{}

func (MockHandlerFinnished) ServeReq(c *ChainedReq) {
	c.State = FINNISHED
}
