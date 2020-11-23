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

package widgets

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/models"
	"tryffel.net/go/jellycli/util"
	"tryffel.net/go/twidgets"
)

// AlbumView shows user a header (album name, info, buttons) and list of songs
type PlaylistView struct {
	*twidgets.Banner
	*previous
	list        *twidgets.ScrollList
	songs       []*albumSong
	playlist    *models.Playlist
	listFocused bool

	context contextOperator

	playSongFunc  func(song *models.Song)
	playSongsFunc func(songs []*models.Song)

	description *cview.TextView
	prevBtn     *button
	playBtn     *button
	options     *dropDown
	prevFunc    func()
}

//NewAlbumView initializes new album view
func NewPlaylistView(playSong func(song *models.Song), playSongs func(songs []*models.Song),
	operator contextOperator) *PlaylistView {
	p := &PlaylistView{
		Banner:        twidgets.NewBanner(),
		previous:      &previous{},
		list:          twidgets.NewScrollList(nil),
		playSongFunc:  playSong,
		playSongsFunc: playSongs,

		description: cview.NewTextView(),
		prevBtn:     newButton("Back"),
		playBtn:     newButton("Play all"),
		context:     operator,
		options:     newDropDown("Options"),
	}

	p.list.ItemHeight = 2
	p.list.Padding = 1
	p.list.SetInputCapture(p.listHandler)
	p.list.SetBorder(true)
	p.list.SetBorderColor(config.Color.Border)
	p.list.Grid.SetColumns(1, -1)

	p.SetBorder(true)
	p.SetBorderColor(config.Color.Border)
	p.list.SetBackgroundColor(config.Color.Background)
	p.Grid.SetBackgroundColor(config.Color.Background)
	p.listFocused = false
	p.playBtn.SetSelectedFunc(p.playAll)

	p.Banner.Grid.SetRows(1, 1, 1, 1, -1)
	p.Banner.Grid.SetColumns(6, 2, 10, -1, 10, -1, 10, -3)
	p.Banner.Grid.SetMinSize(1, 6)

	p.Banner.Grid.AddItem(p.prevBtn, 0, 0, 1, 1, 1, 5, false)
	p.Banner.Grid.AddItem(p.description, 0, 2, 2, 6, 1, 10, false)
	p.Banner.Grid.AddItem(p.playBtn, 3, 2, 1, 1, 1, 10, true)
	p.Banner.Grid.AddItem(p.options, 3, 4, 1, 1, 1, 10, false)
	p.Banner.Grid.AddItem(p.list, 4, 0, 1, 8, 4, 10, false)

	btns := []*button{p.prevBtn, p.playBtn}
	selectables := []twidgets.Selectable{p.prevBtn, p.playBtn, p.options, p.list}
	for _, btn := range btns {
		btn.SetLabelColor(config.Color.ButtonLabel)
		btn.SetLabelColorFocused(config.Color.ButtonLabelSelected)
		btn.SetBackgroundColor(config.Color.ButtonBackground)
		btn.SetBackgroundColorFocused(config.Color.ButtonBackgroundSelected)
	}

	p.prevBtn.SetSelectedFunc(p.goBack)

	p.Banner.Selectable = selectables
	p.description.SetBackgroundColor(config.Color.Background)
	p.description.SetTextColor(config.Color.Text)

	if p.context != nil {
		p.list.AddContextItem("Play all from here", 0, func(index int) {
			p.playFromSelected()
		})
		p.list.AddContextItem("View album", 0, func(index int) {
			selected := p.list.GetSelectedIndex()
			song := p.songs[selected]
			p.context.ViewSongAlbum(song.song)
		})
		p.list.AddContextItem("View artist", 0, func(index int) {
			if index < len(p.songs) && p.context != nil {
				index := p.list.GetSelectedIndex()
				song := p.songs[index]
				p.context.ViewSongArtist(song.song)
			}
		})
		p.list.AddContextItem("Instant mix", 0, func(index int) {
			if index < len(p.songs) && p.context != nil {
				index := p.list.GetSelectedIndex()
				song := p.songs[index]
				p.context.InstantMix(song.song)
			}
		})

		opts := cview.NewDropDownOption("Instant Mix")
		opts.SetSelectedFunc(func(index int, option *cview.DropDownOption) {
			p.context.InstantMix(p.playlist)
		})

		optsBrowser := cview.NewDropDownOption("Instant Mix")
		optsBrowser.SetSelectedFunc(func(index int, option *cview.DropDownOption) {
			p.context.OpenInBrowser(p.playlist)
		})

		p.options.AddOptions(opts, optsBrowser)
	}

	p.list.ContextMenuList().SetBorder(true)
	p.list.ContextMenuList().SetBackgroundColor(config.Color.Background)
	p.list.ContextMenuList().SetBorderColor(config.Color.BorderFocus)
	p.list.ContextMenuList().SetSelectedBackgroundColor(config.Color.BackgroundSelected)
	p.list.ContextMenuList().SetMainTextColor(config.Color.Text)
	p.list.ContextMenuList().SetSelectedTextColor(config.Color.TextSelected)

	return p
}

func (p *PlaylistView) SetPlaylist(playlist *models.Playlist) {
	p.list.Clear()
	p.playlist = playlist
	p.songs = make([]*albumSong, len(playlist.Songs))
	items := make([]twidgets.ListItem, len(playlist.Songs))

	text := playlist.Name

	text += fmt.Sprintf("\n%d tracks  %s",
		len(playlist.Songs), util.SecToStringApproximate(playlist.Duration))

	p.description.SetText(text)

	for i, v := range playlist.Songs {
		p.songs[i] = newAlbumSong(v, false, i+1)
		p.songs[i].updateTextFunc = p.updateSongText
		items[i] = p.songs[i]
	}

	p.list.AddItems(items...)
}

func (p *PlaylistView) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()
		if p.listFocused {
			index := p.list.GetSelectedIndex()
			if index == 0 && (key == tcell.KeyUp || key == tcell.KeyCtrlK) {
				p.listFocused = false
				p.prevBtn.Focus(func(p cview.Primitive) {})
				p.list.Blur()
			} else if key == tcell.KeyEnter && event.Modifiers() == tcell.ModNone {
				p.playSong(index)
			} else {
				p.list.InputHandler()(event, setFocus)
			}
		} else {
			if key == tcell.KeyDown || key == tcell.KeyCtrlJ {
				p.listFocused = true
				p.list.Focus(func(p cview.Primitive) {})
			} else {
			}
		}
	}
}

func (p *PlaylistView) playSong(index int) {
	if p.playSongFunc != nil {
		song := p.songs[index].song
		p.playSongFunc(song)
	}
}

func (p *PlaylistView) playAll() {
	if p.playSongsFunc != nil {
		songs := make([]*models.Song, len(p.songs))
		for i, v := range p.songs {
			songs[i] = v.song
		}
		p.playSongsFunc(songs)
	}
}

func (p *PlaylistView) playFromSelected() {
	if p.playSongsFunc != nil {
		index := p.list.GetSelectedIndex()
		songs := make([]*models.Song, len(p.songs)-index)
		for i, v := range p.songs[index:] {
			songs[i] = v.song
		}
		p.playSongsFunc(songs)
	}
}

func (p *PlaylistView) listHandler(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyEnter && key.Modifiers() == tcell.ModNone {
		index := p.list.GetSelectedIndex()
		p.playSong(index)
		return nil
	}
	return key
}

func (p *PlaylistView) updateSongText(song *albumSong) {
	var name string
	if song.showDiscNum {
		name = fmt.Sprintf("%d %d. %s", song.song.DiscNumber, song.song.Index, song.song.Name)
	} else {
		name = fmt.Sprintf("%d. %s", song.index, song.song.Name)
	}

	text := song.getAlignedDuration(name)
	if len(song.song.Artists) > 0 {
		text += "\n     " + song.song.Artists[0].Name

	}
	song.SetText(text)
}
