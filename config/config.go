package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"scp_delegator/logger"
)

type Config struct {
	Version   string `json:"scp_version"`
	Period    uint32 `json:"execution_period"`
	FreeSpace uint32 `json:"free_disk_space_GB"`
	LogLevl   string `json:"log_level"`
	AsyncRun  bool   `json:"async_execute"`
	Tasks     []Task `json:"tasks"`
}
type Task struct {
	Id       uint32 `json:"id"`
	Name     string `json:"name"`
	Condtion uint32 `json:"condition"`
	Action   uint32 `json:"action"`
}
type Templates struct {
	Actions              []Action             `json:"actions"`
	ActionsProperties    []ActionsProperty    `json:"action_properties"`
	Conditions           []Condition          `json:"condtions"`
	ConditionsProperties []ConditionsProperty `json:"condition_properties"`
}
type Action struct {
	Id         uint32     `json:"id"`
	PreAction  uint32     `json:"pre_action"`
	Name       string     `json:"name"`
	Executable string     `json:"executable"`
	Arguments  []Argument `json:"arguments"`
	PostAction uint32     `json:"post_action"`
	Policy     uint32     `json:"policy"`
}
type Argument struct {
	Command string `json:"command"`
	Value   string `json:"value"`
}
type ActionsProperty struct {
	Id       uint32         `json:"id"`
	TimeoutS uint32         `json:"timeout_sec"`
	PeriodS  uint32         `json:"period_sec"`
	Repeat   RepeatProperty `json:"repeat"`
}
type RepeatProperty struct {
	Count     uint32 `json:"count"`
	IntervalS uint32 `json:"interval_sec"`
}

type Condition struct {
	Id     uint32 `json:"id"`
	Name   string `json:"name"`
	Target string `json:"monitor_process"`
}
type ConditionsProperty struct {
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

func ParseProfile() *Config {
	cfg := &Config{}
	profile, err := readProfile()
	if err != nil {
		return nil
	}
	json.Unmarshal(profile, cfg)
	return cfg
}
