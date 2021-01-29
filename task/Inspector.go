package task

import (
	"context"
	"errors"
	"fmt"
	"math"
	"scp_delegator/config"
	"scp_delegator/logger"
	"scp_delegator/metric"
	"syscall"
	"time"
)

// =================================================
// Add method at below if add new condition checker
// =================================================
type ConditionChecker func(string, *config.ConditionCriteria) (bool, error)

var ConditionCheckerMap = map[string]ConditionChecker{
	"CPU":                ConditionCheckerCPU,
	"Memory":             ConditionCheckerMemory,
	"DiskAvailableUsage": ConditionCheckerDiskFreeSpace,
}

func ConditionCheckerCPU(processName string, c *config.ConditionCriteria) (bool, error) {
	currentUsage, err := metric.GetProcessCpuUsage(processName)
	if err != nil {
		return false, err
	}

	return compareWithOperator(uint64(math.Ceil(currentUsage)), uint64(c.Threshold), c.Operator), nil
}

func ConditionCheckerMemory(processName string, c *config.ConditionCriteria) (bool, error) {
	currentUsage, err := metric.GetProcessMemoryUsageMB(processName)
	if err != nil {
		return false, err
	}
	return compareWithOperator(uint64(currentUsage), uint64(c.Threshold), c.Operator), nil
}

func ConditionCheckerDiskFreeSpace(processName string, c *config.ConditionCriteria) (bool, error) {
	path, err := syscall.Getwd()
	if err != nil {
		return false, err
	}
	currentUsage, err := metric.GetDiskFreeGB(path)
	if err != nil {
		return false, err
	}
	return compareWithOperator(uint64(currentUsage), uint64(c.Threshold), c.Operator), nil
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
	} else if operator == "" {
		logger.Wrapper.LogInfo("No operator given, invalid comparison")
		return false
	} else {
		logger.Wrapper.LogError("Unrecognized operator %s", operator)
		return false
	}
}

type OnRunAction func()
type CriteriaID uint32
type Inspector struct {
	ctx           context.Context
	cancelFunc    context.CancelFunc
	material      *config.ConditionMaterial
	runAction		OnRunAction

	lastCheckTime map[CriteriaID]uint64
}

func CreateInspector(ctx *context.Context, m *config.ConditionMaterial, runAction OnRunAction) *Inspector {
	i := Inspector{
		material:      m,
		runAction:  runAction,
	}
	i.ctx, i.cancelFunc = context.WithTimeout(*ctx, time.Duration(m.Condition.TimeoutS) * time.Second )

	i.init()
	logger.Wrapper.LogTrace("Inspector created.")
	return &i
}

func (i *Inspector) init() {
}

func (i *Inspector) Start() error {
	logger.Wrapper.LogTrace("Inspector start begin")
	defer i.cancelFunc()
loop:
	for {
		select
		{
		case <-i.ctx.Done():
			{
				logger.Wrapper.LogInfo("Notified to close inspector")
				return errors.New("inspector cancelled by parent")
			}
		default:
			{
				conditionHit, err := i.validateWithCriteria()
				if err != nil {
					return err
				} else if conditionHit == true {
					// Notify task handler to perform action
					logger.Wrapper.LogTrace("Condition are satisfied to set.")
					i.runAction()
					break loop
				}
				time.Sleep(time.Millisecond * 1000)
			}
		}
	}
	logger.Wrapper.LogTrace("Inspector start end")
	return nil
}

func (i *Inspector) Stop() {
	logger.Wrapper.LogTrace("Inspector stopped by manual trigger")
	i.cancelFunc()
}

func (i *Inspector) validateWithCriteria() (bool, error) {
	compareWithSingleCriterion := func(c *config.ConditionCriteria) (bool, error) {
		checker := ConditionCheckerMap[c.Type]
		if checker == nil {
			return false, errors.New(fmt.Sprintf("invalid criteria, type=%s, ID=%d", c.Type, c.ID))
		}

		isSatisfied, err := checker(i.material.Condition.TargetProcess, c)
		if err != nil {
			return false, err
		} else if !isSatisfied {
			return false, nil
		}

		if c.MaturityMS > 0 {
			// Wait for maturity
			time.Sleep(time.Duration(c.MaturityMS) * time.Millisecond)
			// Validate again
			isSatisfied, err = checker(i.material.Condition.TargetProcess, c)
			if err != nil {
				return false, err
			} else if !isSatisfied {
				return false, nil
			}
		}
		return true, nil
	}

	// Check for each mandatory criteria
	for _, mc := range i.material.MandatoryCriteria {
		result, err:=compareWithSingleCriterion(mc)
		if err !=nil{
			return false, err
		}
		if !result {
			// Early return because one of mandatory not satisfied.
			return false, nil
		}
	}

	// Check whether hit any optional criteria
	for _, oc := range i.material.OptionalCriteria {
		result, err:=compareWithSingleCriterion(oc)
		if err !=nil {
			return false, err
		}
		if result {
			// Satisfied to all needed conditions
			return true, nil
		}

	}

	return false, nil
}
