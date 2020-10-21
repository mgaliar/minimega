package mm

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type GroupError struct {
	Err  error
	Args []interface{}
}

func NewGroupError(err error, args ...interface{}) GroupError {
	return GroupError{Err: err, Args: args}
}

func (this GroupError) Error() string {
	return this.Err.Error()
}

type ErrGroup struct {
	sync.Mutex     // embed
	sync.WaitGroup // embed

	Errors []GroupError
}

func (this *ErrGroup) AddError(err error, args ...interface{}) {
	this.Lock()
	defer this.Unlock()

	this.Errors = append(this.Errors, GroupError{Err: err, Args: args})
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
	Expected func(string) error
}

func ScheduleC2ParallelCommand(cmd *C2ParallelCommand) {
	cmd.Wait.Add(1)

	go func() {
		defer cmd.Wait.Done()

		var (
			o  = NewC2Options(cmd.Options...)
			id string
		)

		for {
			var err error

			id, err = ExecC2Command(cmd.Options...)
			if err != nil {
				if errors.Is(err, ErrC2ClientNotActive) {
					time.Sleep(1 * time.Second)
					continue
				}

				var groupError GroupError

				if errors.As(err, &groupError) {
					cmd.Wait.AddGroupError(groupError)
				} else {
					cmd.Wait.AddError(fmt.Errorf("executing command '%s': %w", o.command, err))
				}

				return
			}

			break
		}

		opts := []C2Option{C2NS(o.ns), C2CommandID(id)}

		resp, err := WaitForC2Response(opts...)
		if err != nil {
			var groupError GroupError

			if errors.As(err, &groupError) {
				cmd.Wait.AddGroupError(groupError)
			} else {
				cmd.Wait.AddError(fmt.Errorf("getting response for command '%s': %w", o.command, err))
			}

			return
		}

		if err := cmd.Expected(resp); err != nil {
			var (
				retryError C2RetryError
				groupError GroupError
			)

			if errors.As(err, &retryError) {
				time.Sleep(retryError.Delay)
				ScheduleC2ParallelCommand(cmd)
			} else if errors.As(err, &groupError) {
				cmd.Wait.AddGroupError(groupError)
			} else {
				cmd.Wait.AddError(fmt.Errorf("unexpected value: %w", err))
			}
		}
	}()
}
