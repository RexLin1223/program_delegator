package task

import (
	"context"
	"errors"
	"fmt"
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
	StateMonitoring    = "Monitoring"
	StateExecuting     = "Executing"
	StateStopping      = "Stopping"
	StateCancelled     = "Cancelled"
	StateDoneSuccess   = "FinishWithSuccessful"
	StateDoneFail      = "FinishWithFailed"
)

const (
	DefaultValueTimeoutS    = 600
	DefaultValueRepeatCount = 1
)

type SingleTaskHandler struct {
	config   *config.Config
	material *config.TaskMaterial

	inspector  *Inspector
	ctx        context.Context
	cancelFunc context.CancelFunc
	state      string
}

func CreateSingleTaskHandler(ctx *context.Context, cfg *config.Config, m *config.TaskMaterial) *SingleTaskHandler {
	h := &SingleTaskHandler{
		config:   cfg,
		material: m,
	}
	h.ctx, h.cancelFunc = context.WithCancel(*ctx)
	h.state = StateUninitialized

	h.init()
	return h
}

func (h *SingleTaskHandler) Start() {
	h.run()
}

func (h *SingleTaskHandler) Stop() {
	h.state = StateStopping
	h.cancelFunc()
	h.state = StateCancelled
}

func (h *SingleTaskHandler) GetState() string {
	return h.state
}

func (h *SingleTaskHandler) init() {
	h.inspector = CreateInspector(&h.ctx, &h.material.ConditionMaterial, h.execCommand)
	h.state = StateInitialized
}

func (h *SingleTaskHandler) run() {
	if h.material.ConditionMaterial.Condition != nil {
		// Start inspector to monitoring
		// Blocking here that wait for condition hit
		h.state = StateMonitoring
		err := h.inspector.Start()
		if err != nil {
			logger.LogInfo("Stop task %d because received error %s", h.material.TaskID, err.Error())
			return
		}
	}
	// Run action
	h.execCommand()
}

func (h *SingleTaskHandler) execCommand() {
	// Set timeout for context
	timeout := h.material.ActionMaterial.ActProperty.TimeoutS
	if timeout == 0 {
		timeout = DefaultValueTimeoutS
	}
	h.ctx, h.cancelFunc = context.WithTimeout(h.ctx, time.Duration(timeout))
	defer h.cancelFunc()

	count := h.material.ActionMaterial.ActProperty.Repeat.Count
	if count == 0 {
		count = DefaultValueRepeatCount
	}

	h.state = StateExecuting
	for {
		if count == 0 {
			logger.LogTrace("Task finished")
			break
		}
		err := h.execCommandOnce()
		if err != nil {
			logger.LogError("Error occurs when execute command, error %s", err.Error())
			break
		}
		count -= 1
		time.Sleep(time.Duration(h.material.ActionMaterial.ActProperty.Repeat.IntervalS) * time.Second)
	}

	if count ==0 {
		h.state = StateDoneSuccess
	}else{
		h.state = StateDoneFail
	}
}

func (h *SingleTaskHandler) execCommandOnce() error {
	if h.material.ActionMaterial.Action.PreAction != 0 {
		act := config.GetTemplateAction(h.config, h.material.ActionMaterial.Action.PreAction)
		if act == nil {
			return errors.New(fmt.Sprintf("can't get pre-action from template ID %d", h.material.ActionMaterial.Action.PreAction))
		}

		actProperty := config.GetTemplateActionProperty(h.config, act.Property)
		if actProperty == nil {
			return errors.New(fmt.Sprintf("can't get pre-action property from template ID %d", act.Property))
		}
		// Get pre-action & action property
		exeCommand(&h.ctx, act, actProperty)
	}

	exeCommand(&h.ctx, h.material.ActionMaterial.Action, h.material.ActionMaterial.ActProperty)
	if h.material.ActionMaterial.ActProperty.PeriodS != 0 {
		time.Sleep(time.Second * time.Duration(h.material.ActionMaterial.ActProperty.PeriodS))
	}

	if h.material.ActionMaterial.Action.PostAction != 0 {
		act := config.GetTemplateAction(h.config, h.material.ActionMaterial.Action.PostAction)
		if act == nil {
			return errors.New(fmt.Sprintf("can't get post-action from template ID %d", h.material.ActionMaterial.Action.PostAction))
		}

		actProperty := config.GetTemplateActionProperty(h.config, act.Property)
		if actProperty == nil {
			return errors.New(fmt.Sprintf("can't get post-action property from template ID %d", act.Property))
		}
		// Get pre-action & action property
		exeCommand(&h.ctx, act, actProperty)
	}
	return nil
}

func composeArgument(action *config.Action) string {
	var sb strings.Builder

	// Stuff prefix options command of embed tool
	sb.WriteString(EmbedBinaryOptions[action.Executable])

	if len(action.Arguments) > 0 {
		sb.WriteString(` --args `)
		sb.WriteString(`"`) // Prefix sign and quote
		for _, arg := range action.Arguments {
			sb.WriteString(arg.Command)
			sb.WriteString(" ")
			sb.WriteString(arg.Value)
			sb.WriteString(" ")
		}
		sb.WriteString(`"`) // Post quote
	}
	return sb.String()
}

func composeBinaryPath() string {
	path, err := os.Getwd()
	if err != nil {
		logger.LogError("Get error when query executable path, %s", err.Error())
		return "rp_main.exe"
	}
	return path + "\\rp_main.exe"
}

func exeCommand(parentCtx *context.Context, action *config.Action, actProperty *config.ActionProperty) {
	bin := composeBinaryPath()
	arg := composeArgument(action)
	ctx, _ := context.WithTimeout(*parentCtx, time.Millisecond*time.Duration(actProperty.TimeoutS))

	cmd := exec.CommandContext(ctx, bin, arg)
	// Blocking here util command finished or timeout triggered by context
	err := cmd.Run()
	if err != nil {
		logger.LogError("Execute command error %s", err.Error())
	}
	// Execute finished
	if action.Output != "" {
		result, _ := cmd.Output()
		outputFile(action.Output, result)
	}
}

func outputFile(fileName string, output []byte) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.LogError("Get error during opening file %s, error %s", fileName, err.Error())
		return
	}
	defer f.Close()
	if _, err := f.Write(output); err != nil {
		logger.LogError("Get error during writing output to file %s, error %s", fileName, err.Error())
		return
	}
}
