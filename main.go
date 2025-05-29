package main

import (
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/rivo/uniseg"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const APP_NAME = "kumago"

type Color struct {
	WarnBeat string `yaml:"ko" default:"yellow" help:"Terminal color used to display a warn beat"`
	OkBeat   string `yaml:"ko" default:"green" help:"Terminal color used to display an OK beat"`
	KoBeat   string `yaml:"ko" default:"red" help:"Terminal color used to display a KO beat"`
}

type Symbol struct {
	TermIcon string `yaml:"ko" default:"█" help:"Symbol used to display a beat"`

	Warn  string `yaml:"warn" default:"🤔" help:"Symbol used to indicate a warning state"`
	Ok    string `yaml:"ok" default:"👌" help:"Symbol used to indicate an OK state"`
	Ko    string `yaml:"ko" default:"🔥" help:"Symbol used to indicate a KO state"`
	Error string `yaml:"ko" default:"🏩" help:"Symbol used to indicate an error state"`

	WarnBeatEmoji string `yaml:"ko" default:"🟧" help:"Emoji used to display a warn beat"`
	OkBeatEmoji   string `yaml:"ko" default:"🟩" help:"Emoji used to display an OK beat"`
	KoBeatEmoji   string `yaml:"ko" default:"🟥" help:"Emoji used to display a KO beat"`
}

func (s *Symbol) Get(state State) string {
	switch state {
	case OK:
		return s.Ok
	case KO:
		return s.Ko
	case Warn:
		return s.Warn
	}
	return " "
}

func (s *Symbol) GetBeat(state State, c Color) string {
	switch state {
	case OK:
		return fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[strings.ToLower(c.OkBeat)], s.TermIcon)
	case KO:
		return fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[strings.ToLower(c.KoBeat)], s.TermIcon)
	case Warn:
		return fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[strings.ToLower(c.WarnBeat)], s.TermIcon)
	}
	return " "
}

func (s *Symbol) GetBeatEmoji(state State) string {
	switch state {
	case OK:
		return s.OkBeatEmoji
	case KO:
		return s.KoBeatEmoji
	case Warn:
		return s.WarnBeatEmoji
	}
	return " "
}

type Config struct {
	Status          []string         `help:"Show statuses" default:"KO,Warn"`
	Xbar            bool             `help:"Show Xbar statuses" default:"false"`
	Notify          bool             `help:"Show notify statuses" default:"false"`
	Url             *url.URL         `help:"Kuma URL" default:"" short:"u"`
	DashboardPage   []string         `help:"Dashboard page" default:"all" arg:""`
	IgnoreList      []string         `help:"Ignore list" short:"i"`
	IgnoreRegexList []string         `help:"Ignore list (regex)" short:"I"`
	RegexList       []*regexp.Regexp `kong:"-"`
	NotifyUrl       []string         `help:"Discord URL" default:""`
	Beat            bool             `help:"Show/hide heartbeat" negatable:"" default:"true"`
	BeatEmoji       bool             `help:"Use emoji" default:"false"`
	Emoji           bool             `help:"Use emoji" default:"true" negatable:""`
	Color           Color            `help:"Color" default:"" embed:"" prefix:"color-"`
	Symbol          Symbol           `help:"Symbol" default:"" embed:"" prefix:"icon-"`
}

func (c *Config) KeepOk() bool {
	return ContainsStringFold(c.Status, "all") || ContainsStringFold(c.Status, "ok")
}

func (c *Config) KeepWarn() bool {
	return ContainsStringFold(c.Status, "all") || ContainsStringFold(c.Status, "warn")
}
func (c *Config) KeepKo() bool {
	return ContainsStringFold(c.Status, "all") || ContainsStringFold(c.Status, "ko")
}

func (c *Config) Validate() error {
	var errs []error
	for _, ignoreStr := range c.IgnoreRegexList {
		regex, err := regexp.Compile(ignoreStr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		c.RegexList = append(c.RegexList, regex)
	}

	_, err := StringToRune(c.Symbol.TermIcon)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid term icon (%s): %s", c.Symbol.TermIcon, err))
	}
	return errors.Join(errs...)
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

	if !config.Emoji {
		config.Symbol.Warn = ""
		config.Symbol.Ko = ""
		config.Symbol.Ok = ""
	}
	for _, dash := range config.DashboardPage {
		titles, err := GetTitleDict(dash, config.Url)
		if err != nil {
			Error(fmt.Errorf("Dashboard unavailable: %s", err), config)
			return
		}

		dashboard, err := GetDashboard(dash, titles, config.Url)
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

		content, globalState, _ := Parse(config, groups, dashboard, dash)

		PrintContent(content)
		if config.Notify {
			err = Notify(content, config)
			if err != nil {
				fmt.Println(err)
			}
		}

		if globalState == KO {

		}
	}
}

func PrintContent(content Content) {
	fmt.Println(content.String())
}

func countChar(s string, c Config) int {

	count := 0
	var (
		ok   rune
		ko   rune
		warn rune
	)
	if !c.Beat {
		return 0
	}
	if c.BeatEmoji && c.Emoji {
		r, _ := StringToRune(c.Symbol.OkBeatEmoji)
		ok = r
		r, _ = StringToRune(c.Symbol.KoBeatEmoji)
		ko = r
		r, _ = StringToRune(c.Symbol.WarnBeatEmoji)
		warn = r
		for _, r := range s {
			if r == ok || r == ko || r == warn {
				count++
			}
		}
		return count
	} else {
		t, _ := StringToRune(c.Symbol.TermIcon)
		for _, r := range s {
			if r == t {
				count++
			}
		}
		return count
	}

}

func Error(err error, config Config) {
	if config.Xbar {
		fmt.Printf("%s\n---\n", config.Symbol.Error)
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

func Notify(hb Content, c Config) error {
	if c.NotifyUrl == nil {
		return fmt.Errorf("no notify url")
	}
	err, notifier := NewNotifier(c)

	if err != nil {
		return err
	}
	notifier.Notify(hb, c)
	return nil
}

func Parse(config Config, groups []Group, dashboard HeartBeatList, dashName string) (Content, State, HeartBeatList) {

	globalState := OK

	downList := HeartBeatList{}
	length := 0
	for _, monitors := range dashboard {
		for _, monitor := range monitors {
			if (monitor.IsOK() && !config.KeepOk()) || monitor.IsWarn() && !config.KeepWarn() || monitor.IsKO() && !config.KeepKo() {
				continue
			}
			l := len(monitor.Name)
			if l > length {
				length = l
			}
		}
	}

	content := Content{}

	maxWidth := 0

	for _, group := range groups {
		monitors := dashboard[group]
		for _, monitor := range monitors {
			localStatus, _ := monitor.analyzeStatus(config.IgnoreList, config.RegexList)
			if (localStatus == KO && !config.KeepKo()) || (localStatus == Warn && !config.KeepWarn()) || (localStatus == OK && !config.KeepOk()) {
				continue
			}

			l := uniseg.GraphemeClusterCount(removeANSICodes(monitor.Beats(config)))

			if l > maxWidth {
				maxWidth = l
			}
		}
	}

	for _, group := range groups {
		monitors := dashboard[group]

		contentGroup := ParsedGroups{
			GroupName: group.Name,
		}
		sort.Slice(monitors, func(i, j int) bool {
			return monitors[i].Name < monitors[j].Name
		})

		for _, monitor := range monitors {
			var icon string
			localStatus, globalStatus := monitor.analyzeStatus(config.IgnoreList, config.RegexList)
			if (localStatus == KO && !config.KeepKo()) || (localStatus == Warn && !config.KeepWarn()) || (localStatus == OK && !config.KeepOk()) {
				continue
			}
			icon = config.Symbol.Get(localStatus)
			if localStatus == KO {
				downList[group] = append(downList[group], monitor)
			}
			nb := countChar(monitor.Beats(config), config)

			pad := maxWidth - nb
			if pad < 0 {
				pad = 0
			}
			if config.BeatEmoji && config.Emoji {
				pad *= 2
			}

			beats := fmt.Sprintf("%-*s%s ", pad, "", monitor.Beats(config))

			if config.Xbar {
				beats = fmt.Sprintf("%s | font=\"FiraCode Nerd Font\"\n", beats)
			} else {
				beats = fmt.Sprintf("%s\n", beats)
			}

			contentGroup.Monitors = append(contentGroup.Monitors, ParsedMonitor{
				State:      localStatus,
				Emoji:      icon,
				Beats:      beats,
				EmojiBeats: fmt.Sprintf("%-*s%s \n", pad, "", monitor.EmojiBeats(config)),
				Name:       monitor.GetName(length, config),
			})
			if globalState == KO {
				continue
			}

			if globalStatus == KO {
				globalState = KO
				continue
			}
			if globalStatus == Warn && globalState == OK {
				globalState = Warn
				continue
			}
		}
		if contentGroup.IsOK() {
			contentGroup.GroupName = fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[config.Color.OkBeat], contentGroup.GroupName)
		}

		if contentGroup.IsWarn() {
			contentGroup.GroupName = fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[config.Color.WarnBeat], contentGroup.GroupName)
		}

		if contentGroup.IsKO() {
			contentGroup.GroupName = fmt.Sprintf("\u001B[%dm%s\u001B[0m", colors[config.Color.KoBeat], contentGroup.GroupName)
		}
		if len(contentGroup.Monitors) > 0 {
			content.Content = append(content.Content, contentGroup)
		}
	}

	content.Header = dashName
	if config.Xbar {
		icon := config.Symbol.Get(globalState)

		content.Header = fmt.Sprintf("%s %s\n---", dashName, icon)
		content.Footer = "Refresh... | refresh=true"
	}

	//fmt.Println(content)
	return content, globalState, downList
}

type Content struct {
	Header  string
	Footer  string
	Content []ParsedGroups
}

func (c *Content) String() string {
	sb := strings.Builder{}
	if c.Header != "" {
		sb.WriteString(c.Header)
		sb.WriteString("\n")
	}
	for _, group := range c.Content {
		sb.WriteString(group.GroupName)
		sb.WriteString("\n")
		for _, monitor := range group.Monitors {
			sb.WriteString(monitor.Name)
			sb.WriteString(" ")
			sb.WriteString(monitor.Emoji)
			sb.WriteString(monitor.Beats)
		}
		sb.WriteString("\n")

	}

	if c.Footer != "" {
		sb.WriteString(c.Footer)
	}
	return strings.TrimSpace(sb.String())
}

type ParsedGroups struct {
	GroupName string
	Monitors  []ParsedMonitor
	Color     bool
}

func (group ParsedGroups) IsOK() bool {
	for _, monitor := range group.Monitors {
		if monitor.State == KO || monitor.State == Warn {
			return false
		}
	}
	return true
}

func (group ParsedGroups) IsKO() bool {
	for _, monitor := range group.Monitors {
		if monitor.State == OK || monitor.State == Warn {
			return false
		}
	}
	return true
}

func (group ParsedGroups) IsWarn() bool {
	for _, monitor := range group.Monitors {
		if monitor.State == KO || monitor.State == OK {
			return false
		}
	}
	return true
}

type ParsedMonitor struct {
	State      State
	Emoji      string
	Beats      string
	EmojiBeats string
	Name       string
}

func ContainsStringFold(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

func StringToRune(s string) (rune, error) {
	if s == "" {
		return 0, fmt.Errorf("string is empty")
	}

	runes := []rune(s)
	if len(runes) != 1 {
		return 0, fmt.Errorf("string must contain exactly one character")
	}

	return runes[0], nil
}

func removeANSICodes(input string) string {
	// Regex to match ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(input, "")
}
