package redis

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/qp/go"
	"github.com/stretchr/pat/stop"
)

// requests implements the Transport interface using
// Redis as the underlying transport technology.
type requests struct {
	pool      *redis.Pool
	callback  qp.MessageFunc
	listeners []string
	lock      sync.Mutex
	once      sync.Once
	kill      chan struct{}
	killed    chan struct{}
	running   uint32
	timeout   time.Duration
}

// NewReqTransport gets a new qp.RequestTransport.
func NewReqTransport(url string) qp.RequestTransport {
	var pool = &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", url, 1*time.Second, 0, 0)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	return &requests{
		pool:      pool,
		listeners: []string{},
		kill:      make(chan struct{}),
		killed:    make(chan struct{}),
		timeout:   5 * time.Second,
	}
}

// Send sends a message to requests and returns an error
// if there was a problem sending.
func (r *requests) Send(to string, data []byte) error {
	if atomic.LoadUint32(&r.running) == 0 {
		return qp.ErrTransportStopped
	}
	conn := r.pool.Get()
	_, err := conn.Do("LPUSH", to, data)
	conn.Close()
	return err
}

// ListenFor instructs the transport to register for
// messages on the given channel.
func (r *requests) ListenFor(channel string) {
	r.lock.Lock()
	r.listeners = append(r.listeners, channel)
	r.lock.Unlock()
}

// OnMessage sets the callback function to be called whenever
// a message is received.
func (r *requests) OnMessage(messageFunc qp.MessageFunc) {
	r.callback = messageFunc
}

// Start spins up the requests transport processing system
// and begins processing messages.
func (r *requests) Start() error {
	r.once.Do(func() {
		go r.processMessages()
		atomic.StoreUint32(&r.running, 1)
	})
	return nil
}

// Stop spins down the requests transport processing system
func (r *requests) Stop() <-chan struct{} {
	fmt.Println("STOP")
	c := stop.Make()
	go func() {
		c <- stop.Done
		fmt.Println("above kill")
		close(r.kill)
		<-r.killed
		fmt.Println("killed came back, resetting")
		r.kill = make(chan struct{})
		r.killed = make(chan struct{})
		r.once = sync.Once{}
		atomic.StoreUint32(&r.running, 0)
		fmt.Println("sending stop done for stop chan", c)
		//c <- stop.Done
		fmt.Println("finished sending stop")
	}()
	fmt.Println("returning stop chan", c)
	return c
}

// SetTimeout sets the timeout to the given value.
// This timeout is used when gracefully shutting down the
// transport. In-flight requests will have this much time
// to complete before being abandoned.
// The default timeout value is 5 seconds.
func (r *requests) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *requests) processMessages() {
	r.lock.Lock()
	listeners := r.listeners
	r.lock.Unlock()
	fmt.Println("pmesgs")
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(listeners))
		fmt.Println("about to listen on", listeners)
		for _, c := range listeners {
			fmt.Println("listening on ", c)
			go func(channel string) {
				var data []byte
				done := make(chan struct{})
				for {
					conn := r.pool.Get()
					select {
					case <-r.kill:
						fmt.Println("got kill")
						r.kill = nil
						go func() {
							time.Sleep(r.timeout)
							fmt.Println("sending done")
							done <- struct{}{}
						}()
					case <-done:
						conn.Close()
						done = nil
						fmt.Println("wg done")
						wg.Done()
						return
					default:
						// BRPOP on the channel to wait for a new message
						message, err := redis.Values(conn.Do("BRPOP", channel, "1"))
						log.Printf("ERROR FROM VALUES: %#v", err)
						if err != nil {
							// Did we get a timeout? That's fine. Continue.
							if err == redis.ErrNil {
								continue
							}
							// or a network timeout... also fine
							if netErr, ok := err.(net.Error); ok {
								if netErr.Timeout() {
									continue
								}
							}
							// Not a timeout? Something went wrong.
							// TODO: Log this out.. maybe fire a metric to the logging endpoint
							fmt.Println("Error trying to BRPOP", err)
							conn.Close()
							return
						}
						if _, err := redis.Scan(message, &channel, &data); err != nil {
							// there was an error decoding the message into the data field
							// TODO: Log this out.. maybe fire a metric to the logging endpoint
							fmt.Println("Error trying to Scan", err)
							continue
						}
						go r.callback(&qp.Message{Source: channel, Data: data})
						conn.Close()
					}
				}
			}(c)
		}
		wg.Wait()
		fmt.Println("sending killed")
		r.killed <- struct{}{}
	}()
}
