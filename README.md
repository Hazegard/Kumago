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
  [<dashboard-page> ...]    Dashboard pages to parse

Flags:
  -h, --help                                       Show context-sensitive help.
      --status=KO,Warn,...                         Status to display (OK,KO,Warn) ($KUMAGO_STATUS)
      --xbar                                       Enable Xbar mode ($KUMAGO_XBAR)
      --notify                                     Send notification ($KUMAGO_NOTIFY)
  -u, --url=                                       Kuma URL ($KUMAGO_URL)
  -i, --ignore-list=IGNORE-LIST,...                Ignore list ($KUMAGO_IGNORE_LIST)
  -I, --ignore-regex-list=IGNORE-REGEX-LIST,...    Ignore list (regex) ($KUMAGO_IGNORE_REGEX_LIST)
      --notify-url=,...                            Notification URL ($KUMAGO_NOTIFY_URL)
      --[no-]beat                                  Show/hide heartbeat ($KUMAGO_BEAT)
      --beat-emoji                                 Use emoji in beats ($KUMAGO_BEAT_EMOJI)
      --[no-]emoji                                 Show synthesis emoji ($KUMAGO_EMOJI)
      --color-warn-beat="yellow"                   Terminal color used to display a warn beat (ANSI color name) ($KUMAGO_COLOR_WARN_BEAT)
      --color-ok-beat="green"                      Terminal color used to display an OK beat (ANSI color name) ($KUMAGO_COLOR_OK_BEAT)
      --color-ko-beat="red"                        Terminal color used to display a KO beat (ANSI color name) ($KUMAGO_COLOR_KO_BEAT)
      --icon-term-icon="█"                         Symbol used to display a beat ($KUMAGO_ICON_TERM_ICON)
      --icon-warn="🤔"                              Emoji used to indicate a warning state ($KUMAGO_ICON_WARN)
      --icon-ok="👌"                                Emoji used to indicate an OK state ($KUMAGO_ICON_OK)
      --icon-ko="🔥"                                Emoji used to indicate a KO state ($KUMAGO_ICON_KO)
      --icon-error="🏩"                             Emoji used to indicate an error state ($KUMAGO_ICON_ERROR)
      --icon-warn-beat-emoji="🟧"                   Emoji used to display a warn beat ($KUMAGO_ICON_WARN_BEAT_EMOJI)
      --icon-ok-beat-emoji="🟩"                     Emoji used to display an OK beat ($KUMAGO_ICON_OK_BEAT_EMOJI)
      --icon-ko-beat-emoji="🟥"                     Emoji used to display a KO beat ($KUMAGO_ICON_KO_BEAT_EMOJI)
  ```
