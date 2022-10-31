package main

import (
	"errors"
	"fmt"
)

type StateMachine struct {
	configs        map[State]*StateConfig
	onStateChanged func(prev, curr State)
	prevState      State
	currState      State
}

func NewStateMachine(initialState State) *StateMachine {
	return &StateMachine{
		configs:   make(map[State]*StateConfig),
		currState: initialState,
	}
}

func (sm *StateMachine) State(newState State) *StateConfig {
	if _, ok := sm.configs[newState]; !ok {
		sm.configs[newState] = &StateConfig{allowable: make(map[Trigger]State)}
	}
	return sm.configs[newState]
}

func (sm *StateMachine) IState(newState IState) *StateConfig {
	state := newState.State()
	if _, ok := sm.configs[state]; !ok {
		sm.configs[state] = &StateConfig{
			allowable: make(map[Trigger]State),
			onEntry:   newState.OnEntry,
			onExit:    newState.OnExit,
			onSuccess: newState.OnSuccess,
			onFail:    newState.OnFail,
			fallback:  state,
		}
	}

	return sm.configs[state]
}

func (sm *StateMachine) OnStateChanged(onStateChanged func(prev, curr State)) {
	if onStateChanged == nil {
		sm.onStateChanged = func(prev, curr State) {
			fmt.Printf("State transitions from %s to %s\n", prev, curr)
		}
	}
	sm.onStateChanged = onStateChanged
}

func (sm *StateMachine) Fire(trigger Trigger) error {
	currConf, ok := sm.configs[sm.currState]
	if !ok {
		return errors.New(fmt.Sprintf("statemachine has no current state %v\n", sm.currState))
	}

	nextState, ok := currConf.allowable[trigger]
	if !ok {
		return errors.New(fmt.Sprintf("trigger %s not allowed...\n", trigger))
	}

	if err := currConf.onExit(sm.currState); err != nil {
		return err
	}

	newConf, ok := sm.configs[nextState]
	if !ok {
		return errors.New(fmt.Sprintf("no configured state %s\n", trigger))
	}

	if newConf.onEntry == nil {
		return errors.New(fmt.Sprintf("no state behavior set for triggered state %s.\n", nextState))
	}

	if err := newConf.onEntry(nextState); err != nil {
		if newConf.onFail != nil {
			newConf.onFail(err)
		}

		if newConf.fallback == "" {
			return err
		}

		newConf, ok = sm.configs[newConf.fallback]
		if !ok {
			return errors.New(fmt.Sprintf("No configured state %s\n", trigger))
		}
	}

	sm.onStateChanged(sm.currState, nextState)
	sm.prevState = sm.currState
	sm.currState = nextState

	if newConf.onSuccess != nil {
		newConf.onSuccess()
	}

	return nil
}

func (sm *StateMachine) GetCurrentState() State {
	return sm.currState
}
