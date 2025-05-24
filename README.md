# Kumago

> Simple golang tool used to fetch the current state of a dashboard


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
> Given that my notifications are sometime spammed, it can get difficult to track which monitors are down with the discord notifications.
> So I configured `kumago` in a daily cron that gives me a daily synthesis of the down monitors

> Note: Other notification services might work as it uses [shoutrrr](https://github.com/containrrr/shoutrrr) under the hood

```shell
Usage: kumago [<dashboard-page> ...] [flags]

Arguments:
  [<dashboard-page> ...]    Dashboard page

Flags:
  -h, --help                                       Show context-sensitive help.
      --status=KO,Warn,...                         Show statuses ($KUMAGO_STATUS)
      --xbar                                       Show Xbar statuses ($KUMAGO_XBAR)
      --notify                                     Show notify statuses ($KUMAGO_NOTIFY)
  -u, --url=                                       Kuma URL ($KUMAGO_URL)
  -i, --ignore-list=IGNORE-LIST,...                Ignore list ($KUMAGO_IGNORE_LIST)
  -I, --ignore-regex-list=IGNORE-REGEX-LIST,...    Ignore list (regex) ($KUMAGO_IGNORE_REGEX_LIST)
      --notify-url=,...                            Discord URL ($KUMAGO_NOTIFY_URL)
      --beat-emoji                                 Use emoji ($KUMAGO_BEAT_EMOJI)
      --[no-]emoji                                 Use emoji ($KUMAGO_EMOJI)
      --icon-term-icon="█"                         Symbol used to display a beat ($KUMAGO_ICON_TERM_ICON)
      --icon-warn-beat="yellow"                    Terminal color used to display a warn beat ($KUMAGO_ICON_WARN_BEAT)
      --icon-ok-beat="green"                       Terminal color used to display an OK beat ($KUMAGO_ICON_OK_BEAT)
      --icon-ko-beat="red"                         Terminal color used to display a KO beat ($KUMAGO_ICON_KO_BEAT)
      --icon-warn="🤔"                              Symbol used to indicate a warning state ($KUMAGO_ICON_WARN)
      --icon-ok="👌"                                Symbol used to indicate an OK state ($KUMAGO_ICON_OK)
      --icon-ko="🔥"                                Symbol used to indicate a KO state ($KUMAGO_ICON_KO)
      --icon-error="🏩"                             Symbol used to indicate an error state ($KUMAGO_ICON_ERROR)
      --icon-warn-beat-emoji="🟧"                   Emoji used to display a warn beat ($KUMAGO_ICON_WARN_BEAT_EMOJI)
      --icon-ok-beat-emoji="🟩"                     Emoji used to display an OK beat ($KUMAGO_ICON_OK_BEAT_EMOJI)
      --icon-ko-beat-emoji="🟥"                     Emoji used to display a KO beat ($KUMAGO_ICON_KO_BEAT_EMOJI)
```