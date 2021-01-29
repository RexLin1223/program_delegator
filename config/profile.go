package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"scp_delegator/logger"
)

// Config is root structure of profile file
type Config struct {
	Version   string     `json:"scp_version"`
	Period    uint32     `json:"Period"`
	LogLevel  string     `json:"log_level"`
	AsyncRun  bool       `json:"async_execute"`
	Tasks     []Task     `json:"tasks"`
	Template  Template   `json:"template"`
	Variables []Variable `json:"variables"`
	Upload    Upload     `json:"upload"`
}

// Task struct formed by customized task which user defined.
type Task struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	ConditionID uint32 `json:"condition"`
	ActionID    uint32 `json:"action"`
}

// Template includes action and condition materials which used to compose the user defined task.
type Template struct {
	Actions           []Action            `json:"actions"`
	ActionsProperties []ActionProperty    `json:"action_properties"`
	Conditions        []Condition         `json:"conditions"`
	ConditionCriteria []ConditionCriteria `json:"condition_criteria"`
}

// Action material is the operation adopt by user defined task.
type Action struct {
	ID         uint32     `json:"id"`
	PreAction  uint32     `json:"pre_action"`
	Name       string     `json:"name"`
	Executable string     `json:"executable"`
	Arguments  []Argument `json:"arguments"`
	PostAction uint32     `json:"post_action"`
	Property   uint32     `json:"property"`
	Output     string     `json:"output"`
}

// Argument is command of executable program
type Argument struct {
	Command string `json:"command"`
	Value   string `json:"value"`
}

// ActionProperty is properties structure which bind to an action.
type ActionProperty struct {
	ID       uint32         `json:"id"`
	TimeoutS uint32         `json:"timeout_sec"`
	PeriodS  uint32         `json:"period_sec"`
	Repeat   RepeatProperty `json:"repeat"`
}

// RepeatProperty struct stores repeat count and interval of each operation.
type RepeatProperty struct {
	Count     uint32 `json:"count"`
	IntervalS uint32 `json:"interval_sec"`
}

// Condition struct is condition template before perform actions.
type Condition struct {
	ID            uint32   `json:"id"`
	Name          string   `json:"name"`
	TargetProcess string   `json:"monitor_process"`
	TimeoutS      uint32   `json:"timeout_sec"`
	Criteria      Criteria `json:"criteria"`
}

// Criteria is rule set of ConditionsCriteria which need fulfill mandatory criteria and least one of optional criteria.
type Criteria struct {
	Mandatory []uint32 `json:"mandatory"`
	Optional  []uint32 `json:"optional"`
}

// ConditionCriteria is boundary of condition trigger point.
type ConditionCriteria struct {
	ID         uint32 `json:"id"`
	Type       string `json:"type"`
	Interval   uint32 `json:"interval_sec"`
	Threshold  uint32 `json:"threshold"`
	Operator   string `json:"operator"`
	MaturityMS uint32 `json:"maturity_ms"`
}

// Variable composed by alias and value store in an element of map.
type Variable struct {
	Alias string `json:"alias"`
	Value string `json:"value"`
}

// Upload structure stores upload setting
type Upload struct {
	AzBlob         AzBlob `json:"azure_blob"`
	Proxy          Proxy  `json:"proxy"`
	MaxBlockSizeMB uint32 `json:"max_block_size_MB"`
	RateLimitMB    uint32 `json:"rate_limit_MB"`
	TimeoutS       uint32 `json:"timeout_sec"`
	MaxRetryCount  uint32 `json:"max_retry_count"`
	SEGCaseID      string `json:"seg_case_id"`
	CompanyID      string `json:"company_id"`
	DeviceID       string `json:"device_id"`
}

// AzBlob struct stores upload properties of Azure Blob needed
type AzBlob struct {
	HostName      string `json:"host_name"`
	AccountName   string `json:"account_name"`
	ContainerName string `json:"container_name"`
	SASToken      string `json:"sas_token"`
}

// Proxy struct composed by host and port.
type Proxy struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func readProfile() ([]byte, error) {
	f, err := os.Open("profile.json")
	defer f.Close()
	if err != nil {
		logger.Wrapper.LogError("Can't open profile with error %s", err)
		return nil, err
	}

	s, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Wrapper.LogError("Can't read profile with error %s", err)
		return nil, err
	}
	return s, nil
}

// ParseProfile will parse profile at current directory.
func ParseProfile() *Config {
	cfg := &Config{}
	profile, err := readProfile()
	if err != nil {
		return nil
	}
	json.Unmarshal(profile, cfg)

	// Parse variable
	cfg = traverseConfig(cfg)
	return cfg
}
