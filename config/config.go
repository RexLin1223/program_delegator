package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scp_delegator/constant"
	"scp_delegator/logger"
	"scp_delegator/system/windows"
	"strings"
)

const (
	VariableRegKeyPrefix = "reg_key:"
	VariableOutputDir    = "{output_dir}"
)

type TemplateReader struct {
	ActionReader            TemplateAction
	ActionPropertiesReader  TemplateActionProperty
	ConditionsReader        TemplateCondition
	ConditionCriteriaReader TemplateConditionCriteria
}

var VariableMap = map[string]string{}

type TemplateAction map[uint32]Action
type TemplateActionProperty map[uint32]ActionProperty
type TemplateCondition map[uint32]Condition
type TemplateConditionCriteria map[uint32]ConditionCriteria

func removeBrackets(s string) string {
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	return s
}

func removeSubString(s string, sub string) string {
	return strings.ReplaceAll(s, sub, "")
}

func trimAll(s string, cutset string) string {
	for {
		temp := strings.Trim(s, cutset)
		if strings.EqualFold(temp, s) {
			break
		}
		s = temp
	}
	return s
}

func queryRegKey(value string) (string, error) {
	s := removeSubString(value, VariableRegKeyPrefix)
	s = removeBrackets(s)
	s = strings.Trim(s, " ")

	args := strings.Split(s, ",")
	if len(args) != 3 {
		return "", errors.New("input value format is incorrect, format should be `{HKEY_LOCAL_MACHINE}, {SOFTWARE\\TrendMicro\\Deep Security Agent}, {InstallationFolder}`")
	}
	s, err := windows.QueryRegKey64(trimAll(args[0], " "), trimAll(args[1], " "), trimAll(args[2], " "))
	if err != nil {
		return "", err
	}
	return s, nil
}

func traverseConfig(cfg *Config) *Config {
	// Parse variable structure to amp
	for _, v := range cfg.Variables {
		val, _ := getVariable(cfg, v.Alias)
		VariableMap[v.Alias] = val
	}

	variablesInterpreter := func(s string) string {
		matched := strings.ContainsAny(s, "{}")
		if !matched {
			return s
		}

		for key, val := range VariableMap {
			if strings.Contains(s, key) {
				s = removeBrackets(s)
				s = strings.ReplaceAll(s, key, val)
			}
		}

		return s
	}

	// Add any string type filed here for parsing by variable
	cfg.Version = variablesInterpreter(cfg.Version)
	cfg.LogLevel = variablesInterpreter(cfg.LogLevel)

	// Traverse all element of Upload
	cfg.Upload.AzBlob.HostName = variablesInterpreter(cfg.Upload.AzBlob.HostName)
	cfg.Upload.AzBlob.ContainerName = variablesInterpreter(cfg.Upload.AzBlob.ContainerName)
	cfg.Upload.AzBlob.AccountName = variablesInterpreter(cfg.Upload.AzBlob.AccountName)
	cfg.Upload.AzBlob.SASToken = variablesInterpreter(cfg.Upload.AzBlob.SASToken)
	cfg.Upload.Proxy.Host = variablesInterpreter(cfg.Upload.Proxy.Host)
	cfg.Upload.SEGCaseID = variablesInterpreter(cfg.Upload.SEGCaseID)
	cfg.Upload.CompanyID = variablesInterpreter(cfg.Upload.CompanyID)
	cfg.Upload.DeviceID = variablesInterpreter(cfg.Upload.DeviceID)

	// Traverse Tasks
	for i, task := range cfg.Tasks {
		cfg.Tasks[i].Name = variablesInterpreter(task.Name)
	}

	// Traverse Template
	for i, act := range cfg.Template.Actions {
		cfg.Template.Actions[i].Name = variablesInterpreter(act.Name)
		cfg.Template.Actions[i].Executable = variablesInterpreter(act.Executable)
		cfg.Template.Actions[i].Output = variablesInterpreter(act.Output)
		for j, arg := range act.Arguments {
			cfg.Template.Actions[i].Arguments[j].Command = variablesInterpreter(arg.Command)
			cfg.Template.Actions[i].Arguments[j].Value = variablesInterpreter(arg.Value)
		}
	}
	for i, cond := range cfg.Template.Conditions {
		cfg.Template.Conditions[i].Name = variablesInterpreter(cond.Name)
		cfg.Template.Conditions[i].TargetProcess = variablesInterpreter(cond.TargetProcess)
	}

	for i, cri := range cfg.Template.ConditionCriteria {
		cfg.Template.ConditionCriteria[i].Type = variablesInterpreter(cri.Type)
		cfg.Template.ConditionCriteria[i].Operator = variablesInterpreter(cri.Operator)
	}
	return cfg
}

func getVariable(cfg *Config, alias string) (string, error) {
	if len(alias) == 0 {
		return "", errors.New("given alias is empty")
	}

	for _, v := range cfg.Variables {
		if !strings.EqualFold(v.Alias, alias) {
			continue
		}

		if strings.Contains(v.Value, VariableRegKeyPrefix) {
			// Query value from reg key
			s, err := queryRegKey(v.Value)
			if err != nil {
				return "", err
			}
			return s, nil
		}

		if strings.EqualFold(v.Value, VariableOutputDir) {
			v.Value = GetOutputDir()
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
	templateCriteria := cfg.Template.ConditionCriteria
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
		mc := make([]*ConditionCriteria, len(cond.Criteria.Mandatory))

		for i, cid := range cond.Criteria.Mandatory {
			mc[i] = GetTemplateCriteria(cfg, cid)
		}
		cm.MandatoryCriteria = mc

		oc := make([]*ConditionCriteria, len(cond.Criteria.Optional))
		for i, cid := range cond.Criteria.Optional {
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

func GetOutputDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Wrapper.LogError("Get output dir failed with error=%s", err)
		return ""
	}

	outputDir := filepath.Join(currentDir, constant.OutputDirectory)
	err = os.MkdirAll(outputDir, 0744)
	if err != nil {
		log.Printf("Can't get create directory for output dir, error=%s", err)
		return currentDir
	}

	return filepath.Join(outputDir, "")
}
