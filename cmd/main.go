package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/Dowdow/deckline/control"
	"github.com/Dowdow/deckline/hid"
	"github.com/Dowdow/deckline/hid/dualshock4"
	"github.com/Dowdow/deckline/mixer"
	"github.com/Dowdow/deckline/ui"
)

func main() {
	audioMixer, err := mixer.NewMixer()
	if err != nil {
		log.Fatal(err)
	}
	go audioMixer.Run()

	controllerEventChan := make(chan control.ControllerEvent)
	monitorEventChan := make(chan control.ControllerMonitorEvent)

	hidMonitor := hid.NewHIDMonitor()
	for _, hidController := range hidMonitor.Scan() {
		go hidController.Listen(controllerEventChan)
	}
	go hidMonitor.Watch(monitorEventChan)
	go func() {
		for mEvent := range monitorEventChan {
			if mEvent.Connected {
				go mEvent.Controller.Listen(controllerEventChan)
			}
		}
	}()

	mapping := dualshock4.NewMapping(audioMixer.GetDeck("A"), audioMixer.GetDeck("B"))
	go mapping.Run(controllerEventChan)

	p := tea.NewProgram(ui.NewMainModel(audioMixer))
	mapping.AttachProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
