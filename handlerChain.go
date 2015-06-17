//HandlerChain and ChainedHandler for creating chains of Handlers.
package web

import (
	"net/http"
)

//The HandlerChain handles a list of ChainedHandlers.
//Each of the ChainedHandlers executes ServeHTTP on an incomming request.
//It can be used to pre or/and post -process incomming requests.
//The list of ChainedHandlers are handled by the HandlerChain.
//The ChainedHandler executes on the http request. The http request is wrapped in
// a ChainedReq instance and passed throud the list.
//A request contains stateinfo which a handler can use to signal the
// HandlerChain to stop execution on the request.
//
//The HandlerChain holds a list with ChainedHandlers and receives
//the requst which it passes through the list.
//
//Ex.
//	hc := &web.HandlerChain{}
//
//	hc.Add(handler.NewStaticHtmlHandler("index.html"))
//	hc.Add(handler.NewFileHandler("./html/"))
//	r.Handler(hc)
type HandlerChain struct {
	numExc   int
	handlers []ChainedHandler
}

//Adds a ChainedHndler to the list.
func (h *HandlerChain) Add(c ChainedHandler) {
	h.handlers = append(h.handlers, c)
}

//The method executes the chained handlers on the inncomming request.
//When receiving a request, its wrapped in a ChainedReq object and passed
//to a ChainedHandler to be processed. Its passed to every ChainedHandler in
//the list, or until a ChainedHandler marks the ChainedReq.State to stopped.
func (h *HandlerChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	chainedReq := &ChainedReq{w, r, READY}

	for i, c := range h.handlers {

		h.numExc = i + 1

		c.ServeReq(chainedReq)

		if chainedReq.State == STOPPED {
			break
		}
	}

	chainedReq.State = FINNISHED
}

//The ChainedReq wraps the http request received by the HandlerChain .
//It contains request, response and state for the chainhandler.
//State is used by the chainhandler to check if execution should stop.
type ChainedReq struct {
	Resp  http.ResponseWriter
	Req   *http.Request
	State ReqState
}

//State defined for a chainedReq.
type ReqState int

const (
	READY ReqState = 0 + iota
	ACTIVE
	FINNISHED
	STOPPED
)

//Interface which must be implemented by a handler for handling the request.
type ChainedHandler interface {
	ServeReq(c *ChainedReq)
}
