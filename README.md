# Kumago

> Simple golang tool used to fetch the current state of a dashboard

It relies on dashboard pages to fetch the current information.

No credentials needed, just setting up an uptime-kuma dashboard with the monitors

It will parse the dashboard and use the public API to fetch the information needed to rebuild locally the informations.

![img_1.png](img/img_1.png)

## xbar

It supports xbar:
![img_2.png](img/img_2.png)

![img.png](img/img.png)

## Discord

As well as discord notifications

![discord.png](img/discord.png)

> Note: the goal of the discord notification is to get a synthesis of the down monitors
>
> Given that my notifications are sometimes spammed, it can get challenging to track which monitors are down with the discord notifications.
> So I configured `kumago` in a daily cron that gives me a daily synthesis of the down monitors

> Note: Other notification services might work as it uses [shoutrrr](https://github.com/containrrr/shoutrrr) under the hood

```shell
Usage: kumago [<dashboard-page> ...] [flags]

Arguments:
  [<dashboard-page> ...]    Dashboard pages to parse

Flags:
  -h, --help                           Show context-sensitive help.
      --status=KO,Warn,...             Status to display (OK,KO,Warn) ($KUMAGO_STATUS)
      --xbar                           Enable Xbar mode ($KUMAGO_XBAR)
      --notify                         Send notification ($KUMAGO_NOTIFY)
  -u, --url=                           Kuma URL ($KUMAGO_URL)
  -i, --ignore=IGNORE,...              List of ignored monitor (prefix with "re:" to match using regexes) ($KUMAGO_IGNORE)
  -I, --onlylast=ONLYLAST,...          List of monitor that must be analyzed based on the last status only (prefix with "re:" to match using regexes) ($KUMAGO_ONLYLAST)
      --notify-url=,...                Notification URL ($KUMAGO_NOTIFY_URL)
      --[no-]beat                      Show/hide heartbeat ($KUMAGO_BEAT)
      --beat-emoji                     Use emoji in beats ($KUMAGO_BEAT_EMOJI)
      --[no-]emoji                     Show synthesis emoji ($KUMAGO_EMOJI)
      --color-warn-beat="yellow"       Terminal color used to display a warn beat (ANSI color name) ($KUMAGO_COLOR_WARN_BEAT)
      --color-ok-beat="green"          Terminal color used to display an OK beat (ANSI color name) ($KUMAGO_COLOR_OK_BEAT)
      --color-ko-beat="red"            Terminal color used to display a KO beat (ANSI color name) ($KUMAGO_COLOR_KO_BEAT)
      --icon-term="█"                  Symbol used to display a beat ($KUMAGO_ICON_TERM_ICON)
      --icon-warn="🤔"                  Emoji used to indicate a warning state ($KUMAGO_ICON_WARN)
      --icon-ok="👌"                    Emoji used to indicate an OK state ($KUMAGO_ICON_OK)
      --icon-ko="🔥"                    Emoji used to indicate a KO state ($KUMAGO_ICON_KO)
      --icon-error="🏩"                 Emoji used to indicate an error state ($KUMAGO_ICON_ERROR)
      --icon-warn-beat-emoji="🟧"       Emoji used to display a warn beat ($KUMAGO_ICON_WARN_BEAT_EMOJI)
      --icon-ok-beat-emoji="🟩"         Emoji used to display an OK beat ($KUMAGO_ICON_OK_BEAT_EMOJI)
      --icon-ko-beat-emoji="🟥"         Emoji used to display a KO beat ($KUMAGO_ICON_KO_BEAT_EMOJI)
```


## Ignore and OnlyLast lists

### Ignored

Ignoring a monitor will have two impacts:
- The Monitor will never impact the global state (ie: it will always be considered as OK)
- The local state is either OK if no issues occurs, or Warn if the monitor is KO or Warn

### OnlyLast list

Adding a monitor in the OnlyLast list will have the following impacts:
- The global state of the monitor will only reflect the last status of the monitor
- The local state will still be computed according using all the states provided by uptime kuma
