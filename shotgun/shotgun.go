package shotgun

import (
	"fmt"
	"github.com/romanoff/fsmonitor"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Shotgun struct {
	addr, cmdAddr string
	runner        *Runner
	path          string
	timeout       time.Duration
}

func New(srcPort, cmdPort uint, cmds []string, path string) (*Shotgun, error) {
	runner, err := NewRunner(cmds)
	if err != nil {
		return nil, err
	}

	return &Shotgun{
		addr:    fmt.Sprintf(":%d", srcPort),
		cmdAddr: fmt.Sprintf("127.0.0.1:%d", cmdPort),
		runner:  runner,
		path:    path,
		timeout: 10 * time.Second,
	}, nil
}

func (s *Shotgun) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

func (s *Shotgun) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var message string
	u := new(url.URL)
	*u = *(r.URL)
	u.Host = s.cmdAddr
	u.Scheme = "http"

	err := s.runner.CheckRestart()
	if err != nil {
		message = fmt.Sprintf("restart failed: %v", err)
		log.Println(message)
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	if err != nil {
		message = fmt.Sprintf("new request failed: %v", err)
		log.Println(message)
		http.Error(w, message, http.StatusBadGateway)
		return
	}
	req.Header = r.Header

	type Result struct {
		response *http.Response
		error    error
	}

	timeout := time.After(s.timeout)
	result := make(chan *Result)

	go func() {
		var last_error error
		for {
			select {
			case <-timeout:
				result <- &Result{nil, last_error}
				return
			default:
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				last_error = err
				time.Sleep(time.Millisecond * 10)
				continue
			}

			result <- &Result{res, nil}
		}
	}()

	res := <-result

	if res.response == nil {
		message = fmt.Sprintf(res.error.Error())
		log.Println(message)
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	defer res.response.Body.Close()

	for k, v := range res.response.Header {
		for _, p := range v {
			w.Header().Add(k, p)
		}
	}
	w.WriteHeader(res.response.StatusCode)
	io.Copy(w, res.response.Body)
}

func (s *Shotgun) Run() error {
	w, err := fsmonitor.NewWatcher()
	if err != nil {
		return err
	}

	w.Watch(s.path)

	go func() {
		for {
			select {
			case <-w.Event:
				s.runner.SetNeedRestart()
			case err := <-w.Error:
				log.Printf("fs error: %s\n", err)
			}
		}
	}()

	return http.ListenAndServe(s.addr, s)
}
