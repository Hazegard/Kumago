package main

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	colors = map[string]int{
		"black":   30,
		"red":     31,
		"green":   32,
		"yellow":  33,
		"blue":    34,
		"magenta": 35,
		"cyan":    36,
		"white":   37,
	}
)

type State int

func (s *State) Min(t State) State {
	if *s > t {
		return t
	}
	return *s
}

func (s *State) UnmarshalJSON(data []byte) error {
	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value {
	case 0:
		*s = KO
	case 1:
		*s = OK
	case 2:
		*s = Warn
	default:
		return errors.New("invalid state value")
	}

	return nil
}

const (
	KO State = iota
	Warn
	OK
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

func (hbl *HeartBeatList) IsOK(group Group, ignore map[string]struct{}) bool {
	monitors := (*hbl)[group]
	for _, monitor := range monitors {
		if _, ok := ignore[monitor.Name]; ok {
			continue
		}
		for _, status := range monitor.Status {
			if status.Status != OK {
				return false
			}
		}
	}
	return true
}
func (hbl *HeartBeatList) IsKO(group Group, ignore map[string]struct{}) bool {
	monitors := (*hbl)[group]
	for _, monitor := range monitors {
		if _, ok := ignore[monitor.Name]; ok {
			continue
		}
		for _, status := range monitor.Status {
			if status.Status != OK {
				return false
			}
		}
	}
	return true
}

func (hbl *HeartBeatList) IsWarn(group Group, ignore map[string]struct{}) bool {
	monitors := (*hbl)[group]
	for _, monitor := range monitors {
		if _, ok := ignore[monitor.Name]; ok {
			continue
		}
		for _, status := range monitor.Status {
			if status.Status != Warn {
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

func (m *Monitor) analyzeStatus(ignoreList []string, ignoreRegex []*regexp.Regexp) (State, State) {
	var ignore = map[string]struct{}{}
	for _, ignoreStr := range ignoreList {
		ignore[ignoreStr] = struct{}{}
	}
	ignored := false
	if _, ok := ignore[m.Name]; ok {
		ignored = true
	}

	for _, regex := range ignoreRegex {
		if regex.MatchString(m.Name) {
			ignored = true
			break
		}
	}

	n := len(m.Status)
	if n == 0 {
		return OK, OK
	}

	// Check if status warn
	if m.Status[n-1].Status == Warn {
		if ignored {
			return Warn, OK
		}
		return Warn, Warn
	}
	if m.Status[n-1].Status == KO {
		if ignored {
			return Warn, OK
		}
		return KO, KO
	}
	for i := len(m.Status) - 1; i >= 0; i-- {
		if i == len(m.Status)-1 {
			// explicitly skip the last element as here it is always OK
			continue
		}
		if m.Status[i].Status == KO {
			if ignored {
				return Warn, OK
			}
			return Warn, Warn
		}
	}
	return OK, OK
}

func (m *Monitor) Beats(c Config) string {
	sb := strings.Builder{}
	for _, status := range m.Status {
		if c.BeatEmoji && c.Emoji {
			sb.WriteString(c.Symbol.GetBeatEmoji(status.Status))
		} else {
			sb.WriteString(c.Symbol.GetBeat(status.Status))
		}
	}
	return sb.String()
}

func (m *Monitor) EmojiBeats(c Config) string {
	sb := strings.Builder{}
	for _, status := range m.Status {
		sb.WriteString(c.Symbol.GetBeatEmoji(status.Status))
	}
	return sb.String()
}

func (m *Monitor) IsOK() bool {
	for _, status := range m.Status {
		if status.Status != OK {
			return false
		}
	}
	return true
}

func (m *Monitor) IsWarn() bool {
	if len(m.Status) == 0 {
		return false
	}
	if m.Status[len(m.Status)-1].Status == KO || m.Status[len(m.Status)-1].Status == OK {
		return false
	}
	for _, status := range m.Status {
		if status.Status == Warn || status.Status == KO {
			return false
		}
	}
	return false
}

func (m *Monitor) IsKO() bool {
	for _, status := range m.Status {
		if status.Status == OK || status.Status == Warn {
			return false
		}
	}
	return true
}
