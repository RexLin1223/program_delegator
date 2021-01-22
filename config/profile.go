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
	Period    uint32     `json:"execution_period"`
	FreeSpace uint32     `json:"free_disk_space_GB"`
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

// Template includes action and condtion materials which used to compose the user defined task.
type Template struct {
	Actions            []Action            `json:"actions"`
	ActionsProperties  []ActionProperty    `json:"action_properties"`
	Conditions         []Condition         `json:"conditions"`
	ConditionCriterias []ConditionCriteria `json:"condition_criterias"`
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
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	Target    string `json:"monitor_process"`
	Criterias Criterias `json:"criterias"`
}

// Criterias is rule set of ConditionsCriterias which need fulfill all criteria.
type Criterias struct {
	Mandatory []uint32 `json:"mandatory"`
	Optional  []uint32 `json:"optional"`
}

// ConditionCriteria is boundary of condition trigger ponint.
type ConditionCriteria struct {
	ID         uint32 `json:"id"`
	Type       string `json:"type"`
	Interval   uint32 `json:"interval"`
	Threshold  uint32 `json:"threshold"`
	Operator   string `json:"operator"`
	MaturityMS uint32 `json:"maturity"`
}

// Variable composed by alias and value store in an element of map.
type Variable struct {
	Alias string `json:"alias"`
	Value string `json:"value"`
}

// Upload structure stores upload setting
type Upload struct {
	AzureBlobHost    string         `json:"azure_blob_host"`
	ProxySetting     Proxy          `json:"proxy"`
	Authentication   Authentication `json:"authentication"`
	MaxChunkSizeMB   uint32         `json:"max_chunk_size_MB"`
	BandwidthLimitMB uint32         `json:"bandwidth_limit_MB"`
	Grouping         []Group        `json:"grouping"`
}

// Proxy struct composed by host and port.
type Proxy struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// Authentication struct composed by SAS token and access key.
type Authentication struct {
	SAS       string `json:"sas"`
	AccessKey string `json:"access_key"`
}

// Group struct specific which task will be compressed together.
type Group struct {
	Tasks []uint32 `json:"tasks"`
}

func readProfile() ([]byte, error) {
	f, err := os.Open("profile.json")
	defer f.Close()
	if err != nil {
		logger.LogError("Can't open profile with error %s", err)
		return nil, err
	}

	s, err := ioutil.ReadAll(f)
	if err != nil {
		logger.LogError("Can't read profile with error %s", err)
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
	return cfg
}
