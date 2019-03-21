#Hysteria - Hystrix Wrapper

Hysteria enables:
- Specify errors to trip circuit breaker
- HTTP GET/POST operation with trip trigger when obtain HTTP 5xx
- Better timeout handling for HTTP operation


## Usage

Several usage of hysteria for generic and http operation

### Configuration:

```aidl
hysteria.Configure("do.something", &hysteria.Config{
    MaxConcurrency:   200,
    ErrorThreshold:   2,
    Timeout:          10,
    TriggeringErrors: []error{errors.New("something"), errors.New("another")},
})
```

Here, the circuit will be polled to trigger the trip state when obtaining specified errors in `TriggeringErrors` array

### General Operation

```aidl
err := hysteria.Exec("do.something", func() error {
    return errors.New("something")
})
```

This operation above will get circuit open if it is executed several times 

### HTTP Operation

1. GET Operation

```aidl
hysteria.Configure("do.get", &hysteria.Config{
    MaxConcurrency: 200,
    ErrorThreshold: 2,
    Timeout:        10000,
})
url := "https://httpstat.us/500"
htimeout := time.Second

req := hysteria.NewRequest(url, http.MethodGet, nil, &htimeout, nil)
resp, body, err := hysteria.ExecHTTPCtx(context.Background(), "do.get", req)
```

The operation above will trigger circuit open eventually since, the HTTP get 500.

2. POST Operation

```aidl
url := "http://dummy.restapiexample.com/api/v1/create"
hysteria.Configure("do.post", &hysteria.Config{
    MaxConcurrency: 200,
    ErrorThreshold: 2,
    Timeout:        10000,
})
htimeout := time.Second / 10
req := hysteria.NewRequest(url, http.MethodPost, Data{Name: "Faris Muhammad", Salary: "20000", Age: "28"}, &htimeout, nil)
resp, body, err := hysteria.ExecHTTPCtx(context.Background(), "do.post", req)
```

This operation above will likely get HTTP timeout (not hystrix timeout), and eventually will get circuit trip if occurrence happens too often.