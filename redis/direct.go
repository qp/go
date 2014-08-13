package redis

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/pat/sleep"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/slog"
)

// Direct represents a qp.DirectTransport.
type Direct struct {
	pool     *redis.Pool
	stopChan chan stop.Signal
	running  uint32
	handlers map[string]qp.Handler
	lock     sync.Mutex
	shutdown chan qp.Signal
	log      slog.Logger
}

// ensure the interface is satisfied
var _ qp.DirectTransport = (*Direct)(nil)

// NewDirect makes a new Direct redis transport.
func NewDirect(url string) *Direct {
	return NewDirectTimeout(url, 1*time.Second, 1*time.Second, 1*time.Second)
}

// NewDirectTimeout makes a new Direct redis transport and allows you to specify timeout values.
func NewDirectTimeout(url string, connectTimeout, readTimeout, writeTimeout time.Duration) *Direct {
	if readTimeout == 0 {
		readTimeout = 1 * time.Second
	}
	var pool = &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", url, connectTimeout, readTimeout, writeTimeout)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	p := &Direct{
		pool:     pool,
		handlers: make(map[string]qp.Handler),
		shutdown: make(chan qp.Signal),
		stopChan: stop.Make(),
		log:      slog.NilLogger,
	}
	return p
}

// SetLogger sets the Logger to log to.
func (d *Direct) SetLogger(log slog.Logger) {
	d.log = log
}

// Send sends data on the channel.
func (d *Direct) Send(channel string, data []byte) error {
	if atomic.LoadUint32(&d.running) == 0 {
		return qp.ErrNotRunning
	}
	if d.log.Info() {
		d.log.Info("sending to", channel, string(data))
	}
	conn := d.pool.Get()
	_, err := conn.Do("LPUSH", channel, data)
	conn.Close()
	if err != nil && d.log.Err() {
		d.log.Err("LPUSH failed", err)
	}
	return err
}

// OnMessage binds the handler to the specified channel.
func (d *Direct) OnMessage(channel string, handler qp.Handler) error {
	if atomic.LoadUint32(&d.running) == 1 {
		return qp.ErrRunning
	}
	if d.log.Info() {
		d.log.Info("listening to", channel)
	}
	d.lock.Lock()
	d.handlers[channel] = handler
	d.lock.Unlock()
	return nil
}

func (d *Direct) processMessages() {
	go func() {
		for c, h := range d.handlers {
			go func(channel string, handler qp.Handler) {
				sleeper := sleep.New()
				sleeper.Add(1*time.Minute, 1*time.Second)
				sleeper.Add(5*time.Minute, 10*time.Second)
				sleeper.Add(10*time.Minute, 30*time.Second)
				for {
					select {
					case <-d.shutdown:
						if d.log.Info() {
							d.log.Info("shutting down")
						}
						return
					default:
						conn := d.pool.Get()
						if err := d.handleMessage(conn, channel, handler); err != nil {
							if d.log.Warn() {
								d.log.Warn("failed to handle message:", err, "sleeping for", sleeper.Duration())
							}
							if sleeper.Sleep() == sleep.Abort {
								if d.log.Err() {
									d.log.Err("unable to connect to redis - aborting:", err)
								}
								return
							}
						} else {
							if sleeper.Reset() {
								if d.log.Warn() {
									d.log.Warn("reconnected to redis after interruption")
								}
							}
						}
						conn.Close()
					}
				}
			}(c, h)
		}
	}()
}

func (d *Direct) handleMessage(conn redis.Conn, channel string, handler qp.Handler) error {
	var data []byte
	// BRPOP on the channel to wait for a new message
	message, err := redis.Values(conn.Do("BRPOP", channel, "1"))
	if err != nil {
		// Did the BRPOP return with no data?
		if err == redis.ErrNil {
			return nil
		}
		// Network timeout is fine also.
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return nil
			}
		}
		return err
	}
	if _, err := redis.Scan(message, &channel, &data); err != nil {
		return err
	}
	if d.log.Info() {
		d.log.Info("handling message on", channel+":", string(data))
	}
	go handler.Handle(&qp.Message{Source: channel, Data: data})
	return nil
}

// Start starts the transport.
func (d *Direct) Start() error {
	if atomic.LoadUint32(&d.running) == 0 {
		atomic.StoreUint32(&d.running, 1)
		if d.log.Info() {
			d.log.Info("starting")
		}
		go d.processMessages()
	} else {
		return qp.ErrRunning
	}
	return nil
}

// Stop instructs the transport to gracefully stop and close the
// StopChan when stopping has completed.
//
// In-flight requests will have "wait" duration to complete
// before being abandoned.
func (d *Direct) Stop(grace time.Duration) {
	if d.log.Info() {
		d.log.Info("stopping...")
	}
	// stop processing new Sends
	atomic.StoreUint32(&d.running, 0)
	// wait for duration to allow in-flight requests to finish
	time.Sleep(grace)
	// instruct all receiving goroutines to shutdown
	close(d.shutdown)
	// inform caller of stop complete
	close(d.stopChan)
	if d.log.Info() {
		d.log.Info("stopped")
	}
}

// StopChan gets the stop channel which will block until
// stopping has completed, at which point it is closed.
// Callers should never close the stop channel.
func (d *Direct) StopChan() <-chan stop.Signal {
	return d.stopChan
}
