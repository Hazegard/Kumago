package main

import (
	"errors"
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"strings"
)

// Notifier holds the shoutrrr service used to send discord notification
type Notifier struct {
	Urls    []string
	senders []*router.ServiceRouter
	debug   bool
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

func (n *Notifier) Notify(heartBeat HeartBeatList) {
	for group, heartBeats := range heartBeat {
		params := &types.Params{
			"title": group.Name,
			"Color": "0xDC143C",
		}
		messages := []string{}
		for _, heartBeat := range heartBeats {
			messages = append(messages, heartBeat.Name)
		}

		for _, sender := range n.senders {
			sender.Send(strings.Join(messages, "\n"), params)
		}
	}
}
