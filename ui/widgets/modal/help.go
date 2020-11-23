/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package modal

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
)

type Help struct {
	*cview.TextView
	visible bool
	closeCb func()

	page       int
	totalPages int
	stats      models.Stats
}

func (h *Help) SetDoneFunc(doneFunc func()) {
	h.closeCb = doneFunc
}

func (h *Help) View() cview.Primitive {
	return h
}

func (h *Help) SetVisible(visible bool) {
	h.visible = visible

}

func (h *Help) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			h.closeCb()
		} else if key == tcell.KeyLeft {
			if h.page > 0 {
				h.page -= 1
				h.setContent()
			}
		} else if key == tcell.KeyRight {
			if h.page < h.totalPages-1 {
				h.page += 1
				h.setContent()
			}
		} else {
			h.TextView.InputHandler()(event, setFocus)
		}
	}
}

func (h *Help) Focus(delegate func(p cview.Primitive)) {
	h.TextView.SetBorderColor(config.Color.BorderFocus)
	h.TextView.Focus(delegate)
}

func (h *Help) Blur() {
	h.TextView.SetBorderColor(config.Color.Border)
	h.TextView.Blur()
}

func (h *Help) GetFocusable() cview.Focusable {
	return h.TextView.GetFocusable()
}

func NewHelp(doneCb func()) *Help {
	h := &Help{TextView: cview.NewTextView()}
	h.closeCb = doneCb

	colors := config.Color.Modal
	h.SetBackgroundColor(colors.Background)
	h.SetBorder(true)
	h.SetTitle("Help")
	h.SetBorderColor(config.Color.Border)
	h.SetTitleColor(config.Color.TextSecondary)
	h.SetDynamicColors(true)
	h.SetPadding(0, 1, 2, 2)

	h.totalPages = 3
	h.setContent()
	return h
}

func (h *Help) setContent() {
	title := ""
	got := ""
	switch h.page {
	case 0:
		got = h.mainPage()
		title = "About"
	case 1:
		got = h.shortcutsPage()
		title = "Usage"
	case 2:
		got = h.statsPage()
		title = "Info"
	default:
	}

	if title != "" {
		title = "[yellow::b]" + title + "[-::-]"
	}

	if got != "" {
		h.Clear()
		text := fmt.Sprintf("< %d / %d > %s \n\n", h.page+1, h.totalPages, title)
		text += got
		h.SetText(text)
		h.ScrollToBeginning()
	}
}

func (h *Help) SetStats(stats models.Stats) {
	h.stats = stats
}

func (h *Help) mainPage() string {
	text := fmt.Sprintf("%s\n[yellow]v%s[-]\n\n", logo(), config.Version)
	text += "License: Apache-2.0, http://www.apache.org/licenses/LICENSE-2.0"

	text += "\n" + helpText()
	return text
}

func (h *Help) shortcutsPage() string {
	return `[yellow]Movement[-]:
* Up/Down: Key up / down
* VIM-like keys: 
	* Up / Down: J / K, 
	* Top / Bottom of list: g / G 
	* Page Up / Down: Ctrl+F / Ctrl+B
* Switch between panels: Tab 
* Select button or item: Enter
* Open context menu: Alt+Enter
* Close application: Ctrl-C

[yellow]Queue[-]:
* Delete song: Del
* Move up song: Ctrl-K
* Move down song: Ctrl-J
* Clear queue with 'clear'. This does not remove current song

[yellow]Mouse[-]:
You can use mouse (if enabled) to navigate in application.
* Select: Left click / double click
* Open context menu: right click

`
}

func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprint(bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%d KiB", bytes/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d MiB", bytes/1024/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%d GiB", bytes/1024/1024/1024)
	}
	return ""
}

func (h *Help) statsPage() string {
	text := "[yellow]Server info[-]\n"
	text += fmt.Sprintf("Name: %s\nVersion: %s\nIdentity: %s\n"+
		"Restart pending: %t\nShutdown pending: %t\n"+
		"Websocket enabled: %t\nRemote control enabled: %t",
		h.stats.ServerName, h.stats.ServerVersion, h.stats.ServerId,
		h.stats.ServerRestartPending, h.stats.ServerShutdownPending,
		h.stats.WebSocket, h.stats.RemoteControl)

	text += "\n\n[yellow]Configuration[-]\n"
	text += fmt.Sprintf("Log file: %s\nConfig file: %s",
		h.stats.LogFile, h.stats.ConfigFile)

	text += "\n\n[yellow]Statistics[-]\n"
	text += fmt.Sprintf("Cache items: %d\nMemory allocated: %s\nWebsocket enabled: %t",
		h.stats.CacheObjects, h.stats.HeapString(), h.stats.WebSocket)
	return text
}

func logo() string {
	text := `
   __         _  _               _  _ 
   \ \   ___ | || | _   _   ___ | |(_)
    \ \ / _ \| || || | | | / __|| || |
 /\_/ /|  __/| || || |_| || (__ | || |
 \___/  \___||_||_| \__, | \___||_||_|
                    |___/             
`
	return strings.TrimLeft(text, "\n")
}

func helpText() string {
	return `
[darkorange]Jellycli[-] is a command-line / terminal music player for Jellyfin media server.
Source code: https://github.com/tryffel/jellycli

[yellow::b]Features [-:-:-]
* View artists, songs, albums, playlists, favorite artists and albums, genres, similar albums and artists
* Queue: add songs and albums, reorder & delete songs, clear queue
* Control (and view) play state through Dbus integration
* Remote control over Jellyfin server. Currently implemented:
    * [x] Play / pause / stop
    * [x] Set volume
    * [x] Next/previous track
    * [x] Control queue
    * [ ] Seeking, see (https://github.com/tryffel/jellycli/issues/8
* Supported formats (server transcodes everything else to mp3): mp3,ogg,flac,wav
* headless mode (--no-gui)

Platforms tested:
* [x] Windows 10 (amd64)
* [x] Linux 64 bit (amd64)
* [x] Linux 32 bit (armv7 / raspi 2)
* [ ] MacOS

Jellycli (headless & Gui) should work on Windows. However, there are some limitations, 
namely poor colors and some keybindings
might not work as expected. Windows Console works better than Cmd.

On raspi 2 you need to increase audio buffer duration in config file to somewhere around 400.

[yellow::b]Configuration[-::-]

On first time application asks for Jellyfin host, username, password and default collection for music. 
All this is stored in configuration file:
* ~/.config/jellycli/jellycli.yaml 
* C:\Users\<user>\AppData\Roaming\jellycli\jellycli.yaml

See config.sample.yaml for more info and up-to-date version of config file.

Configuration file location is also visible in help page. 
You can use multiple config files by providing argument:

[#005fff]jellycli --config temp.yaml[:]

Log file is located at '/tmp/jellycli.log' or 'C:\Users\<user>\AppData\Local\Temp/jellycli.log' by default. 
This can be overridden with config file. 
At the moment jellycli does not inform user about errors but rather just silently logs them.
For development purposes you should set log-level either to debug or trace.

[yellow::b]Keybindings[-::-] are hardcoded at build time. 
They are located in file [#005fff]config/keybindings.go:73[-] in function 
[#005fff]func DefaultKeybindings()[-]

edit that function as you like. 

Press Escape to return.

`
}
