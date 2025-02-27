package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

const URL = "https://status.xmc.ovh"

type State int

const (
	KO State = iota
	OK
	Recovered
)

var (
	colors = map[string]int{
		"Black":   30,
		"Red":     31,
		"Green":   32,
		"Yellow":  33,
		"Blue":    34,
		"Magenta": 35,
		"Cyan":    36,
		"White":   37,
	}

	ignore = map[string]struct{}{
		"updog - Docker":              struct{}{},
		"UpDog - drop.newtechjob.com": struct{}{},
	}
)

type Config struct {
	All    bool `help:"Show all statuses" default:"false"`
	Xbar   bool `help:"Show Xbar statuses" default:"false"`
	Notify bool `help:"Show notify statuses" default:"false"`
}

type StatusTime time.Time

func (st *StatusTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02 15:04:05.000", s)
	if err != nil {
		return err
	}
	*st = StatusTime(t)
	return nil
}

func (st StatusTime) MarshalJSON() ([]byte, error) {
	t := time.Time(st)
	s := t.Format("2006-01-02 15:04:05.000")
	return []byte(s), nil
}

type Status struct {
	Status State      `json:"status"`
	Date   StatusTime `json:"time"`
	Msg    string     `json:"msg"`
	Ping   float64    `json:"ping"`
}

func (st *Status) Beat() string {
	color := 0
	switch st.Status {
	case KO:
		color = colors["Red"]
	case OK:
		color = colors["Green"]
	case Recovered:
		color = colors["Yellow"]
	default:
		color = colors["White"]
	}
	return fmt.Sprintf("\033[%dm%s\033[0m", color, "█")
}

func (st *Status) HasDowntime() bool {
	return st.Status == 0
}

type KumaHeartBeatList map[string][]Status

type UptimeList map[string]float64
type All struct {
	Uptime    UptimeList        `json:"uptimeList"`
	HeartBeat KumaHeartBeatList `json:"heartbeatList"`
}

type Group struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type HeartBeatList map[Group][]Monitor

func (hbl *HeartBeatList) IsFullGreen(group Group, ignore map[string]struct{}) bool {
	monitors := (*hbl)[group]
	for _, monitor := range monitors {
		if _, ok := ignore[monitor.Name]; ok {
			continue
		}
		for _, status := range monitor.Status {
			if status.Status != 1 {
				return false
			}
		}
	}
	return true
}

type Monitor struct {
	Id     string
	Name   string
	Status []Status
}

func (m *Monitor) analyzeStatus() State {
	n := len(m.Status)
	if n == 0 {
		return OK
	}

	// Check if status recovered
	if m.Status[n-1].Status == Recovered {
		return Recovered
	}
	if m.Status[n-1].Status == KO {
		return KO
	}
	for i := len(m.Status) - 1; i >= 0; i-- {
		if i == len(m.Status)-1 {
			// explicitly skip the last element as here it is always OK
			continue
		}
		if m.Status[i].Status == KO {
			return Recovered
		}
	}
	return OK
}

func (m *Monitor) HasResolvedDowntime() bool {

	for i := len(m.Status) - 1; i >= 0; i-- {

	}
	for _, status := range m.Status {
		if status.HasDowntime() {
			return true
		}
	}
	return false
}

func (m *Monitor) HasDowntime() bool {
	for _, status := range m.Status {
		if status.HasDowntime() {
			return true
		}
	}
	return false
}

func (m *Monitor) Beats() string {
	sb := strings.Builder{}
	for _, status := range m.Status {
		sb.WriteString(status.Beat())
	}
	return sb.String()
}
func (m *Monitor) IsFullGreen() bool {
	for _, status := range m.Status {
		if status.Status != 1 {
			return false
		}
	}
	return true
}

func main() {
	config := Config{}
	_ = kong.Parse(&config)

	if !CheckAvailability() {
		fmt.Println("🏩")
		return
	}
	titles, err := GetTitleDict()
	if err != nil {
		fmt.Println(err)
		return
	}

	all, err := GetAll(titles)
	if err != nil {
		fmt.Println(err)
		return
	}

	length := 0
	for group, monitors := range all {
		if all.IsFullGreen(group, ignore) && !config.All {
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

	for group, monitors := range all {
		if all.IsFullGreen(group, ignore) && !config.All {
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
