package main

import (
	"fmt"
	"time"

	"github.com/Dowdow/deckline/control"
	"github.com/Dowdow/deckline/hid"
)

func main() {
	// uiChan := make(chan tea.Msg)

	monitorEventChan := make(chan control.ControllerMonitorEvent)
	controllerEventChan := make(chan control.ControllerEvent)

	hidMonitor := hid.NewHIDMonitor()
	hidControllers := hidMonitor.Scan()
	for _, hidController := range hidControllers {
		fmt.Printf("Nouveau contrôleur détecté sur %s\n", hidController)
		go hidController.Listen(controllerEventChan)
	}

	go hidMonitor.Watch(monitorEventChan)

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
			if cEvent.Type == control.EventTypeButton {
				fmt.Printf("Input reçu du contrôleur : %v %v %v\n", cEvent.ID, cEvent.Type, cEvent.Value)
			}

		case <-time.After(10 * time.Second):
		}
	}

	// mixer := mixer.NewMixer(uiChan)
	// go mixer.Run()

	/*
		p := tea.NewProgram(ui.NewMainModel(uiChan))
		if _, err := p.Run(); err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		close(uiChan)
	*/
}
