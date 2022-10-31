package main

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

const (
	Idle      = State("Idle")
	Running   = State("Running")
	Updating  = State("Updating")
	Disposing = State("Disposing")

	Start   = Trigger("Start")
	Update  = Trigger("Update")
	Stop    = Trigger("Stop")
	Dispose = Trigger("Dispose")
)

func TestStateMachine(t *testing.T) {
	sm := NewStateMachine(Idle)

	sm.OnStateChanged(func(prev, curr State) {
		fmt.Printf("Changing states from: %s to: %s...\n", prev, curr)
	})

	configure(sm)

	t.Run("Start", func(t *testing.T) {
		err := sm.Fire("Start")
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
	})
	t.Run("Update", func(t *testing.T) {
		err := sm.Fire("Update")
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
	})
	t.Run("Stop", func(t *testing.T) {
		err := sm.Fire("Stop")
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
	})
	t.Run("Dispose", func(t *testing.T) {
		err := sm.Fire("Dispose")
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
	})

}
func Entering(state State) error {
	fmt.Println("Entering State ", state)
	return nil
}
func Exiting(state State) error {
	fmt.Println("Exiting State ", state)
	return nil
}
func Success() {
	fmt.Println("Success Changing State")
}
func configure(sm *StateMachine) {
	sm.State(Idle).
		OnEntry(Entering).
		OnSuccess(Success).
		Allow(Start, Running).
		Allow(Dispose, Disposing).
		OnExit(Exiting)

	sm.State(Running).
		OnEntry(Entering).
		OnSuccess(Success).
		Allow(Stop, Idle).
		Allow(Update, Updating).
		OnExit(Exiting)

	sm.State(Updating).
		OnEntry(Entering).
		OnSuccess(Success).
		Allow(Stop, Idle).
		Allow(Dispose, Disposing).
		OnExit(Exiting)

	sm.State(Disposing).
		OnEntry(Entering).
		OnSuccess(Success)
}
