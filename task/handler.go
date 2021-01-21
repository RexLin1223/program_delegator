package task

import (
	"context"
	"os"
	"os/exec"
	"scp_delegator/config"
	"scp_delegator/logger"
	"strings"
	"time"
)

const (
	StateUninitialized = "Uninitialized"
	StateInitialized   = "Initialized"
	StateRunning       = "Running"
	StateStopping      = "Stopping"
	StateCancelled     = "Cancelled"
	StateDoneSuccess   = "TaskDoneWithSuccessful"
	StateDoneFail      = "TaskDoneWithFailed"
)

type SingleTaskHandler struct {
	ctx     context.Context
	cmd     *exec.Cmd
	state   string
	logPath string

	templateReader *config.TemplateReader
	action         *config.Action
	condition      *config.Condition
}

func CreateSingleTaskHandler(ctx *context.Context, action *config.Action, condition *config.Condition, templateReader *config.TemplateReader) *SingleTaskHandler {
	h := &SingleTaskHandler{}
	h.ctx, _ = context.WithCancel(*ctx)
	h.templateReader = templateReader
	h.action = action
	h.condition = condition
	h.state = StateUninitialized
	return h
}

func (h *SingleTaskHandler) composeArgument() string {
	var sb strings.Builder
	for _, arg := range h.action.Arguments {
		sb.WriteString(arg.Command)
		sb.WriteString(" ")
		sb.WriteString(arg.Value)
		sb.WriteString(" ")
	}
	return sb.String()
}

func (h *SingleTaskHandler) composeBinary() string {
	path, err := os.Executable()
	if err != nil {
		logger.LogError("Get error when query executable path, %s", err.Error())
	}
	return path + "\\rp_main.exe"
}

func (h *SingleTaskHandler) init() error {
	bin := h.composeBinary()
	arg := h.composeArgument()
	h.cmd = exec.CommandContext(h.ctx, bin, arg)
	h.state = StateInitialized
	return nil
}

func (h *SingleTaskHandler) run() {
	err := h.cmd.Start()
	if err != nil {
		logger.LogError("Get error during command starting. %s", err.Error())
		h.state = StateDoneFail
		return
	}

	h.state = StateRunning
	// Wait for command finished
	for {
		select {
		case <-h.ctx.Done():
			h.state = StateCancelled
			break
		default:
			if h.CheckForProcessFinished() {
				break
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func (h *SingleTaskHandler) CheckForProcessFinished() bool {
	exited := h.cmd.ProcessState.Exited()
	if exited {
		if h.cmd.ProcessState.Success() {
			h.state = StateDoneSuccess
		} else {
			h.state = StateDoneFail
		}
		return true
	}
	return false
}

func (h *SingleTaskHandler) stop() {
	h.state = StateStopping
	h.cmd.Process.Kill()
	h.state = StateCancelled
}

func (h *SingleTaskHandler) getState() string {
	return h.state
}
