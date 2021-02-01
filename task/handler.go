package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"scp_delegator/config"
	"scp_delegator/constant"
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
	logDir   string

	inspector  *Inspector
	ctx        context.Context
	cancelFunc context.CancelFunc
	state      string
}

func CreateSingleTaskHandler(ctx *context.Context, cfg *config.Config, m *config.TaskMaterial) *SingleTaskHandler {
	h := &SingleTaskHandler{
		config:   cfg,
		material: m,
		state:    StateUninitialized,
	}
	h.ctx, h.cancelFunc = context.WithCancel(*ctx)
	h.logDir = logger.GetOutputDir()

	h.init()
	logger.Wrapper.LogTrace("Task handler created")
	return h
}

func (h *SingleTaskHandler) Start() error {
	return h.run()
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
	if h.material.ConditionMaterial.Condition!=nil {
		h.inspector = CreateInspector(&h.ctx, &h.material.ConditionMaterial, h.execCommand)
	}
	h.state = StateInitialized
}

func (h *SingleTaskHandler) run() error {
	if h.material.ConditionMaterial.Condition != nil {
		logger.Wrapper.LogTrace("Task %d running with condition wait", h.material.TaskID)
		// Start inspector to monitoring
		// Blocking here that wait for condition hit
		h.state = StateMonitoring
		err := h.inspector.Start()
		if err != nil {
			logger.Wrapper.LogInfo("Stop task %d because received error %s", h.material.TaskID, err.Error())
			h.state = StateDoneFail
			return err
		}
	} else {
		// Run action immediately
		h.execCommand()
	}
	logger.Wrapper.LogTrace("Task handler %d finished", h.material.TaskID)
	return nil
}

func (h *SingleTaskHandler) execCommand() {
	logger.Wrapper.LogTrace("Task handler %d start to execute command", h.material.TaskID)
	// Set timeoutS for context
	timeoutS := h.material.ActionMaterial.ActProperty.TimeoutS
	if timeoutS == 0 {
		timeoutS = DefaultValueTimeoutS
	}
	h.ctx, h.cancelFunc = context.WithTimeout(h.ctx, time.Second*time.Duration(timeoutS))
	defer h.cancelFunc()

	count := h.material.ActionMaterial.ActProperty.Repeat.Count
	if count == 0 {
		count = DefaultValueRepeatCount
	}

	h.state = StateExecuting
	for {
		if count == 0 {
			break
		}
		err := h.execCommandOnce()
		if err != nil {
			logger.Wrapper.LogError("Error occurs when execute command, error %s", err.Error())
			break
		}
		count -= 1
		time.Sleep(time.Duration(h.material.ActionMaterial.ActProperty.Repeat.IntervalS) * time.Second)
	}

	if count == 0 {
		h.state = StateDoneSuccess
	} else {
		h.state = StateDoneFail
	}
	logger.Wrapper.LogTrace("Task handler %d end to execute command", h.material.TaskID)
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
		exeCommand(&h.ctx, act)
	}

	exeCommand(&h.ctx, h.material.ActionMaterial.Action)
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
		exeCommand(&h.ctx, act)
	}
	return nil
}

func composeInnerArguments(action *config.Action) string{
	var sb strings.Builder
	if action.Arguments!=nil && len(action.Arguments) > 0 {
		for _, arg := range action.Arguments {
			if arg.Command != "" {
				sb.WriteString(arg.Command)
				sb.WriteString(" ")
			}
			if arg.Value != "" {
				sb.WriteString(arg.Value)
				sb.WriteString(" ")
			}
		}
	}

	return strings.Trim(sb.String(), " ")
}

func composeArgument(action *config.Action) string {
	return strings.Trim(constant.EmbedBinaryOptions[action.Executable], " ")
}

func composeBinaryPath() string {
	// Get current path
	path, err := filepath.Abs(constant.ExecutorName)
	if err != nil {
		logger.Wrapper.LogError("Get error when query executable path, %s", err.Error())
		return constant.ExecutorName
	}
	return strings.Trim(path, " ")
}

func exeCommand(ctx *context.Context, action *config.Action) {
	bin := composeBinaryPath()
	arg := composeArgument(action)
	innerArg := composeInnerArguments(action)

	cmd := exec.CommandContext(*ctx, bin, arg, innerArg)
	logger.Wrapper.LogTrace("Execute command %s\n", cmd.String())

	// Blocking here util process finished or timeout triggered by context
	// If output path is not empty then output to specific path.
	if action.Output != "" {
		result, err := cmd.CombinedOutput()
		outputFile(action.Output, result)
		if err != nil {
			logger.Wrapper.LogError("Execute command error %s", err.Error())
		}
	} else {
		err := cmd.Run()
		if err != nil {
			logger.Wrapper.LogError("Execute command error %s", err.Error())
		}
	}
}

func outputFile(fileName string, data []byte) {
	p := filepath.Join(config.GetOutputDir(), fileName)
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Wrapper.LogError("Get error during opening file %s, error %s", fileName, err.Error())
		return
	}
	defer f.Close()

	_,err = f.Write(data)
	if err!=nil {
		logger.Wrapper.LogError("Get error during opening file %s, error %s", fileName, err.Error())
	}
}
