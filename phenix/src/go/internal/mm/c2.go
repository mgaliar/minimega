package mm

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type GroupError struct {
	Err  error
	Meta map[string]interface{}
}

func NewGroupError(err error, meta map[string]interface{}) GroupError {
	return GroupError{Err: err, Meta: meta}
}

func (this GroupError) Error() string {
	return this.Err.Error()
}

type ErrGroup struct {
	sync.Mutex     // embed
	sync.WaitGroup // embed

	Errors []GroupError
}

func (this *ErrGroup) AddError(err error, meta map[string]interface{}) {
	this.Lock()
	defer this.Unlock()

	this.Errors = append(this.Errors, NewGroupError(err, meta))
}

func (this *ErrGroup) AddGroupError(err GroupError) {
	this.Lock()
	defer this.Unlock()

	this.Errors = append(this.Errors, err)
}

type C2RetryError struct {
	Delay time.Duration
}

func (C2RetryError) Error() string {
	return "retry"
}

type C2ParallelCommand struct {
	Wait     *ErrGroup
	Options  []C2Option
	Meta     map[string]interface{}
	Expected func(string) error
}

func ScheduleC2ParallelCommand(cmd *C2ParallelCommand) {
	cmd.Wait.Add(1)

	go func() {
		defer cmd.Wait.Done()

		var (
			o  = NewC2Options(cmd.Options...)
			id string

			retryUntil = time.Now().Add(5 * time.Minute)
		)

		for {
			var err error

			id, err = ExecC2Command(cmd.Options...)
			if err != nil {
				if errors.Is(err, ErrC2ClientNotActive) {
					if time.Now().After(retryUntil) {
						cmd.Wait.AddError(fmt.Errorf("C2 client took too long to activate"), cmd.Meta)
						return
					}

					time.Sleep(5 * time.Second)
					continue
				}

				cmd.Wait.AddError(fmt.Errorf("executing command '%s': %w", o.command, err), cmd.Meta)
				return
			}

			break
		}

		opts := []C2Option{C2NS(o.ns), C2CommandID(id)}

		resp, err := WaitForC2Response(opts...)
		if err != nil {
			cmd.Wait.AddError(fmt.Errorf("getting response for command '%s': %w", o.command, err), cmd.Meta)
			return
		}

		if err := cmd.Expected(resp); err != nil {
			var retry C2RetryError

			if errors.As(err, &retry) {
				time.Sleep(retry.Delay)
				ScheduleC2ParallelCommand(cmd)
			} else {
				cmd.Wait.AddError(err, cmd.Meta)
			}
		}
	}()
}
