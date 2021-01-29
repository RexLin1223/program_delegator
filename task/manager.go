package task

import (
	"context"
	"scp_delegator/config"
	"scp_delegator/logger"
	"time"
)

type TasksManager struct {
	cfg       *config.Config
	ctx       context.Context
	cancelFun context.CancelFunc
	handlers  map[uint32]*SingleTaskHandler
}

// CreateTaskHandler is factory pattern to create task handler
func CreateTaskHandler(ctx *context.Context, cfg *config.Config) (*TasksManager, error) {
	tm := &TasksManager{
		cfg:       cfg,
		ctx:       nil,
		cancelFun: nil,
		handlers:  make(map[uint32]*SingleTaskHandler, len(cfg.Tasks)),
	}
	tm.ctx, tm.cancelFun = context.WithTimeout(*ctx, time.Duration(tm.cfg.Period)*time.Second)
	err := tm.init()
	if err != nil {
		return nil, err
	}
	logger.Wrapper.LogTrace("Task Manager created.")
	return tm, nil
}

func (tm *TasksManager) init() error {
	for _, task := range tm.cfg.Tasks {
		m, err := config.GetTaskMaterial(tm.cfg, task.ID)
		if err != nil {
			logger.Wrapper.LogError("Error: %s", err.Error())
		}

		handler := CreateSingleTaskHandler(&tm.ctx, tm.cfg, m)
		tm.handlers[task.ID] = handler
	}
	return nil
}

func (tm *TasksManager) Run() {
	isAsync := tm.cfg.AsyncRun
	for _, handler := range tm.handlers {
		if isAsync {
			go handler.Start()
		} else {
			handler.Start()
		}
	}
	if isAsync {
		// Block here
		tm.WaitAllTasks()
	}
}

func (tm *TasksManager) RunOne(taskID uint32) {
	handler := tm.handlers[taskID]
	if handler != nil {
		handler.Start()
	}
}

func (tm *TasksManager) Stop() {
	// Call cancel function to trigger context done.
	tm.cancelFun()
	logger.Wrapper.LogTrace("Task manager stopped")
}

func (tm *TasksManager) StopOne(taskID uint32) {
	h := tm.handlers[taskID]
	if h != nil {
		logger.Wrapper.LogInfo("Stop task %d", taskID)
		h.Stop()
	}
}

func (tm *TasksManager) WaitAllTasks() {
	var allTaskDone bool = false
	for !allTaskDone {
		// Assume all task are done.
		allTaskDone = true
		for _, task := range tm.handlers {
			state := task.GetState()
			// Check for all command ready
			if state != StateDoneSuccess && state != StateDoneFail {
				allTaskDone = false
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	logger.Wrapper.LogTrace("All task finished")
}
