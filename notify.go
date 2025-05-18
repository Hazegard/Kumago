package main

import (
	"errors"
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"strings"
	"time"
)

const (
	red    = "0xDC143C"
	green  = "0x228B22"
	yellow = "0xFFD700"
)

// Notifier holds the shoutrrr service used to send discord notification
type Notifier struct {
	Urls    []string
	senders []*router.ServiceRouter
	debug   bool
}

func NewColoredStringBuilder() ColoredStringBuilder {
	return ColoredStringBuilder{
		Builder: &strings.Builder{},
		State:   OK,
	}
}

type ColoredStringBuilder struct {
	*strings.Builder
	State State
}

func (csb *ColoredStringBuilder) Colorize(s State) {
	csb.State = csb.State.Min(s)
}

func (csb *ColoredStringBuilder) Color() string {
	switch csb.State {
	case Warn:
		return yellow
	case OK:
		return green
	case KO:
		return red
	}
	return ""
}

// NewNotifier returns the Notifier struct used to send discord notification
// it returns an error if the discord notification URL cannot be parsed by the underlying shoutrrr
func NewNotifier(c Config) (error, *Notifier) {
	var senders []*router.ServiceRouter
	var errs []error
	for _, url := range c.NotifyUrl {
		sender, err := shoutrrr.CreateSender(url)
		if err != nil {
			errs = append(errs, fmt.Errorf("error while parsing discord url (%s): %s", c.NotifyUrl, err))
			continue
		}
		senders = append(senders, sender)
	}
	if len(senders) == 0 {
		return fmt.Errorf("error, no valid webhook found"), nil
	}

	return errors.Join(errs...), &Notifier{
		Urls:    c.NotifyUrl,
		senders: senders,
	}

}

func (n *Notifier) Notify(content Content, config Config) {
	webhookLimit := 1990

	var messages []ColoredStringBuilder
	// message.WriteString("# Down status\n")
	for _, group := range content.Content {
		message := NewColoredStringBuilder()
		if (group.IsOK() && !config.KeepOk()) || (group.IsKO() && !config.KeepKo()) || (group.IsWarn() && !config.KeepWarn()) {
			continue
		}
		message.WriteString(fmt.Sprintf("\n### %s\n```ansi\n", group.GroupName))
		for _, monitor := range group.Monitors {
			if monitor.State != KO && !config.KeepOk() {
				continue
			}
			message.Colorize(monitor.State)
			if message.Len() > webhookLimit {
				message.WriteString("\n```")
				messages = append(messages, message)
				message = NewColoredStringBuilder()
				message.WriteString("```ansi\n")
			}
			message.WriteString(fmt.Sprintf("%s %s\n%s", monitor.Emoji, monitor.Name, monitor.EmojiBeats))
		}
		message.WriteString("```\n")
		messages = append(messages, message)
	}

	for _, message := range messages {
		msg := message.String()
		for _, sender := range n.senders {
			errs := sender.Send(msg, &types.Params{"Color": message.Color()})
			if len(errs) > 0 {
				for _, err := range errs {
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

}
