package main

import (
	"fmt"
	"log"
	"os"

	"github.com/c3l3si4n/styx/api"
	"github.com/c3l3si4n/styx/config"
	autosubmit "github.com/c3l3si4n/styx/tools/auto_submit"
	"github.com/c3l3si4n/styx/tools/bingo"
	"github.com/c3l3si4n/styx/utils"

	g "github.com/AllenDang/giu"
	"github.com/fstanis/screenresolution"
)

var possibleVPN = []string{"Release Arena", "Free"}
var possibleRegions = []string{"us", "eu"}
var (
	sashPos1 float32 = 200
	sashPos2 float32 = 200
	sashPos3 float32 = 200
	sashPos4 float32 = 200
	t32              = 100
)

type Machine struct {
	Name string
	IP   string
}

func selectedMachineChanged() {
	go func() {

		config.IsLoadingMachine = true
		//machine, err := api.GetMachineDetails(config.SelectedMachine.Name)
		api.PollQueue <- 1

	}()

}

func vpnDataChanged() {
	selectedVPNRegion := possibleRegions[config.VPNRegionSelected]
	selectedVPNType := possibleVPN[config.VPNTypeSelected]
	vpnListingType := "labs"
	if selectedVPNType == "Release Arena" {
		vpnListingType = "competitive"
	}
	serverId := api.FindVPNServer(selectedVPNRegion, vpnListingType, selectedVPNType)
	if serverId == 0 {
		fmt.Println("Server not found")
		return
	}
	api.SwitchVPNServer(serverId, true)
	api.DownloadVPNFile(config.VPNRegionCurrentID)
	api.ConnectToVPN()

	log.Println("Switched to server", serverId)
}
func drawMachineTab() g.Layout {
	initialLayout := g.Layout{}

	// add label

	// add input text
	selectedMachineInput := g.Layout{}
	if config.IsLoadingMachine {
		selectedMachineInput = append(selectedMachineInput, g.Label("Loading..."))
	} else {
		selectedMachineInputRow := g.Row(
			g.Label("Machine:"),
			g.InputText(&config.SelectedMachine.Name),
			g.Event().OnKeyReleased(g.KeyEnter, selectedMachineChanged),
			g.Button("Hack!"),
			g.Event().OnClick(g.MouseButtonLeft, selectedMachineChanged),
		)
		selectedMachineInput = append(selectedMachineInput, selectedMachineInputRow)

	}
	initialLayout = append(initialLayout, g.Separator())

	initialLayout = append(initialLayout, selectedMachineInput)

	contentWindow := g.Layout{}
	if config.SelectedMachineFound {
		ControlRow := g.Row(
			g.Button("Start"),
			g.Event().OnClick(g.MouseButtonLeft, func() {
				api.StartMachine(config.SelectedMachine.Details.Id, config.SelectedMachine.Details.MachineMode)
			}),
			g.Button("Stop"),
			g.Event().OnClick(g.MouseButtonLeft, func() {
				api.StopMachine(config.SelectedMachine.Details.Id, config.SelectedMachine.Details.MachineMode)
			}),
		)
		contentWindow = append(contentWindow, ControlRow)
		contentWindow = append(contentWindow, g.Separator())

		contentWindow = append(contentWindow, g.Label(fmt.Sprintf("Name: %s", config.SelectedMachine.Details.Name)))
		machineIp := config.SelectedMachine.Details.IP
		machineIpWidgets := []g.Widget{
			g.Label(fmt.Sprintf("IP: %s", machineIp)),
		}
		if machineIp != "" {
			machineIpWidgets = append(machineIpWidgets, g.Button("Copy").OnClick(func() {
				utils.SetClipboard(machineIp)
			}))
		}
		contentWindow = append(contentWindow, g.Row(machineIpWidgets...))
		contentWindow = append(contentWindow, g.Label(fmt.Sprintf("OS: %s", config.SelectedMachine.Details.OS)))
		contentWindow = append(contentWindow, g.Label(fmt.Sprintf("Solved: %t", config.SelectedMachine.Details.Solved)))
		contentWindow = append(contentWindow, g.Label(fmt.Sprintf("Mode: %s", config.SelectedMachine.Details.MachineMode)))
		contentWindow = append(contentWindow, g.Label(fmt.Sprintf("Active: %t", config.SelectedMachine.Details.PlayInfo.IsActive)))
		contentWindow = append(contentWindow, g.Checkbox("Auto flag submit", &config.AutoFlagSubmit))

		// separator
		// show current vpn

	}

	contentWindow = append(contentWindow, g.Separator())
	contentWindow = append(contentWindow, g.Label(fmt.Sprintf("Current VPN: %s", config.CurrentVPN.Name)))
	vpnIp := config.CurrentVPN.IP
	vpnIpWidgets := []g.Widget{
		g.Label(fmt.Sprintf("VPN IP: %s", vpnIp)),
	}
	if vpnIp != "" {
		vpnIpWidgets = append(vpnIpWidgets, g.Button("Copy").OnClick(func() {
			utils.SetClipboard(vpnIp)
		}))
	}
	contentWindow = append(contentWindow, g.Row(vpnIpWidgets...))
	contentWindow = append(contentWindow, g.Label(fmt.Sprintf("VPN Status: %s", config.CurrentVPN.Status)))

	vpnCombo := g.Layout{}

	vpnCombo = append(vpnCombo,
		g.Combo("VPN Region", possibleRegions[config.VPNRegionSelected], possibleRegions, &config.VPNRegionSelected).OnChange(vpnDataChanged),
	)
	vpnCombo = append(vpnCombo,
		g.Combo("VPN Type", possibleVPN[config.VPNTypeSelected], possibleVPN, &config.VPNTypeSelected).OnChange(vpnDataChanged),
	)

	contentWindow = append(contentWindow, vpnCombo)

	initialLayout = append(initialLayout, contentWindow)

	return initialLayout
}

func drawToolsTab() g.Layout {
	initialLayout := g.Layout{}
	if !config.ServerEnabled {
		button := g.Layout{
			g.Button("Start bingo listener"),
			g.Event().OnClick(g.MouseButtonLeft, func() {
				if !config.ServerEnabled {
					bingo.Bingo()
					config.ServerEnabled = true
				} else {
					fmt.Println("Server already enabled")
				}
			}),
		}

		initialLayout = append(initialLayout, button)
	} else {
		layout := g.Layout{}
		layout = append(layout, g.Label("Bingo server is running"))

		ipAddress, err := bingo.GetInterfaceIpv4Addr("tun0")
		if err != nil {
			layout = append(layout, g.Label("Error getting HTB VPN address, are you connected to openvpn?"))
		} else {
			layout = append(layout, g.Label(fmt.Sprintf("HTB IP: %s", ipAddress)))

			layout = append(layout, g.Row(g.Layout{
				g.Label(fmt.Sprintf("Linux: http://%s:61234/lin", ipAddress)),
				g.Button("Copy"),
				g.Event().OnClick(g.MouseButtonLeft, func() {
					utils.SetClipboard(fmt.Sprintf("http://%s:61234/lin", ipAddress))
				}),
			}))
			layout = append(layout, g.Row(
				g.Layout{
					g.Label(fmt.Sprintf("Windows: http://%s:61234/win", ipAddress)),
					g.Button("Copy"),
					g.Event().OnClick(g.MouseButtonLeft, func() {
						utils.SetClipboard(fmt.Sprintf("http://%s:61234/win", ipAddress))
					}),
				},
			))
		}

		initialLayout = append(initialLayout, layout)
	}

	return initialLayout
}

func drawLayout() g.Layout {
	initialLayout := g.Layout{}
	initialLayout = append(initialLayout,
		g.Row(
			g.Label("Styx HTB GUI v0.0.1"),
			g.Button("Machine"),
			g.Event().OnClick(g.MouseButtonLeft, func() {
				config.SelectedTab = 0
			}),
			g.Button("Tools"),
			g.Event().OnClick(g.MouseButtonLeft, func() {
				config.SelectedTab = 1
			}),
		),
	)
	initialLayout = append(initialLayout, g.Separator())

	switch config.SelectedTab {
	case 0:
		initialLayout = append(initialLayout, drawMachineTab())
	case 1:
		initialLayout = append(initialLayout, drawToolsTab())
	}
	return initialLayout
}

func loop() {

	g.SingleWindow().Layout(
		drawLayout(),
	)

	g.Update()
}

func setup() {
	htbToken := os.Getenv("HTB_TOKEN")
	if htbToken == "" {
		panic("HTB_TOKEN environment variable is not set")
	}
	config.HtbToken = htbToken

	// set subscription type
	accountType, err := api.GetUserInfo()
	if err != nil {
		panic("Error getting user info")
	}
	if accountType == "vip" || accountType == "vip+" {
		config.MachineAPIType = "vm"
	} else {
		config.MachineAPIType = "machine"
	}
}

func main() {
	resolution := screenresolution.GetPrimary()
	setup()
	go api.HydrateMachineDetails()
	go api.PollMachineDetails()
	autosubmit.StartSubmitter()
	wnd := g.NewMasterWindow("Styx", 800, resolution.Height/2, g.MasterWindowFlagsNotResizable)
	wnd.Run(loop)
}
