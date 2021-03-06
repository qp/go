package redis

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
	"github.com/stretchr/pat/sleep"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/slog"
)

// PubSub represents a qp.PubSubTransport.
type PubSub struct {
	pool     *redis.Pool
	handlers map[string]qp.Handler
	lock     sync.Mutex
	running  uint32
	shutdown chan qp.Signal
	stopChan chan stop.Signal
	log      slog.Logger
}

// ensure the interface is satisfied
var _ qp.PubSubTransport = (*PubSub)(nil)

// NewPubSub makes a new PubSub redis transport.
func NewPubSub(url string) *PubSub {
	return NewPubSubTimeout(url, 1*time.Second, 1*time.Second, 1*time.Second)
}

// NewPubSubTimeout makes a new PubSub redis transport and allows you to specify timeout values.
func NewPubSubTimeout(url string, connectTimeout, readTimeout, writeTimeout time.Duration) *PubSub {
	var pool = &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", url, connectTimeout, 0, writeTimeout)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	p := &PubSub{
		pool:     pool,
		handlers: make(map[string]qp.Handler),
		shutdown: make(chan qp.Signal),
		stopChan: stop.Make(),
		log:      slog.NilLogger,
	}
	return p
}

// SetLogger sets the Logger to log to.
func (p *PubSub) SetLogger(log slog.Logger) {
	p.log = log
}

// Publish publishes data on the specified channel.
func (p *PubSub) Publish(channel string, data []byte) error {
	if atomic.LoadUint32(&p.running) == 0 {
		return qp.ErrNotRunning
	}
	if p.log.Info() {
		p.log.Info("publish to", channel, string(data))
	}
	conn := p.pool.Get()
	_, err := conn.Do("PUBLISH", channel, data)
	conn.Close()
	if err != nil && p.log.Err() {
		p.log.Err("publish failed", err)
	}
	return err
}

// Subscribe binds the handler to the specified channel.
func (p *PubSub) Subscribe(channel string, handler qp.Handler) error {
	if atomic.LoadUint32(&p.running) == 1 {
		return qp.ErrRunning
	}
	if p.log.Info() {
		p.log.Info("subscribing to", channel)
	}
	p.lock.Lock()
	p.handlers[channel] = handler
	p.lock.Unlock()
	return nil
}

func (p *PubSub) processMessages() {
	go func() {
		for c, h := range p.handlers {
			go func(channel string, handler qp.Handler) {
				sleeper := sleep.New()
				sleeper.Add(1*time.Minute, 1*time.Second)
				sleeper.Add(5*time.Minute, 10*time.Second)
				sleeper.Add(10*time.Minute, 30*time.Second)

				var conn redis.Conn
				var psc redis.PubSubConn
				closed := false

				go func() {
					<-p.shutdown
					if p.log.Info() {
						p.log.Info("received shutdown signal - shutting down")
					}
					closed = true
					psc.Close()
				}()
				for {
					conn = p.pool.Get()
					psc = redis.PubSubConn{Conn: conn}
					psc.PSubscribe(channel)
					switch v := psc.Receive().(type) {
					case redis.PMessage:
						if sleeper.Reset() {
							if p.log.Warn() {
								p.log.Warn("reconnected to redis after interruption")
							}
						}
						if p.log.Info() {
							p.log.Info("handling message from", v.Channel+":", string(v.Data))
						}
						go handler.Handle(&qp.Message{Source: v.Channel, Data: v.Data})
					case error, net.Error:
						if closed {
							return
						}
						if p.log.Warn() {
							p.log.Warn("error when receiving from redis:", v)
						}
						if sleeper.Sleep() == sleep.Abort {
							if p.log.Err() {
								p.log.Err("unable to connect to redis - aborting:", v)
							}
							return
						}
					}
				}
			}(c, h)
		}
	}()
}

// Start starts the transport.
func (p *PubSub) Start() error {
	// TODO: discuss blocking this until all subscription
	// acks are received?
	if atomic.LoadUint32(&p.running) == 0 {
		atomic.StoreUint32(&p.running, 1)
		p.log.Info("starting")
		p.processMessages()
	} else {
		return qp.ErrRunning
	}
	return nil
}

// Stop stops the transport and closes StopChan() when finished.
func (p *PubSub) Stop(grace time.Duration) {
	if p.shutdown == nil {
		return
	}
	p.log.Info("stopping...")
	// stop processing new Publish calls
	atomic.StoreUint32(&p.running, 0)
	// instruct all listening goroutines to shutdown
	close(p.shutdown)
	p.shutdown = nil
	// wait for duration to allow in-flight requests to finish
	time.Sleep(grace)
	// inform caller of stop complete
	close(p.stopChan)
	p.log.Info("stopped")
}

// StopChan gets the stop channel which will be closed when
// this transport has successfully stopped.
func (p *PubSub) StopChan() <-chan stop.Signal {
	return p.stopChan
}
