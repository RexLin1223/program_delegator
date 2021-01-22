package task

import (
	"context"
	"errors"
	"math"
	"scp_delegator/config"
	"scp_delegator/logger"
	"scp_delegator/metric"
	"syscall"
	"time"
)

const (
	StateInspectorPending   = "Pending"
	StateInspectorSatisfied = "Satisfied"
	StateInspectorFinished  = "Finished"
)

// =================================================
// Add method at below if add new condition checker
// =================================================
type ConditionChecker func(string, *config.ConditionCriteria) bool

var ConditionCheckerMap = map[string]ConditionChecker{
	"CPU":           ConditionCheckerCPU,
	"Memory":        ConditionCheckerMemory,
	"DiskFreeSpace": ConditionCheckerDiskFreeSpace,
}

func ConditionCheckerCPU(processName string, c *config.ConditionCriteria) bool {
	currentUsage := metric.GetProcessCpuUsage(processName)
	if currentUsage == -1 {
		return false
	}

	return compareWithOperator(uint64(math.Ceil(currentUsage)), uint64(c.Threshold), c.Operator)
}

func ConditionCheckerMemory(processName string, c *config.ConditionCriteria) bool {
	currentUsage := metric.GetProcessMemoryUsageMB(processName)
	if currentUsage == -1 {
		return false
	}
	return compareWithOperator(uint64(currentUsage), uint64(c.Threshold), c.Operator)
}

func ConditionCheckerDiskFreeSpace(processName string, c *config.ConditionCriteria) bool {
	path, err := syscall.Getwd()
	if err != nil {
		logger.LogError("Get current path error %s", err.Error())
		return false
	}
	currentUsage := metric.GetDiskFreeGB(path)
	if currentUsage == -1 {
		return false
	}
	return compareWithOperator(uint64(currentUsage), uint64(c.Threshold), c.Operator)
}

// =================================================

func compareWithOperator(value uint64, threshold uint64, operator string) bool {
	if operator == ">" {
		return value > threshold
	} else if operator == ">=" {
		return value >= threshold
	} else if operator == "<" {
		return value < threshold
	} else if operator == "<=" {
		return value <= threshold
	} else if operator == "!=" {
		return value != threshold
	} else if operator == "==" || operator == "<>" {
		return value == threshold
	} else {
		logger.LogError("Unrecognized operator %s", operator)
		return false
	}
}

type ConditionSatisfiedFunc func()

type Inspector struct {
	ctx           context.Context
	cancelFunc    context.CancelFunc
	material      *config.ConditionMaterial
	satisfiedFunc ConditionSatisfiedFunc
}

func CreateInspector(ctx *context.Context, m *config.ConditionMaterial, conditionHit ConditionSatisfiedFunc) *Inspector {
	i := Inspector{
		material: m,
		satisfiedFunc: conditionHit,
	}
	i.ctx, i.cancelFunc = context.WithCancel(*ctx)

	i.init()
	return &i
}

func (i *Inspector) init() {

}

func (i *Inspector) Start() error {
	logger.LogTrace("Inspector start begin")
	for {
		select
		{
		case <-i.ctx.Done():
			{
				logger.LogInfo("Notified to close inspector")
				return errors.New("inspector cancelled by parent")
			}
		default:
			{
				if i.checkCriterias() {
					// Notify task handler to perform action
					i.satisfiedFunc()
					break
				}
				time.Sleep(time.Millisecond * 1000)
			}
		}
	}
	logger.LogTrace("Inspector start end")
	return nil
}

func (i *Inspector) Stop() {
	logger.LogTrace("Inspector stop triggered")
	i.cancelFunc()
}

func (i *Inspector) checkCriterias() bool {
	// Check for each mandatory criteria
	for _, mc := range i.material.MandatoryCriteria {
		checker := ConditionCheckerMap[mc.Type]
		if checker == nil {
			return false
		}
		if checker(i.material.Condition.Name, mc) == false {
			// Lack of mandatory criteria, so that rerun earlier.
			return false
		}
	}

	// Check whether hit any optional criteria
	for _, oc := range i.material.OptionalCriteria {
		checker := ConditionCheckerMap[oc.Type]
		if checker == nil {
			return false
		}
		if checker(i.material.Condition.Name, oc) == true {
			// Hit the optional
			return true
		}
	}

	return false
}
