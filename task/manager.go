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
	tp        config.TemplateReader
	handlers  map[uint32]*SingleTaskHandler
}

// CreateTaskHandler is factory pattern to create task handler
func CreateTaskHandler(cfg *config.Config) (*TasksManager, error) {
	tm := &TasksManager{}
	err := tm.init(cfg)
	if err != nil {
		return nil, err
	}
	return tm, nil
}

func (tm *TasksManager) init(cfg *config.Config) error {
	tm.cfg = cfg
	tm.ctx, tm.cancelFun = context.WithTimeout(context.Background(), time.Duration(tm.cfg.Period)*time.Second)
	tm.tp = config.ParseTemplates(cfg)
	for _, task := range tm.cfg.Tasks {
		act := config.GetTemplateAction(cfg, task.ActionID)
		cond := config.GetTemplateCondition(cfg, task.ConditionID)
		if act ==nil || cond ==nil{
			logger.LogError("Can't create task %d, because can't find relevant actionID %d or conditionID %d.", task.ID, task.ActionID, task.ConditionID)
		}
		handler :=CreateSingleTaskHandler(&tm.ctx, act, cond, &tm.tp)
		err := handler.init()
		if err !=nil{
			logger.LogError("Get error when initial task%d, error=%s", task.ID, err.Error())
		}
		tm.handlers[task.ID] = handler
	}
	return nil
}

func (tm *TasksManager) Run() {
	isAsync := tm.cfg.AsyncRun
	for _, handler := range tm.handlers {
		if isAsync {
			go handler.run()
		} else{
			handler.run()
		}
	}
}

func (tm *TasksManager) RunOne(taskID uint32) {
	handler:= tm.handlers[taskID]
	if handler != nil{
		handler.run()
	}
}

func (tm *TasksManager) Stop() {
	// Call cancel function to trigger context done.
	tm.cancelFun()
}

func (tm *TasksManager) StopOne(taskID uint32) {
	h := tm.handlers[taskID]
	if h!=nil{
		h.stop()
	}
}


