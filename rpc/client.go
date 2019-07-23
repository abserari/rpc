package rpc

import (
	"context"
	"log"
	"net"
	"sync"
	"time"
)

// dialOptions configure a Dial call And are set by
// the DialOptions values passed to Dial
// DialOptions do func - apply.
type dialOptions struct {
	timeout time.Duration
}

// DialOption configure how we set up the connection
type DialOption interface {
	apply(*dialOptions)
}

// EmptyDialOption does not alter the dial configuration. It can be embedded in
// another structure to build custom dial options.
//
// This API is EXPERIMENTAL.
type EmptyDialOption struct{}

func (EmptyDialOption) apply(*dialOptions) {}

// funcDialOption wraps a function that modifies dialOptions into an
// implementation of the DialOption interface.
type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}

func defaultDialOptions() dialOptions {
	return dialOptions{
		timeout: time.Second * 3,
	}
}

// Client connect to Server
// It makes by Dial and dialOptions.
type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	conn  net.Conn
	dopts dialOptions

	mu sync.RWMutex
}

// Dial creates a new Client with DialOptions
func Dial(addr string, opts ...DialOption) (*Client, error) {
	return DialContext(context.Background(), addr, opts...)
}

// DialContext to create a connect to target on server .
// DialOption configure Dial.
func DialContext(ctx context.Context, addr string, opts ...DialOption) (c *Client, err error) {
	c = &Client{
		dopts: defaultDialOptions(),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	for _, opt := range opts {
		opt.apply(&c.dopts)
	}

	if c.dopts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.dopts.timeout)
		defer cancel()
	}
	defer func() {
		select {
		case <-ctx.Done():
			c, err = nil, ctx.Err()
		default:
		}
	}()

	c.conn, err = net.Dial("tcp", addr)

	return c, nil
}
func (client *Client) send(ctx context.Context,call *Call){

}
	// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(ctx, call)
	return call
}

func (client *Client) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	Done := client.Go(ctx, servicePath, serviceMethod, args, reply, make(chan *Call, 1)).Done

	var err error
	select {
	case <-ctx.Done(): //cancel by context
		return ctx.Err()
	case call := <-Done:
		err = call.Error
	}

	return err
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *Client) Close() error {
	client.mu.Lock()
	err := client.conn.Close()
	client.mu.Unlock()
	return err
}
