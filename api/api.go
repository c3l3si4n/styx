package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/c3l3si4n/styx/config"
)

var PollQueue = make(chan int, 1)
var API_URL = "https://labs.hackthebox.com/"

func GetMachineDetails(machineName string) (*config.Machine, error) {
	var machine config.MachineResponse

	url := fmt.Sprintf("%sapi/v4/machine/profile/%s", API_URL, machineName)
	fmt.Println("URL: ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HtbToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyStr := ""
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		bodyStr = string(bodyBytes)
	}
	fmt.Printf("Response: %+v\n", bodyStr)

	if err := json.Unmarshal([]byte(bodyStr), &machine); err != nil {
		return nil, err
	}

	machineDetails := machine.Info

	return &machineDetails, nil
}

func PollMachineDetails() {

	for {
		<-PollQueue
		fmt.Println("Polling machine details")
		machine, err := GetMachineDetails(config.SelectedMachine.Name)
		if err != nil {
			fmt.Println("Error polling machine details: ", err)
		}
		if machine.Name != "" {
			config.SelectedMachine.Details = *machine
			config.SelectedMachineFound = true
		} else {
			config.SelectedMachineFound = false
		}
		config.IsLoadingMachine = false
	}

}

func HydrateMachineDetails() {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			if config.SelectedMachine.Name != "" {
				select {
				case PollQueue <- 1:
				default:
				}
			}
		}
	}
}

type SubmitFlagRequest struct {
	Id   string `json:"id"`
	Flag string `json:"flag"`
}

func SubmitFlag(flag string, machineId int, machineType string) error {
	url := fmt.Sprintf("%sapi/v4/machine/own", API_URL)

	if machineType == "seasonal" {
		url = fmt.Sprintf("%sapi/v4/arena/own", API_URL)
	}

	submitFlagRequest := SubmitFlagRequest{
		Id:   fmt.Sprintf("%d", machineId),
		Flag: flag,
	}
	reqBody, err := json.Marshal(submitFlagRequest)
	fmt.Println("URL: ", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HtbToken))
	req.Header.Set("Content-Type", "application/json")

	// json body

	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody := ""
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else {
		respBody = string(bodyBytes)
	}
	fmt.Printf("Response: %+v\n", respBody)
	defer resp.Body.Close()

	return nil
}

type SpawnMachineRequest struct {
	Id int `json:"machine_id"`
}

type UserInfo struct {
	Info struct {
		CanAccessVip   bool `json:"canAccessVip"`
		isDedicatedVip bool `json:"isDedicatedVip"`
	} `json:"info"`
}

func GetUserInfo() (string, error) {
	url := fmt.Sprintf("%sapi/v4/user/info", API_URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HtbToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyStr := ""
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	} else {
		bodyStr = string(bodyBytes)
	}
	fmt.Printf("Response: %+v\n", bodyStr)

	var userInfo UserInfo
	if err := json.Unmarshal([]byte(bodyStr), &userInfo); err != nil {
		return "", err
	}

	fmt.Printf("User info: %+v\n", userInfo)

	if userInfo.Info.CanAccessVip {
		if userInfo.Info.isDedicatedVip {
			return "vip+", nil
		}
		return "vip", nil
	} else {
		return "free", nil
	}

}

func StartMachine(machineId int, machineType string) error {

	url := ""
	if config.MachineAPIType == "machine" {
		url = fmt.Sprintf("%sapi/v4/machine/start", API_URL)
	} else if config.MachineAPIType == "vm" {
		url = fmt.Sprintf("%sapi/v4/vm/spawn", API_URL)
	}

	if machineType == "seasonal" {
		url = fmt.Sprintf("%sapi/v4/arena/start", API_URL)
	}
	fmt.Println("URL: ", url)
	spawnMachineRequest := SpawnMachineRequest{
		Id: machineId,
	}
	reqBody, err := json.Marshal(spawnMachineRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HtbToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody := ""
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else {
		respBody = string(bodyBytes)
	}
	fmt.Printf("Response: %+v\n", respBody)
	defer resp.Body.Close()

	return nil
}

type StopMachineRequest struct {
	Id int `json:"machine_id"`
}

func StopMachine(machineId int, machineType string) error {
	url := ""
	if config.MachineAPIType == "machine" {
		url = fmt.Sprintf("%sapi/v4/machine/stop", API_URL)
	} else if config.MachineAPIType == "vm" {
		url = fmt.Sprintf("%sapi/v4/vm/terminate", API_URL)
	}

	if machineType == "seasonal" {
		url = fmt.Sprintf("%sapi/v4/arena/stop", API_URL)
	}
	fmt.Println("URL: ", url)
	stopMachineRequest := StopMachineRequest{
		Id: machineId,
	}
	reqBody, err := json.Marshal(stopMachineRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.HtbToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody := ""
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else {
		respBody = string(bodyBytes)
	}
	fmt.Printf("Response: %+v\n", respBody)
	defer resp.Body.Close()

	return nil
}
