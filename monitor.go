package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
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

func (s *State) String() string {
	switch *s {
	case OK:
		return "OK"
	case WarnOk:
		return "WARN_OK"
	case Warn:
		return "WARN"
	case KO:
		return "KO"
	}
	return "UNKNOWN"
}
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
	WarnOk
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
type HeartBeatList map[Group][]*Monitor

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
	Id                string
	Name              string
	Status            []Status
	localState        State
	globalState       State
	analyzeStatusSync sync.Once
}

func IsInList(name string, list []string, regexList []*regexp.Regexp) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}

	for _, regex := range regexList {
		if regex.MatchString(name) {
			return true
		}
	}
	return false
}

func (m *Monitor) analyzeStatus(ignoreConf IgnoreConfig) (State, State) {
	m.analyzeStatusSync.Do(func() {
		ignored := IsInList(m.Name, ignoreConf.Ignore, ignoreConf.RegexList)
		onlyLast := IsInList(m.Name, ignoreConf.Onlylast, ignoreConf.OnlyLastRegexList)

		// If the monitor is empty (no state has been reported to uptime-kuma).
		// We consider it as OK and return
		if len(m.Status) == 0 {
			m.localState = OK
			m.globalState = OK
			return
		}

		lastState := m.Status[len(m.Status)-1].Status
		// Only the last status is relevant.
		if onlyLast {
			// If the last status is either KO or Warn and the monitor is ignored,
			// we set the global state to OK, and the local state to Warn.
			// We then exit as we do not need further processing
			if ignored && lastState == KO || lastState == Warn {
				m.localState = Warn
				m.globalState = OK
				return
			}

			// If the monitor is not ignored and either KO or Warn,
			// We set the local and global states to the last status and return
			// We then exit as we do not need further processing
			if lastState == KO || lastState == Warn {
				m.localState = lastState
				m.globalState = lastState
				return
			}

			// Else, the last status is OK, so we set the local and global states to OK.
			// But we do not return here as we want to check for earlier KO or warn statuses
			// To update accordingly the local stats (the global state should not be updated afterward)
			// If the monitor is in onlyLast mode.
			m.localState = OK
			m.globalState = OK
		}

		// If all states of the monitor are analyzed,
		// We first check if the last state is Warn or KO
		if lastState == Warn || lastState == KO {
			var globalState State

			if ignored {
				// If the monitor is ignored, we set the global state to OK, and the local state to Warn.
				m.localState = Warn
				globalState = OK
			} else {
				// Else, we set the local and global states to the last status.
				m.localState = lastState
				globalState = lastState
			}

			if !onlyLast {
				// If the monitor is not in onlyLast mode, we update the global state accordingly.
				m.globalState = globalState
			}
			// We do not need to process further the monitor as we have already set the local and global states
			// And these states will no change
			return
		}

		// We start to iterate over the status list from the last element to the first element
		// Here we know that the last status is always OK
		isWarnOk := false
		for i := len(m.Status) - 1; i >= 0; i-- {
			if i == len(m.Status)-1 {
				// explicitly skip the last element as here it is always OK
				continue
			}
			if m.Status[i].Status == KO {
				// If we find a KO or Warn status, we set the local state to warn
				m.localState = Warn

				if ignored && !onlyLast {
					// If the monitor is ignored and not in onlyLast mode, we set the global state to OK
					m.globalState = OK
					return
				}
				// Else if the monitor is not ignored and still not in onlyLast mode, we set the global state to warn
				if !onlyLast {
					m.globalState = Warn
					return
				}
				// Note: we never set here a state to KO as the monitor is currently OK
				// Here we just want to highlight a minor issue:
				// Either the monitor was down in the timeframe displayed by uptime-kuma,
				//
				return
			}
			if m.Status[i].Status == Warn {
				isWarnOk = true
			}
		}
		if isWarnOk {
			m.localState = WarnOk
		} else {
			// If we reach this point, it means that the monitor is currently OK
			m.localState = OK
		}
		// If the monitor is not in onlyLast mode, we set the global state to OK
		if !onlyLast {
			m.globalState = OK
		}
	})

	return m.localState, m.globalState
}

func (m *Monitor) CheckFinalStatus(state State, ignored bool, onlyLast bool) bool {
	if m.Status[len(m.Status)-1].Status == state {
		m.localState = Warn
		if ignored && !onlyLast {
			m.globalState = OK
			return true
		}
		m.globalState = state
		if !onlyLast {
			m.globalState = state
		}
		return true
	}
	return false
}

func (m *Monitor) Beats(c Config) string {
	if !c.Beat {
		return ""
	}
	sb := strings.Builder{}
	for _, status := range m.Status {
		if c.BeatEmoji && c.Emoji {
			sb.WriteString(c.Symbol.GetBeatEmoji(status.Status))
		} else {
			sb.WriteString(c.Symbol.GetBeat(status.Status, c.Color))
		}
	}
	return sb.String()
}

func (m *Monitor) GetName(length int, c Config) string {
	color := ""
	status, _ := m.analyzeStatus(c.IgnoreConfig)
	switch status {
	case OK:
		color = c.Color.OkBeat
	case WarnOk:
		color = c.Color.OkBeat
	case Warn:
		color = c.Color.WarnBeat
	case KO:
		color = c.Color.KoBeat
	}
	if !c.Beat && !c.Emoji {
		length = 0
	}
	return fmt.Sprintf("\u001B[%dm%-*s\u001B[0m", colors[color], length, m.Name)
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
