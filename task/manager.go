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
func CreateTaskHandler(cfg *config.Config) (*TasksManager, error) {
	tm := &TasksManager{
		cfg:       cfg,
		ctx:       nil,
		cancelFun: nil,
		handlers:  make(map[uint32]*SingleTaskHandler, len(cfg.Tasks)),
	}
	tm.ctx, tm.cancelFun = context.WithTimeout(context.Background(), time.Duration(tm.cfg.Period)*time.Second)
	err := tm.init()
	if err != nil {
		return nil, err
	}
	return tm, nil
}

func (tm *TasksManager) init() error {
	for _, task := range tm.cfg.Tasks {
		m, err := config.GetTaskMaterial(tm.cfg, task.ID)
		if err != nil {
			logger.LogError("Error: %s", err.Error())
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
	if isAsync{
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
}

func (tm *TasksManager) StopOne(taskID uint32) {
	h := tm.handlers[taskID]
	if h != nil {
		logger.LogInfo("Stop task %d", taskID)
		h.Stop()
	}
}

func (tm *TasksManager) WaitAllTasks() {
	var allTaskDone bool
	for allTaskDone {
		// Assume all task are done.
		allTaskDone =true
		for _, task := range tm.handlers {
			state := task.GetState()
			// Check for all command ready
			if state != StateDoneSuccess || state != StateDoneFail {
				allTaskDone = false
				break
			}
		}
		if allTaskDone {
			break
		}
		time.Sleep(1000 * time.Second)
	}
}
