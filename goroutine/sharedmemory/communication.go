package main

import (
	"log"
	"net/http"
	"time"
)

var urls = []string{
	"http://xx.yy.z.unknown",
	"http://www.baidu.com",
	"http://rodbate.github.io/blog",
}

type state struct {
	url    string
	status string
}

func stateMonitor(interval time.Duration) chan<- *state {
	stateChan := make(chan *state)
	status := make(map[string]string)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				logStatus(&status)
			case state := <-stateChan:
				status[state.url] = state.status
			}
		}
	}()
	return stateChan
}

func logStatus(status *map[string]string) {
	log.Println("Current State: ")
	for k, v := range *status {
		log.Printf(" %s, %s", k, v)
	}
}

type resource struct {
	url      string
	errorCnt int
}

func (r *resource) checkStatus() string {
	resp, err := http.Head(r.url)
	if err != nil {
		log.Printf("Error: url=%s, error=%s\n", r.url, err)
		r.errorCnt++
		return err.Error()
	}
	r.errorCnt = 0
	return resp.Status
}

func (r *resource) sleep(done chan<- *resource) {
	time.Sleep(5*time.Second + time.Duration(r.errorCnt))
	done <- r
}

func doWork(in <-chan *resource, out chan<- *resource, status chan<- *state) {
	for r := range in {
		status <- &state{
			url:    r.url,
			status: r.checkStatus(),
		}
		out <- r
	}
}

func main() {
	pending, complete := make(chan *resource), make(chan *resource)

	status := stateMonitor(5 * time.Second)

	for i := 0; i < 2; i++ {
		go doWork(pending, complete, status)
	}

	go func() {
		for _, u := range urls {
			pending <- &resource{
				url:      u,
				errorCnt: 0,
			}
		}
	}()

	for r := range complete {
		go r.sleep(pending)
	}
}
