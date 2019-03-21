package main

import (
	"context"
	"errors"
	"log"
	"time"

	"net/http"

	"github.com/syariatifaris/hysteria"
)

type Data struct {
	Name   string `json:"name"`
	Salary string `json:"salary"`
	Age    string `json:"age"`
}

func main() {
	get()
	//post()
	//general()
}

func general() {
	hysteria.Configure("do.something", &hysteria.Config{
		MaxConcurrency:   200,
		ErrorThreshold:   2,
		Timeout:          10,
		TriggeringErrors: []error{errors.New("something"), errors.New("another")},
	})
	hysteria.Configure("do.something.else", &hysteria.Config{
		MaxConcurrency:   200,
		ErrorThreshold:   2,
		Timeout:          10,
		PollTripOnError:  true,
		TriggeringErrors: []error{errors.New("another"), errors.New("something else")},
	})
	for i := 0; i < 100; i++ {
		err := hysteria.Exec("do.something", func() error {
			return errors.New("something")
		})
		log.Println("error:", err)
	}
	log.Println("--------------------")
	for i := 0; i < 100; i++ {
		err := hysteria.Exec("do.something.else", func() error {
			return errors.New("something else")
		})
		log.Println("error:", err)
	}
}

func get() {
	url := "https://httpstat.us/500"
	hysteria.Configure("do.get", &hysteria.Config{
		MaxConcurrency: 200,
		ErrorThreshold: 2,
		Timeout:        10000,
	})
	for i := 1; i < 20; i++ {
		for i := 1; i < 10; i++ {
			go func() {
				htimeout := time.Second / 10
				req := hysteria.NewRequest(url, http.MethodGet, nil, &htimeout, nil)
				_, _, err := hysteria.ExecHTTPCtx(context.Background(), "do.get", req)
				log.Println(err)
			}()
		}
		time.Sleep(time.Second / 3)
	}
	time.Sleep(time.Second * 10)
}

func post() {
	url := "http://dummy.restapiexample.com/api/v1/create"
	hysteria.Configure("do.post", &hysteria.Config{
		MaxConcurrency: 200,
		ErrorThreshold: 2,
		Timeout:        10000,
	})
	htimeout := time.Second / 2
	req := hysteria.NewRequest(url, http.MethodPost, Data{Name: "Faris Muhammad", Salary: "20000", Age: "28"}, &htimeout, nil)
	_, _, err := hysteria.ExecHTTPCtx(context.Background(), "do.post", req)
	log.Println(err)
}
