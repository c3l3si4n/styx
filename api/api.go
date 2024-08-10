package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"strconv"
	"strings"
	"time"

	"github.com/c3l3si4n/styx/config"
)

var PollQueue = make(chan int, 1)
var API_URL = "https://labs.hackthebox.com/"

func makeRequest(method, url, body string, headers map[string]string) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func GetRequest(url string) (string, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.HtbToken),
	}
	return makeRequest("GET", url, "", headers)
}

func PostRequest(url, body string, headers map[string]string) (string, error) {
	return makeRequest("POST", url, body, headers)
}

func GetMachineDetails(machineName string) (*config.Machine, error) {
	url := fmt.Sprintf("%sapi/v4/machine/profile/%s", API_URL, machineName)
	bodyStr, err := GetRequest(url)
	if err != nil {
		return nil, err
	}

	var machine config.MachineResponse
	if err := json.Unmarshal([]byte(bodyStr), &machine); err != nil {
		return nil, err
	}

	return &machine.Info, nil
}

func PollMachineDetails() {
	for {
		<-PollQueue
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
			vpnRegionData := GetConnectionDetails()

			config.CurrentVPN = vpnRegionData

		}
	}
}

type SubmitFlagRequest struct {
	Id   string `json:"id"`
	Flag string `json:"flag"`
}

func SubmitFlag(flag string, machineId int, machineType string) bool {
	url := fmt.Sprintf("%sapi/v4/machine/own", API_URL)
	submitFlagRequest := SubmitFlagRequest{
		Id:   fmt.Sprintf("%d", machineId),
		Flag: flag,
	}
	if machineType == "seasonal" {
		url = fmt.Sprintf("%sapi/v4/arena/own", API_URL)
		submitFlagRequest = SubmitFlagRequest{
			Flag: flag,
		}
	}

	reqBody, err := json.Marshal(submitFlagRequest)
	if err != nil {
		return false
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.HtbToken),
		"Content-Type":  "application/json",
	}
	response, err := PostRequest(url, string(reqBody), headers)
	if err != nil {
		return false
	}
	fmt.Println(response)
	if strings.Contains(response, "is now owned.") {
		return true
	}
	return false

}

type SpawnMachineRequest struct {
	Id int `json:"machine_id"`
}

type UserInfo struct {
	Info struct {
		CanAccessVip   bool `json:"canAccessVip"`
		IsDedicatedVip bool `json:"isDedicatedVip"`
	} `json:"info"`
}

func GetUserInfo() (string, error) {
	url := fmt.Sprintf("%sapi/v4/user/info", API_URL)
	bodyStr, err := GetRequest(url)
	if err != nil {
		return "", err
	}

	var userInfo UserInfo
	if err := json.Unmarshal([]byte(bodyStr), &userInfo); err != nil {
		return "", err
	}

	if userInfo.Info.CanAccessVip {
		if userInfo.Info.IsDedicatedVip {
			return "vip+", nil
		}
		return "vip", nil
	}
	return "free", nil
}

func StartMachine(machineId int, machineType string) error {
	url := getMachineURL("start", machineType)
	spawnMachineRequest := SpawnMachineRequest{Id: machineId}
	reqBody, err := json.Marshal(spawnMachineRequest)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.HtbToken),
		"Content-Type":  "application/json",
	}
	response, err := PostRequest(url, string(reqBody), headers)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", response)
	return nil
}

type StopMachineRequest struct {
	Id int `json:"machine_id"`
}

func StopMachine(machineId int, machineType string) error {
	url := getMachineURL("stop", machineType)
	stopMachineRequest := StopMachineRequest{Id: machineId}
	reqBody, err := json.Marshal(stopMachineRequest)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.HtbToken),
		"Content-Type":  "application/json",
	}
	response, err := PostRequest(url, string(reqBody), headers)
	if err != nil {
		return err
	}

	log.Println("Stop machine response: ", response)
	return nil
}

func getMachineURL(action, machineType string) string {
	switch action {
	case "start":
		if machineType == "seasonal" {
			return fmt.Sprintf("%sapi/v4/arena/start", API_URL)
		}
		if config.MachineAPIType == "machine" {
			return fmt.Sprintf("%sapi/v4/machine/start", API_URL)
		}
		return fmt.Sprintf("%sapi/v4/vm/spawn", API_URL)
	case "stop":
		if machineType == "seasonal" {
			return fmt.Sprintf("%sapi/v4/arena/stop", API_URL)
		}
		if config.MachineAPIType == "machine" {
			return fmt.Sprintf("%sapi/v4/machine/stop", API_URL)
		}
		return fmt.Sprintf("%sapi/v4/vm/terminate", API_URL)
	default:
		return ""
	}
}

func GetConnectionDetails() config.VPNRegion {

	url := fmt.Sprintf("%sapi/v4/connection/status", API_URL)
	bodyStr, err := GetRequest(url)
	if err != nil {
		return config.VPNRegion{Status: "Disconnected"}
	}

	var vpnRegion config.VPNRegion
	var responseInterface []map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &responseInterface); err != nil {
		return config.VPNRegion{}
	}
	if len(responseInterface) == 0 {
		return config.VPNRegion{Status: "Disconnected"}
	}
	for _, region := range responseInterface {
		vpnRegion.IP = region["connection"].(map[string]interface{})["ip4"].(string)
		vpnRegion.Name = region["location_type_friendly"].(string)
		vpnRegion.Type = region["type"].(string)
		vpnRegion.Status = "Connected"
	}
	return vpnRegion

}

func FindVPNServer(region, listingType string, machineType string) int {
	// api/v4/connections/servers?product=competitive

	url := fmt.Sprintf("%sapi/v4/connections/servers?product=%s", API_URL, listingType)
	bodyStr, err := GetRequest(url)
	if err != nil {
		panic(err)
	}

	var responseInterface map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &responseInterface); err != nil {
		fmt.Println(bodyStr)

		panic(err)
	}
	// {"status":true,"data":{"disabled":false,"assigned":{"id":1,"friendly_name":"EU Free 1","current_clients":139,"location":"EU","location_type_friendly":"EU - Free"},"options":{"EU":{"EU - Free":{"location":"EU","name":"EU - Free","servers":{"1":{"id":1,"friendly_name":"EU Free 1","full":false,"current_clients":139,"location":"EU"},"201":{"id":201,"friendly_name":"EU Free 2","full":false,"current_clients":100,"location":"EU"},"253":{"id":253,"friendly_name":"EU Free 3","full":false,"current_clients":97,"location":"EU"}}},"EU - Release Arena":{"location":"EU","name":"EU - Release Arena","servers":{"267":{"id":267,"friendly_name":"EU Release Arena 1","full":false,"current_clients":56,"location":"EU"}}}},"US":{"US - Free":{"location":"US","name":"US - Free","servers":{"113":{"id":113,"friendly_name":"US Free 1","full":false,"current_clients":94,"location":"US"},"202":{"id":202,"friendly_name":"US Free 2","full":false,"current_clients":68,"location":"US"},"254":{"id":254,"friendly_name":"US Free 3","full":false,"current_clients":74,"location":"US"}}},"US - Release Arena":{"location":"US","name":"US - Release Arena","servers":{"268":{"id":268,"friendly_name":"US Release Arena 1","full":false,"current_clients":26,"location":"US"}}}},"AU":{"AU - Free":{"location":"AU","name":"AU - Free","servers":{"177":{"id":177,"friendly_name":"AU Free 1","full":false,"current_clients":34,"location":"AU"}}}},"SG":{"SG - Free":{"location":"SG","name":"SG - Free","servers":{"251":{"id":251,"friendly_name":"SG Free 1","full":false,"current_clients":62,"location":"SG"}}}}}}}

	options := responseInterface["data"].(map[string]interface{})["options"].(map[string]interface{})

	regionOptions := options[strings.ToUpper(region)].(map[string]interface{})
	region = strings.ToUpper(region)

	servers := regionOptions[region+" - "+machineType].(map[string]interface{})["servers"].(map[string]interface{})
	if machineType == "Release Arena" {
		// get keys
		for key := range servers {
			serverID, _ := strconv.Atoi(key)
			return serverID
		}
	} else if machineType == "Free" {
		// get keys
		lowestClients := 0
		lowestClientsId := 0
		for key, server := range servers {
			serverClients := int(server.(map[string]interface{})["current_clients"].(float64))
			if lowestClients == 0 || serverClients < lowestClients {
				lowestClients = serverClients
				lowestClientsId, _ = strconv.Atoi(key)
			}

		}
		return lowestClientsId
	}
	return 0
}

func SwitchVPNServer(serverId int, arena bool) error {
	url := fmt.Sprintf("%sapi/v4/connections/servers/switch/%d", API_URL, serverId)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", config.HtbToken),
		"Content-Type":  "application/json",
	}
	// {"arena":true}
	jsonBody := ""
	if arena {
		jsonBody = `{"arena":true}`
	} else {
		jsonBody = `{"arena":false}`
	}

	bodyStr, err := PostRequest(url, jsonBody, headers)
	if err != nil {
		return err
	}

	log.Println(bodyStr)
	config.VPNRegionCurrentID = int32(serverId)
	return nil
}

func WriteToFile(data string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}

func DownloadVPNFile(serverId int32) {
	// api/v4/access/ovpnfile/202/0
	url := fmt.Sprintf("%sapi/v4/access/ovpnfile/%d/0", API_URL, serverId)
	bodyStr, err := GetRequest(url)
	if err != nil {
		panic(err)
	}
	log.Println(bodyStr)

	err = WriteToFile(bodyStr, "/tmp/vpn.ovpn")

	if err != nil {

		panic(err)
	}

}

func DownloadFile(url string, outputPath string) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
}

func ConnectToVPN() {
	cmd := exec.Command("sudo", "openvpn", "/tmp/vpn.ovpn")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
	}()

}
