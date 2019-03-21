package hysteria

import (
	"sync"

	"github.com/afex/hystrix-go/hystrix"
)

var g *glob
var once sync.Once

//Config custom hystrix configuration
type Config struct {
	MaxConcurrency   int
	ErrorThreshold   int
	Timeout          int
	SleepWindow      int
	TriggeringErrors []error
	PollTripOnError  bool
}

type conf struct {
	errs      []error
	pollOnErr bool
}

type glob struct {
	confs map[string]*conf
	mux   sync.Mutex
}

func (g *glob) exists(cmd string, err error) bool {
	g.mux.Lock()
	defer g.mux.Unlock()
	if err == nil {
		return false
	}
	if conf, ok := g.confs[cmd]; ok {
		if conf.pollOnErr {
			return true
		}
		if len(conf.errs) == 0 {
			return false
		}
		for _, e := range conf.errs {
			if e == err || e.Error() == err.Error() {
				return true
			}
		}
	}
	return false
}

//Configure configures hysteria command as hystrix derived command operation
//args:
//  cmd: operation command: string
//  cfg: hysteria command: Config ptr
//returns:
//  void
func Configure(cmd string, cfg *Config) {
	once.Do(func() {
		g = &glob{
			confs: make(map[string]*conf, 1),
		}
	})
	hystrix.ConfigureCommand(cmd, hystrix.CommandConfig{
		Timeout:               cfg.Timeout,
		ErrorPercentThreshold: cfg.ErrorThreshold,
		MaxConcurrentRequests: cfg.MaxConcurrency,
	})
	g.mux.Lock()
	defer g.mux.Unlock()
	if _, ok := g.confs[cmd]; !ok {
		g.confs[cmd] = new(conf)
	}
	if cfg.TriggeringErrors != nil {
		g.confs[cmd].errs = cfg.TriggeringErrors
	}
	g.confs[cmd].pollOnErr = cfg.PollTripOnError
}

// ConfigureMany applies settings for a set of circuit configurations
//args:
//  cmds: hysteria commands: map[string]*Config
//returns:
//  void
func ConfigureMany(cmds map[string]*Config) {
	for k, v := range cmds {
		Configure(k, v)
	}
}

//Exec executes hystrix general operation
//args:
//  cmd: hysteria command: string
//  fn: anonymous function for execution stub: func error
//returns:
//  error: operation error
func Exec(cmd string, fn func() error) error {
	act := make(chan error, 1)
	echan := hystrix.Go(cmd, func() error {
		err := fn()
		if g.exists(cmd, err) {
			return err
		}
		act <- err
		return nil
	}, nil)
	select {
	case err := <-act:
		return err
	case err := <-echan:
		return err
	}
}
