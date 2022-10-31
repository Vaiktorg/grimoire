package main

type State string
type Trigger string

type IState interface {
	State() State
	Allow(trigger Trigger, state State)
	OnEntry(State) error
	OnExit(State) error
	OnFail(error)
	OnSuccess()
	Fallback(fallbackState State)
}

type StateConfig struct {
	allowable map[Trigger]State
	onEntry   func(State) error
	onExit    func(State) error
	onSuccess func()
	onFail    func(error)
	fallback  State
}

func (c *StateConfig) Allow(trigger Trigger, state State) *StateConfig {
	if _, ok := c.allowable[trigger]; !ok {
		c.allowable[trigger] = state
	}
	return c
}
func (c *StateConfig) OnEntry(onEntry func(State) error) *StateConfig {
	c.onEntry = onEntry
	return c
}
func (c *StateConfig) OnSuccess(onSuccess func()) *StateConfig {
	c.onSuccess = onSuccess
	return c
}
func (c *StateConfig) OnFail(onFail func(err error)) *StateConfig {
	c.onFail = onFail
	return c
}
func (c *StateConfig) OnExit(onExit func(State) error) *StateConfig {
	c.onExit = onExit
	return c
}
func (c *StateConfig) Fallback(fallbackState State) *StateConfig {
	c.fallback = fallbackState
	return c
}
