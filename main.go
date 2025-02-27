package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/gen2brain/beeep"
	"os"
	"sort"
	"strings"
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
	for _, monitors := range dashboard {
		for _, monitor := range monitors {
			if monitor.IsFullGreen() && !config.All {
				continue
			}
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

	groups := []Group{}
	for group := range dashboard {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})
	for _, group := range groups {
		monitors := dashboard[group]
		if dashboard.IsFullGreen(group, map[string]struct{}{}) && !config.All {
			continue
		}

		content += fmt.Sprintf("\n%s\n", group.Name)
		sort.Slice(monitors, func(i, j int) bool {
			return monitors[i].Name < monitors[j].Name
		})

		for _, monitor := range monitors {
			if monitor.IsFullGreen() && !config.All {
				continue
			}
			icon := ""
			localStatus, globalStatus := monitor.analyzeStatus(ignore)
			switch localStatus {
			case OK:
				icon = "👌"
			case KO:
				icon = "🔥"
			case Recovered:
				icon = "🤔"
			}
			nb := countBlockChar(monitor.Beats())
			pad := 50 - nb
			if pad < 0 {
				pad = 0
			}

			content += fmt.Sprintf("%-*s  %s %-*s%s %s\n", length, monitor.Name, icon, pad, "", monitor.Beats(), xbar)
			if globalState == KO {
				continue
			}

			if globalStatus == KO {
				globalState = KO
				continue
			}
			if globalStatus == Recovered && globalState == OK {
				globalState = Recovered
				continue
			}
		}
	}

	content = strings.TrimSpace(content)

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
		content = fmt.Sprintf("%s\n%s\nRefresh... | refresh=true", header, content)
	}

	fmt.Print(content)

	if config.Notify && globalState == KO {
		err = beeep.Alert("BEEP", "beep", "beep")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func countBlockChar(s string) int {
	count := 0
	for _, r := range s {
		if r == '█' {
			count++
		}
	}
	return count
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
