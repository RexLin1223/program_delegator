package config

import (
	"errors"
	"fmt"
	"scp_delegator/os/windows"
	"strings"
)

type TemplateReader struct {
	ActionReader            TemplateAction
	ActionPropertiesReader  TemplateActionProperty
	ConditionsReader        TemplateCondition
	ConditionCriteriaReader TemplateConditionCriteria
}

type TemplateAction map[uint32]Action
type TemplateActionProperty map[uint32]ActionProperty
type TemplateCondition map[uint32]Condition
type TemplateConditionCriteria map[uint32]ConditionCriteria

func removeBrackets(s string) string {
	return strings.Trim(s, "{}()[]")
}

func queryRegKey(value string) (string, error) {
	s := removeBrackets(value)
	args := strings.Split(s, ",")
	if len(args) != 3 {
		return "", errors.New("input value format is incorrect, format should be `{HKEY_LOCAL_MACHINE}, {SOFTWARE\\TrendMicro\\Deep Security Agent}, {InstallationFolder}`")
	}
	s, err := windows.QueryRegKey64(args[0], args[1], args[2])
	if err != nil {
		return "", err
	}
	return s, nil
}

func GetVariable(cfg *Config, alias string) (string, error) {
	if len(alias) == 0 {
		return "", errors.New("given alias is empty")
	}

	for _, v := range cfg.Variables {
		if !strings.EqualFold(v.Alias, alias) {
			continue
		}

		regPrefix := strings.Index(v.Value, "reg:")
		if regPrefix != -1 {
			// Query value from reg key
			s, err := queryRegKey(v.Value)
			if err != nil {
				return "", err
			}
			return s, nil
		}

		// Return mapping value directly
		return v.Value, nil
	}
	return "", errors.New("can't find given alias from variables list")
}

func GetTemplateAction(cfg *Config, actionID uint32) *Action {
	templateActions := cfg.Template.Actions
	for _, action := range templateActions {
		if action.ID == actionID {
			return &action
		}
	}
	return nil
}

func GetTemplateCondition(cfg *Config, actionID uint32) *Condition {
	templateCondition := cfg.Template.Conditions
	for _, condition := range templateCondition {
		if condition.ID == actionID {
			return &condition
		}
	}
	return nil
}

func GetTemplateActionProperty(cfg *Config, propertyID uint32) *ActionProperty {
	templateProperty := cfg.Template.ActionsProperties
	for _, property := range templateProperty {
		if property.ID == propertyID {
			return &property
		}
	}
	return nil
}

func GetTemplateCriteria(cfg *Config, criteriaID uint32) *ConditionCriteria {
	templateCriteria := cfg.Template.ConditionCriterias
	for _, criteria := range templateCriteria {
		if criteria.ID == criteriaID {
			return &criteria
		}
	}
	return nil
}

type TaskMaterial struct {
	TaskID            uint32
	ActionMaterial    ActionMaterial
	ConditionMaterial ConditionMaterial
}

type ActionMaterial struct {
	Action      *Action
	ActProperty *ActionProperty
}

type ConditionMaterial struct {
	Condition         *Condition
	MandatoryCriteria []*ConditionCriteria
	OptionalCriteria  []*ConditionCriteria
}

func GetTaskMaterial(cfg *Config, taskID uint32) (*TaskMaterial, error) {
	// Find task from task list
	var t *Task = nil
	for _, task := range cfg.Tasks {
		if task.ID == taskID {
			t = &task
			break
		}
	}
	if t == nil {
		return nil, errors.New(fmt.Sprintf("can't get task ID %d from task list", taskID))
	}

	// Find action from action template
	am := ActionMaterial{}
	act := GetTemplateAction(cfg, t.ActionID)
	if act == nil {
		return nil, errors.New(fmt.Sprintf("can't get action ID %d from action template", t.ActionID))
	}
	am.Action = act

	// Find action property from action property template
	actProperty := GetTemplateActionProperty(cfg, act.Property)
	if actProperty == nil {
		return nil, errors.New(fmt.Sprintf("can't get action prperty ID %d from action property template", act.Property))
	}
	am.ActProperty = actProperty

	// Find action from action template
	cm := ConditionMaterial{}
	cond := GetTemplateCondition(cfg, t.ConditionID)
	cm.Condition = cond

	if cond != nil {
		// Condition can be nil, it represents execute command immediately
		mc := make([]*ConditionCriteria, len(cond.Criterias.Mandatory))

		for i, cid := range cond.Criterias.Mandatory {
			mc[i] = GetTemplateCriteria(cfg, cid)
		}
		cm.MandatoryCriteria = mc

		oc := make([]*ConditionCriteria, len(cond.Criterias.Optional))
		for i, cid := range cond.Criterias.Optional {
			oc[i] = GetTemplateCriteria(cfg, cid)
		}
		cm.OptionalCriteria = oc
	}

	return &TaskMaterial{
		TaskID:            taskID,
		ActionMaterial:    am,
		ConditionMaterial: cm,
	}, nil
}
