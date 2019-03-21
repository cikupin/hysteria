package hysteria

import (
	"context"
	"errors"
	"time"

	"encoding/json"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/parnurzeal/gorequest"
)

const DefaultTimeout = time.Second * 5

//Request custom hysteria http request
type Request struct {
	URL         string
	Method      string
	Data        interface{}
	Headers     map[string]string
	HTTPTimeout *time.Duration
}

//NewRequest create new http request
//args:
//  url: http target URL
//  method: http method (POST/GET)
//  data: POST data: will be converted as JSON
//  timeout: HTTP timeout, should be less or equal than hystrix timeout
//  headers: map of headers
//returns:
//  Request: new hysteria request
func NewRequest(url, method string, data interface{}, timeout *time.Duration, headers map[string]string) *Request {
	return &Request{
		URL:         url,
		Method:      method,
		Data:        data,
		Headers:     headers,
		HTTPTimeout: timeout,
	}
}

//ExecHTTPCtx executes http with context
//args:
//  ctx: derived context
//  cmd: command operation
//  req: http request
//returns:
//  http response: ptr of http.Response
//  string body
//  error: operation error
func ExecHTTPCtx(ctx context.Context, cmd string, req *Request) (*http.Response, string, error) {
	//temp only should use channel
	var r *http.Response
	var b string
	act := make(chan error, 1)
	echan := hystrix.Go(cmd, func() error {
		timeout := DefaultTimeout
		if req.HTTPTimeout != nil {
			timeout = *req.HTTPTimeout
		}

		hctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		rr := gorequest.New()
		var sa *gorequest.SuperAgent
		switch req.Method {
		case http.MethodPost:
			sa = rr.Post(req.URL)
		case http.MethodGet:
			sa = rr.Get(req.URL)
		}
		for k, h := range req.Headers {
			sa.Set(k, h)
		}
		if req.Data != nil && req.Method == http.MethodPost {
			bytes, err := json.Marshal(req.Data)
			if err != nil {
				act <- err
				return nil
			}
			sa.Send(string(bytes))
		}
		rchan := make(chan *http.Response, 1)
		bchan := make(chan string, 1)
		dchan := make(chan bool, 1)
		eschan := make(chan []error, 1)
		go func() {
			defer func() {
				close(rchan)
				close(bchan)
				close(eschan)
				close(dchan)
			}()
			resp, body, errs := sa.End()
			rchan <- resp
			bchan <- body
			eschan <- errs
			dchan <- true
		}()
		select {
		case <-hctx.Done():
			return errors.New("http connection timeout")
		case <-dchan:
			b = <-bchan
			r = <-rchan
			if r != nil {
				if r.StatusCode >= http.StatusInternalServerError {
					//trigger trip
					return errors.New(r.Status)
				}
			}
			errs := <-eschan
			if len(errs) > 0 {
				act <- errs[0]
			}
		}
		act <- nil
		return nil
	}, nil)
	select {
	case err := <-act:
		return r, b, err
	case err := <-echan:
		return r, b, err
	}
}
