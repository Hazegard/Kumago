package main

import (
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const APP_NAME = "kumago"

type Config struct {
	All           bool     `help:"Show all statuses" default:"false"`
	Xbar          bool     `help:"Show Xbar statuses" default:"false"`
	Notify        bool     `help:"Show notify statuses" default:"false"`
	Url           *url.URL `help:"Kuma URL" default:"" short:"u"`
	DashboardPage string   `help:"Dashboard page" default:"all" arg:""`
	IgnoreList    []string `help:"Ignore list" short:"i"`
	NotifyUrl     []string `help:"Discord URL" default:""`
}

func main() {
	config := Config{}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configSearchDir := []string{
		filepath.Join(dir, strings.ToLower(APP_NAME)+".yaml"),
	}
	home, err := os.UserHomeDir()
	if err == nil {
		configSearchDir = append(configSearchDir,
			filepath.Join(home, ".config", strings.ToLower(APP_NAME), strings.ToLower(APP_NAME)+".yaml"),
			filepath.Join(home, ".config", strings.ToLower(APP_NAME)+".yaml"),
		)
	}
	kongOptions := []kong.Option{
		kong.Name(APP_NAME),
		kong.UsageOnError(),
		kong.Configuration(YAML, configSearchDir...),
		kong.DefaultEnvars(strings.ToUpper(APP_NAME)),
	}
	_ = kong.Parse(&config, kongOptions...)
	if !CheckAvailability(config.Url) {
		Error(fmt.Errorf("Dashboard unavailable: not connected to kuma"), config)
		return
	}
	titles, err := GetTitleDict(config.DashboardPage, config.Url)
	if err != nil {
		Error(fmt.Errorf("Dashboard unavailable: %s", err), config)
		return
	}

	dashboard, err := GetDashboard(config.DashboardPage, titles, config.Url)
	if err != nil {
		Error(fmt.Errorf("Dashboard unavailable: %s", err), config)
		return
	}

	groups := []Group{}
	for group := range dashboard {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	content, globalState, downList := Parse(config, groups, dashboard)
	fmt.Println(content)

	for group, heartbeat := range downList {
		fmt.Println(group.Name)
		for _, heartbeatItem := range heartbeat {
			fmt.Println("    " + heartbeatItem.Name)
		}
	}

	Notify(downList, config)

	if globalState == KO {

	}
	// if config.Notify && globalState == KO {
	// 	err = beeep.Alert("BEEP", "beep", "beep")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }
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

func YAML(r io.Reader) (kong.Resolver, error) {
	decoder := yaml.NewDecoder(r)
	config := map[string]interface{}{}
	err := decoder.Decode(config)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("YAML agent decode error: %w", err)
	}
	return kong.ResolverFunc(func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		for _, env := range flag.Envs {
			_, ok := os.LookupEnv(env)
			if ok {
				return nil, nil
			}
		}
		// Build a string path up to this flag.
		path := []string{}
		for n := parent.Node(); n != nil && n.Type != kong.ApplicationNode; n = n.Parent {
			path = append([]string{n.Name}, path...)
		}
		path = append(path, flag.Name)
		path = strings.Split(strings.Join(path, "-"), "-")
		return find(config, path), nil
	}), nil
}

func find(config map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return config
	}
	for i := 0; i < len(path); i++ {
		prefix := strings.Join(path[:i+1], "-")
		if child, ok := config[prefix].(map[string]interface{}); ok {
			return find(child, path[i+1:])
		}
	}
	return config[strings.Join(path, "-")]
}

func Notify(hb HeartBeatList, c Config) error {
	if len(hb) == 0 {
		return nil
	}
	if c.NotifyUrl == nil {
		return nil
	}
	err, notifier := NewNotifier(c)

	if err != nil {
		return err
	}
	notifier.Notify(hb)
	return nil
}

func Parse(config Config, groups []Group, dashboard HeartBeatList) (string, State, HeartBeatList) {

	globalState := OK

	downList := HeartBeatList{}
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
			localStatus, globalStatus := monitor.analyzeStatus(config.IgnoreList)
			switch localStatus {
			case OK:
				icon = "👌"
			case KO:
				icon = "🔥"
			case Recovered:
				icon = "🤔"
			}
			if localStatus == KO /*|| localStatus == Recovered*/ {
				//				downList[group] = []Monitor{}
				downList[group] = append(downList[group], monitor)
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

	//fmt.Println(content)
	return content, globalState, downList
}
