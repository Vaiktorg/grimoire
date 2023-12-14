package events

import (
	"context"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve/ws"
	"sync"
	"time"
)

// --------------------------------------------------

type Context[T any] struct {
	ctx    context.Context
	cancel func()
}

func NewContext[T any](parent context.Context, value T) context.Context {
	ctx := context.WithValue(parent, "init", value)

	return &Context[T]{
		ctx: ctx,
	}
}

func (c *Context[T]) Cancel() {
	c.cancel()
}

func (c *Context[T]) Deadline() (deadline time.Time, ok bool) { return c.ctx.Deadline() }
func (c *Context[T]) Done() <-chan struct{}                   { return c.ctx.Done() }
func (c *Context[T]) Err() error                              { return c.ctx.Err() }
func (c *Context[T]) Value(key any) any                       { return c.ctx.Value(key) }

// --------------------------------------------------

type IAction interface {
	Key() string
	Validate() error
	Handler(ctx context.Context, client *ws.Client)
}

type Config struct {
	Logger log.ILogger
	Ctx    context.Context
}

func (c *Config) GetContext() context.Context {
	if c.Ctx == nil {
		return context.Background()
	}
	return c.Ctx
}
func (c *Config) GetLogger() log.ILogger {
	if c.Logger == nil {
		return log.NewLogger(log.Config{
			CanPrint:    true,
			CanOutput:   true,
			ServiceName: "Events",
		})
	}

	return c.Logger
}

type Events struct {
	mu sync.Mutex
	wg sync.WaitGroup

	evs map[string]IAction // AutoHandler Actions

	l   log.ILogger
	ctx context.Context
}

func NewEvents(config *Config) *Events {
	ec := &Events{
		l:   config.GetLogger(),
		evs: make(map[string]IAction),
		ctx: config.GetContext(),
	}

	return ec
}

func (e *Events) Register(actions ...IAction) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, action := range actions {
		k := action.Key()
		if _, ok := e.evs[k]; !ok {
			e.evs[k] = action
		}
	}
}

func (e *Events) Listen(client *ws.Client) {
	if client == nil {
		panic("client for event is nil")
	}

	for msg := range client.ReaderChan {
		e.wg.Add(1)
		go e.procMsg(msg, client)
	}

	// When client closes ReaderChan
	// ForRange exits and waits
	// for last action to finishs
	e.wg.Wait()
}

func (e *Events) procMsg(msg *ws.Message, client *ws.Client) {
	e.mu.Lock()
	action, ok := e.evs[msg.KEY]
	e.mu.Unlock()

	if !ok {
		return
	}

	defer e.wg.Done()

	if err := action.Validate(); err != nil {
		client.Error(msg.KEY, err)
	}

	// Automatic Handler
	go action.Handler(e.ctx, client)
}
