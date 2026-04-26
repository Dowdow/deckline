package main

import (
	"fmt"
	"time"

	"github.com/Dowdow/deckline/control"
	"github.com/Dowdow/deckline/hid"
	"github.com/Dowdow/deckline/hid/dualshock4"
	"github.com/Dowdow/deckline/mixer"
)

func main() {
	audioMixer := mixer.NewMixer()
	go audioMixer.Run()

	audioMixer.Load("A", "track1.mp3")
	deckA := audioMixer.GetDeck("A")

	monitorEventChan := make(chan control.ControllerMonitorEvent)
	controllerEventChan := make(chan control.ControllerEvent)

	hidMonitor := hid.NewHIDMonitor()
	hidControllers := hidMonitor.Scan()
	for _, hidController := range hidControllers {
		fmt.Printf("Nouveau contrôleur détecté sur %s\n", hidController)
		go hidController.Listen(controllerEventChan)
	}

	go hidMonitor.Watch(monitorEventChan)

	// var ratio float64 = 1.0

	for {
		select {
		case mEvent, ok := <-monitorEventChan:
			if !ok {
				return
			}
			if mEvent.Connected {
				fmt.Printf("Connecté: %s\n", mEvent.Controller)
				go mEvent.Controller.Listen(controllerEventChan)
			} else {
				fmt.Println("Déconnecté")
			}

		case cEvent := <-controllerEventChan:
			switch cEvent.ID {
			case int(dualshock4.LeftStickY):
				if cEvent.Value <= 95 || cEvent.Value >= 159 {
					min := -0.001
					max := 0.001
					delta := min + (float64(cEvent.Value)/255)*(max-min)
					speed := deckA.GetSpeed()
					fmt.Println(speed, delta)
					deckA.SetSpeed(speed + delta)
				}
			case int(dualshock4.Up):
				if cEvent.Value == 1 {
					deckA.Ask()
				}
			case int(dualshock4.Down):
				if cEvent.Value == 1 {
					deckA.TogglePlay()
				}
			case int(dualshock4.Left):
				if cEvent.Value == 1 {
					deckA.Seek(0.1)
				}
			case int(dualshock4.Right):
				if cEvent.Value == 1 {
					deckA.Seek(0.9)
				}
			}

		case <-time.After(10 * time.Second):
		}
	}

	/*
		p := tea.NewProgram(ui.NewMainModel(uiChan))
		if _, err := p.Run(); err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		close(uiChan)
	*/
}
