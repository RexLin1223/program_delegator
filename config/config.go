package config

import (
	"errors"
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

func ParseTemplates(cfg *Config) TemplateReader {
	var templates TemplateReader
	t := cfg.Template
	for _, a := range t.Actions {
		templates.ActionReader[a.ID] = a
	}

	for _, ap := range t.ActionsProperties {
		templates.ActionPropertiesReader[ap.ID] = ap
	}

	for _, c := range t.Conditions {
		templates.ConditionsReader[c.ID] = c
	}

	for _, cc := range t.ConditionCriterias {
		templates.ConditionCriteriaReader[cc.ID] = cc
	}
	return templates
}
