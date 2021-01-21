package task

import "scp_delegator/config"

type Monitor struct {
	cond *config.Condition
	condCriteria *config.ConditionCriteria
}

func (m *Monitor) Wait(){

}

