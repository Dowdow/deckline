package control

type ControllerMonitor interface {
	Scan() []Controller
	Watch(events chan ControllerMonitorEvent)
}

type ControllerMonitorEvent struct {
	Connected  bool
	Controller Controller
}

type Controller interface {
	Listen(events chan ControllerEvent)
}

type ControllerEventType int

const (
	EventTypeButton ControllerEventType = iota
	EventTypeAnalog
)

type ControllerEvent struct {
	Type  ControllerEventType
	ID    int
	Value int16
}
