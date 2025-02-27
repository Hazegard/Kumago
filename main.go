package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/gen2brain/beeep"
	"os"
)

const URL = "https://status.xmc.ovh"

type Config struct {
	All           bool   `help:"Show all statuses" default:"false"`
	Xbar          bool   `help:"Show Xbar statuses" default:"false"`
	Notify        bool   `help:"Show notify statuses" default:"false"`
	DashboardPage string `help:"Dashboard page" default:"all" arg:""`
}

func main() {
	config := Config{}
	_ = kong.Parse(&config)

	if !CheckAvailability() {
		Error(fmt.Errorf("Dashboard unavailable: not connected to kuma"), config)
		return
	}
	titles, err := GetTitleDict(config.DashboardPage)
	if err != nil {
		Error(fmt.Errorf("Dashboard unavailable: %s", err), config)
		return
	}

	dashboard, err := GetDashboard(config.DashboardPage, titles)
	if err != nil {
		Error(fmt.Errorf("Dashboard unavailable: %s", err), config)
		return
	}

	length := 0
	for group, monitors := range dashboard {
		if dashboard.IsFullGreen(group, ignore) && !config.All {
			continue
		}
		for _, monitor := range monitors {
			l := len(monitor.Name)
			if l > length {
				length = l
			}
		}
	}
	content := ""
	xbar := ""
	if config.Xbar {
		xbar = " | font=\"FiraCode Nerd Font\""
	}

	globalState := OK

	for group, monitors := range dashboard {
		if dashboard.IsFullGreen(group, ignore) && !config.All {
			continue
		}
		content += fmt.Sprintf("\n%s\n", group.Name)
		for _, monitor := range monitors {
			if monitor.IsFullGreen() && !config.All {
				continue
			}
			content += fmt.Sprintf("%-*s  %s%s\n", length, monitor.Name, monitor.Beats(), xbar)
			if globalState == KO {
				continue
			}
			monitorStatus := monitor.analyzeStatus()
			if monitorStatus == KO {
				globalState = KO
				continue
			}
			if monitorStatus == Recovered && globalState == OK {
				globalState = Recovered
				continue
			}
		}
	}

	header := ""
	if config.Xbar {
		icon := "👌"
		switch globalState {
		case KO:
			icon = "🔥"
		case OK:
			icon = "👌"
		case Recovered:
			icon = "🤔"
		}
		header = fmt.Sprintf("%s\n---", icon)
		content = fmt.Sprintf("%s%s", header, content)
	}

	fmt.Print(content)

	if config.Notify && globalState == KO {
		err = beeep.Alert("BEEP", "beep", "beep")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func Error(err error, config Config) {
	if config.Xbar {
		fmt.Printf("🏩\n---\n")
	}
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(1)
}
