package main

import (
	"fmt"
	"strings"
	"time"
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
)

type State int

const (
	KO State = iota
	OK
	Recovered
)

var (
	ignore = map[string]struct{}{
		"updog - Docker":                                      struct{}{},
		"UpDog - drop.newtechjob.com":                         struct{}{},
		"python-http - Docker":                                struct{}{},
		"Python HTTP - py.newtechjob.com":                     struct{}{},
		"Python HTTPS - py.newtechjob.com":                    struct{}{},
		"sslipio - Docker":                                    struct{}{},
		"SSLIP DNS subdomain - X.X.X.X.rebind.newtechjob.com": struct{}{},
		"SSLIP DNS subdomain - X.X.X.X.rebind.xmc.ovh":        struct{}{},
	}
)

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
type Dashboard struct {
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

func (m *Monitor) analyzeStatus(ignore map[string]struct{}) (State, State) {
	ignored := false
	if _, ok := ignore[m.Name]; ok {
		ignored = true
	}
	n := len(m.Status)
	if n == 0 {
		return OK, OK
	}

	// Check if status recovered
	if m.Status[n-1].Status == Recovered {
		return Recovered, Recovered
	}
	if m.Status[n-1].Status == KO {
		if ignored {
			return Recovered, OK
		}
		return KO, KO
	}
	for i := len(m.Status) - 1; i >= 0; i-- {
		if i == len(m.Status)-1 {
			// explicitly skip the last element as here it is always OK
			continue
		}
		if m.Status[i].Status == KO {
			return Recovered, Recovered
		}
	}
	return OK, OK
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
