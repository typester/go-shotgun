package shotgun

import (
	"fmt"
	"github.com/romanoff/fsmonitor"
	"io"
	"net/http"
	"net/url"
	"os"
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
	u := new(url.URL)
	*u = *(r.URL)
	u.Host = s.cmdAddr
	u.Scheme = "http"

	s.runner.CheckRestart()

	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	if err != nil {
		fmt.Println("new request failed: ", err)
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
		fmt.Fprintf(os.Stderr, "Error: %s\n", res.error.Error())
		w.Write([]byte("timeout"))
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
				fmt.Fprintf(os.Stderr, "fs error: %s\n", err)
			}
		}
	}()

	return http.ListenAndServe(s.addr, s)
}
