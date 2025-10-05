package host

import (
	"context"
	"errors"
	"log"

	"github.com/vaastav/columbo_go/components"
	"github.com/vaastav/columbo_go/events"
	"github.com/vaastav/columbo_go/trace"
)

type SyscallStatus int

const (
	INSYSCALL SyscallStatus = iota
	INSYSRET
)

type Syscall struct {
	*components.BasePlugin
	InStream     components.Plugin
	trace_states map[string]*trace.ColumboTrace
	status       map[string]SyscallStatus
	cntr         uint64
}

func NewSyscall(ctx context.Context, ins components.Plugin, buffer_size int, ID int) (*Syscall, error) {
	outs, err := components.NewDataStream(ctx, make(chan *trace.ColumboTrace, buffer_size))
	if err != nil {
		return nil, err
	}

	syscall := &Syscall{
		components.NewBasePlugin(ID, outs),
		ins,
		make(map[string]*trace.ColumboTrace),
		make(map[string]SyscallStatus),
		0,
	}
	return syscall, nil
}

func (s *Syscall) mergeStates(state, t2 *trace.ColumboTrace) {
	state.Attributes["span_type"] = "syscall"
	state.Spans[0].Events = append(state.Spans[0].Events, t2.Spans[0].Events...)
}

func (s *Syscall) processTrace(t *trace.ColumboTrace) {
	s.cntr++
	if t.Type == trace.SPAN || t.Type == trace.TRACE {
		s.OutStream.Push(t)
		return
	}
	if t.Type == trace.EVENT {
		event_type := t.Attributes["event_type"]
		if event_type == events.KHostInstrT.String() {
			exec_id := t.Attributes["exec_id"]
			if v, ok := s.status[exec_id]; ok {
				if v == INSYSCALL {
					state := s.trace_states[exec_id]
					s.mergeStates(state, t)
					s.trace_states[exec_id] = state
					instr := t.Attributes["instr_name"]
					if instr == "sysret" {
						s.status[exec_id] = INSYSRET
					}
				} else if v == INSYSRET {
					state := s.trace_states[exec_id]
					s.mergeStates(state, t)
					s.trace_states[exec_id] = state
					is_last_microop := t.Attributes["is_last_microop"]
					if is_last_microop == "true" {
						s.OutStream.Push(state)
						s.trace_states[exec_id] = nil
						delete(s.status, exec_id)
						delete(s.trace_states, exec_id)
					}
				}
			} else {
				// Check if its a syscall, otherwise discard
				instr := t.Attributes["instr_name"]
				if instr == "syscall" {
					s.status[exec_id] = INSYSCALL
					s.trace_states[exec_id] = t
				}
			}
		} else {
			// We can push anything that's not an instruction
			s.OutStream.Push(t)
		}
	}
}

func (s *Syscall) Run(ctx context.Context) error {
	instream := s.InStream.GetOutDataStream()
	if instream == nil {
		return errors.New("Incoming plugin has a nil stream")
	}
	for {
		select {
		case t, ok := <-instream.Data:
			if !ok {
				s.OutStream.Close()
				return nil
			}
			s.processTrace(t)
		case <-ctx.Done():
			log.Println("Context is done. Quitting.")
			s.OutStream.Close()
			return nil
		}
	}
}

func (s *Syscall) IncomingPlugins() []components.Plugin {
	return []components.Plugin{s.InStream}
}
