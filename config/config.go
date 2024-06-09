package config

type Machine struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	IP          string `json:"ip"`
	Solved      bool   `json:"isCompleted"`
	MachineMode string `json:"machine_mode"`
	PlayInfo    struct {
		IsActive bool `json:"isActive"`
	} `json:"playInfo"`
}
type MachineResponse struct {
	Info Machine `json:"info"`
}

type SelectedMachineType struct {
	Name    string
	Found   bool
	Details Machine
}

type VPNRegion struct {
	Region string
	Name   string
	Type   string
	Status string
	IP     string
}

var MachineAPIType string
var SelectedMachine SelectedMachineType = SelectedMachineType{}
var CurrentVPN VPNRegion
var SelectedMachineFound = false
var AutoFlagSubmit bool = false
var IsLoadingMachine bool = false
var HtbToken = ""
var SelectedTab = 0
var ServerEnabled = false
var VPNRegionCurrentID int32
var VPNRegionSelected int32
var VPNTypeSelected int32
