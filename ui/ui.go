/*
 * Copyright 2019 Tero Vierimaa
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

package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"sync"
	"tryffel.net/pkg/jellycli/player"
	"tryffel.net/pkg/jellycli/task"
	"tryffel.net/pkg/jellycli/ui/components"
)

type Gui struct {
	task.Task
	lock        sync.RWMutex
	gui         *gocui.Gui
	initialized bool
	artists     *components.Artists
	progress    *components.StatusBar
	stateChan   chan player.PlayingState
	actionChan  chan player.Action

	lastState *player.PlayingState
}

func NewGui() (*Gui, error) {
	ui := &Gui{}
	ui.SetLoop(ui.loop)
	var err error
	ui.gui, err = gocui.NewGui(gocui.Output256)
	if err != nil {
		return ui, fmt.Errorf("failed to initialize ui: %v", err)
	}
	ui.gui.Mouse = true
	ui.gui.Cursor = true
	ui.gui.InputEsc = true

	ui.artists = components.NewArtistsView()
	ui.progress = components.NewStatusBar(ui.playerCtrl)
	ui.initialized = true
	ui.gui.SetManagerFunc(ui.layout)
	_, err = ui.gui.SetCurrentView(ui.artists.Name())

	components := []components.Component{
		ui.artists, ui.progress,
	}

	// Quit UI on Ctrl+C
	err = ui.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})
	for _, v := range components {
		err := v.AssignKeyBindings(ui.gui)
		if err != nil {
			err = fmt.Errorf("failed to assign keybindings to '%s': %v", v.Name(), err)
		}
	}
	return ui, nil
}

func (g *Gui) AssignChannels(state chan player.PlayingState, action chan player.Action) {
	if g.stateChan == nil {
		g.stateChan = state
	}
	if g.actionChan == nil {
		g.actionChan = action
	}
}

func (g *Gui) Show() error {
	err := g.gui.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	return nil
}

func (g *Gui) Close() {
	g.gui.Close()
}

func (gui *Gui) layout(g *gocui.Gui) error {
	w, h := g.Size()
	// At first init w and h are set as 0, which leads to errors when drawing components
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}

	_, err := gui.progress.Draw(g, components.Rectangle{X0: 0, Y0: 0, X1: w, Y1: 3})
	if err == gocui.ErrUnknownView {
		err = nil
	}

	_, pY := gui.progress.Size()
	_, err = gui.artists.Draw(g, components.Rectangle{X0: 0, Y0: pY, X1: w, Y1: h})
	if err == gocui.ErrUnknownView {
		err = nil
	}
	return err
}

func (g *Gui) loop() {
	for true {
		select {
		case <-g.StopChan():
			// Stop gui updates
			break
		case state := <-g.stateChan:
			// New event on media player
			//fmt.Println(state)
			g.lock.Lock()
			g.lastState = &state
			g.lock.Unlock()
			g.gui.Update(g.updateState)
		}
	}
}

func (g *Gui) updateState(gui *gocui.Gui) error {
	g.lock.RLock()
	state := g.lastState
	g.lock.RUnlock()
	g.progress.Update(state)
	return nil
}

func (g *Gui) playerCtrl(state player.State, volume int) {
	action := player.Action{
		State:   state,
		Type:    0,
		Volume:  volume,
		Artist:  "",
		Album:   "",
		Song:    "",
		AudioId: "",
	}
	g.actionChan <- action
}