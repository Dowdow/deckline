package dualshock4

// Inspired by https://github.com/kpeu3i/gods4

import (
	"encoding/binary"
	"math"
	"os"

	"github.com/Dowdow/deckline/control"
)

type ID int

const (
	Cross ID = iota
	Circle
	Square
	Triangle
	L1
	L2
	L3
	R1
	R2
	R3
	Up
	Down
	Left
	Right
	Share
	Options
	PlayStation
	LeftStickX
	LeftStickY
	RightStickX
	RightStickY
	AccelerometerX
	AccelerometerY
	AccelerometerZ
	GyroscopeRoll
	GyroscopeYaw
	GyroscopePitch
	TouchpadButton
	TouchpadFinger0
	TouchpadFinger0X
	TouchpadFinger0Y
	TouchpadFinger1
	TouchpadFinger1X
	TouchpadFinger1Y
	BatteryCapacity
	BatteryCharging
	BatteryCable
)

type Dualshock4 struct {
	Path string
}

func (d *Dualshock4) Listen(events chan control.ControllerEvent) {
	f, err := os.Open(d.Path)
	if err != nil {
		return
	}
	defer f.Close()

	buf := make([]byte, 64)

	previous := newState(buf, 0, nil)
	for {
		_, err := f.Read(buf)
		if err != nil {
			return
		}

		current := newState(buf, 0, previous)

		// Buttons Left
		if current.up != previous.up {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Up),
				Value: int16(current.up),
			}
		}
		if current.down != previous.down {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Down),
				Value: int16(current.down),
			}
		}
		if current.left != previous.left {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Left),
				Value: int16(current.left),
			}
		}
		if current.right != previous.right {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Right),
				Value: int16(current.right),
			}
		}

		// Buttons right
		if current.cross != previous.cross {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Cross),
				Value: int16(current.cross),
			}
		}
		if current.circle != previous.circle {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Circle),
				Value: int16(current.circle),
			}
		}
		if current.square != previous.square {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Square),
				Value: int16(current.square),
			}
		}
		if current.triangle != previous.triangle {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Triangle),
				Value: int16(current.triangle),
			}
		}

		// Triggers left
		if current.l1 != previous.l1 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(L1),
				Value: int16(current.l1),
			}
		}
		if current.l2 != previous.l2 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(L2),
				Value: int16(current.l2),
			}
		}
		if current.l3 != previous.l3 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(L3),
				Value: int16(current.l3),
			}
		}

		// Triggers right
		if current.r1 != previous.r1 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(R1),
				Value: int16(current.r1),
			}
		}
		if current.r2 != previous.r2 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(R2),
				Value: int16(current.r2),
			}
		}
		if current.r3 != previous.r3 {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(R3),
				Value: int16(current.r3),
			}
		}

		// Special buttons
		if current.share != previous.share {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Share),
				Value: int16(current.share),
			}
		}
		if current.options != previous.options {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(Options),
				Value: int16(current.options),
			}
		}
		if current.ps != previous.ps {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(PlayStation),
				Value: int16(current.ps),
			}
		}

		// Sticks
		if current.leftStick.X != previous.leftStick.X {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(LeftStickX),
				Value: int16(current.leftStick.X),
			}
		}
		if current.leftStick.Y != previous.leftStick.Y {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(LeftStickY),
				Value: int16(current.leftStick.Y),
			}
		}
		if current.rightStick.X != previous.rightStick.X {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(RightStickX),
				Value: int16(current.rightStick.X),
			}
		}
		if current.rightStick.Y != previous.rightStick.Y {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(RightStickY),
				Value: int16(current.rightStick.Y),
			}
		}

		// Accelerometer
		if current.accelerometer.X != previous.accelerometer.X {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(AccelerometerX),
				Value: int16(current.accelerometer.X),
			}
		}
		if current.accelerometer.Y != previous.accelerometer.Y {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(AccelerometerY),
				Value: int16(current.accelerometer.Y),
			}
		}
		if current.accelerometer.Z != previous.accelerometer.Z {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(AccelerometerZ),
				Value: int16(current.accelerometer.Z),
			}
		}

		// Gyroscope
		if current.gyroscope.Roll != previous.gyroscope.Roll {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(GyroscopeRoll),
				Value: int16(current.gyroscope.Roll),
			}
		}
		if current.gyroscope.Yaw != previous.gyroscope.Yaw {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(GyroscopeYaw),
				Value: int16(current.gyroscope.Yaw),
			}
		}
		if current.gyroscope.Pitch != previous.gyroscope.Pitch {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(GyroscopePitch),
				Value: int16(current.gyroscope.Pitch),
			}
		}

		// Touchpad
		if current.touchpad.Press != previous.touchpad.Press {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(TouchpadButton),
				Value: int16(current.touchpad.Press),
			}
		}
		if len(current.touchpad.Swipe) > 0 && len(previous.touchpad.Swipe) > 0 {
			if current.touchpad.Swipe[0].IsActive != previous.touchpad.Swipe[0].IsActive {
				events <- control.ControllerEvent{
					Type:  control.EventTypeButton,
					ID:    int(TouchpadFinger0),
					Value: int16(current.touchpad.Swipe[0].IsActive),
				}
			}
			if current.touchpad.Swipe[0].X != previous.touchpad.Swipe[0].X {
				events <- control.ControllerEvent{
					Type:  control.EventTypeAnalog,
					ID:    int(TouchpadFinger0X),
					Value: int16(current.touchpad.Swipe[0].X),
				}
			}
			if current.touchpad.Swipe[0].Y != previous.touchpad.Swipe[0].Y {
				events <- control.ControllerEvent{
					Type:  control.EventTypeAnalog,
					ID:    int(TouchpadFinger0Y),
					Value: int16(current.touchpad.Swipe[0].Y),
				}
			}
		}
		if len(current.touchpad.Swipe) > 1 && len(previous.touchpad.Swipe) > 1 {
			if current.touchpad.Swipe[1].IsActive != previous.touchpad.Swipe[1].IsActive {
				events <- control.ControllerEvent{
					Type:  control.EventTypeButton,
					ID:    int(TouchpadFinger1),
					Value: int16(current.touchpad.Swipe[1].IsActive),
				}
			}
			if current.touchpad.Swipe[1].X != previous.touchpad.Swipe[1].X {
				events <- control.ControllerEvent{
					Type:  control.EventTypeAnalog,
					ID:    int(TouchpadFinger1X),
					Value: int16(current.touchpad.Swipe[1].X),
				}
			}
			if current.touchpad.Swipe[1].Y != previous.touchpad.Swipe[1].Y {
				events <- control.ControllerEvent{
					Type:  control.EventTypeAnalog,
					ID:    int(TouchpadFinger1Y),
					Value: int16(current.touchpad.Swipe[1].Y),
				}
			}
		}

		// Battery
		if current.battery.Capacity != previous.battery.Capacity {
			events <- control.ControllerEvent{
				Type:  control.EventTypeAnalog,
				ID:    int(BatteryCapacity),
				Value: int16(current.battery.Capacity),
			}
		}
		if current.battery.IsCharging != previous.battery.IsCharging {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(BatteryCharging),
				Value: int16(current.battery.IsCharging),
			}
		}
		if current.battery.IsCableConnected != previous.battery.IsCableConnected {
			events <- control.ControllerEvent{
				Type:  control.EventTypeButton,
				ID:    int(BatteryCable),
				Value: int16(current.battery.IsCableConnected),
			}
		}

		previous = current
	}
}

const analogSticksSmoothing = 4

type state struct {
	up            uint8
	down          uint8
	left          uint8
	right         uint8
	cross         uint8
	circle        uint8
	square        uint8
	triangle      uint8
	l1            uint8
	l2            byte
	l3            uint8
	r1            uint8
	r2            byte
	r3            uint8
	share         uint8
	options       uint8
	ps            uint8
	leftStick     Stick
	rightStick    Stick
	accelerometer Accelerometer
	gyroscope     Gyroscope
	touchpad      Touchpad
	battery       Battery
}

type Stick struct {
	X byte
	Y byte
}

type Accelerometer struct {
	X int16
	Y int16
	Z int16
}

type Gyroscope struct {
	Roll  int16
	Yaw   int16
	Pitch int16
}

type Touchpad struct {
	Press uint8
	Swipe []Touch
}

type Touch struct {
	IsActive uint8
	X        uint16
	Y        uint16
}

type Battery struct {
	Capacity         byte
	IsCharging       uint8
	IsCableConnected uint8
}

func newState(bytes []byte, offset uint, previous *state) *state {
	s := &state{
		up:            buttonUpState(bytes, offset),
		down:          buttonDownState(bytes, offset),
		left:          buttonLeftState(bytes, offset),
		right:         buttonRightState(bytes, offset),
		cross:         buttonCrossState(bytes, offset),
		circle:        buttonCircleState(bytes, offset),
		square:        buttonSquareState(bytes, offset),
		triangle:      buttonTriangleState(bytes, offset),
		l1:            buttonL1State(bytes, offset),
		l2:            buttonL2State(bytes, offset),
		l3:            buttonL3State(bytes, offset),
		r1:            buttonR1State(bytes, offset),
		r2:            buttonR2State(bytes, offset),
		r3:            buttonR3State(bytes, offset),
		share:         buttonShareState(bytes, offset),
		options:       buttonOptionsState(bytes, offset),
		ps:            buttonPSState(bytes, offset),
		leftStick:     buttonLeftStickState(bytes, offset, previous),
		rightStick:    buttonRightStickState(bytes, offset, previous),
		accelerometer: accelerometerState(bytes, offset),
		gyroscope:     gyroscopeState(bytes, offset),
		touchpad:      touchpadState(bytes, offset),
		battery:       batteryState(bytes, offset),
	}

	return s
}

func buttonUpState(bytes []byte, offset uint) uint8 {
	v := bytes[5+offset] & 15
	if v == 0 || v == 1 || v == 7 {
		return 1
	}
	return 0
}

func buttonDownState(bytes []byte, offset uint) uint8 {
	v := bytes[5+offset] & 15
	if v == 3 || v == 4 || v == 5 {
		return 1
	}
	return 0
}

func buttonLeftState(bytes []byte, offset uint) uint8 {
	v := bytes[5+offset] & 15
	if v == 5 || v == 6 || v == 7 {
		return 1
	}
	return 0
}

func buttonRightState(bytes []byte, offset uint) uint8 {
	v := bytes[5+offset] & 15
	if v == 1 || v == 2 || v == 3 {
		return 1
	}
	return 0
}

func buttonCrossState(bytes []byte, offset uint) uint8 {
	if bytes[5+offset]&32 != 0 {
		return 1
	}
	return 0
}

func buttonCircleState(bytes []byte, offset uint) uint8 {
	if bytes[5+offset]&64 != 0 {
		return 1
	}
	return 0
}

func buttonSquareState(bytes []byte, offset uint) uint8 {
	if bytes[5+offset]&16 != 0 {
		return 1
	}
	return 0
}

func buttonTriangleState(bytes []byte, offset uint) uint8 {
	if bytes[5+offset]&128 != 0 {
		return 1
	}
	return 0
}

func buttonL1State(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&1 != 0 {
		return 1
	}
	return 0
}

func buttonL2State(bytes []byte, offset uint) byte {
	if bytes[6+offset]&4 != 0 {
		return bytes[8+offset]
	}

	return 0
}

func buttonL3State(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&64 != 0 {
		return 1
	}
	return 0
}

func buttonR1State(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&2 != 0 {
		return 1
	}
	return 0
}

func buttonR2State(bytes []byte, offset uint) byte {
	if bytes[6+offset]&8 != 0 {
		return bytes[9+offset]
	}
	return 0
}

func buttonR3State(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&128 != 0 {
		return 1
	}
	return 0
}

func buttonShareState(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&16 != 0 {
		return 1
	}
	return 0
}

func buttonOptionsState(bytes []byte, offset uint) uint8 {
	if bytes[6+offset]&32 != 0 {
		return 1
	}
	return 0
}

func buttonPSState(bytes []byte, offset uint) uint8 {
	if bytes[7+offset]&1 != 0 {
		return 1
	}
	return 0
}

func buttonLeftStickState(bytes []byte, offset uint, previous *state) Stick {
	var prevX, prevY byte

	if previous == nil {
		prevX, prevY = bytes[1+offset], bytes[2+offset]
	} else {
		prevX, prevY = previous.leftStick.X, previous.leftStick.Y
	}

	if math.Abs(float64(bytes[1+offset])-float64(prevX)) >= float64(analogSticksSmoothing) ||
		math.Abs(float64(bytes[2+offset])-float64(prevY)) >= float64(analogSticksSmoothing) {
		return Stick{X: bytes[1+offset], Y: bytes[2+offset]}
	}

	return Stick{X: prevX, Y: prevY}
}

func buttonRightStickState(bytes []byte, offset uint, previous *state) Stick {
	var prevX, prevY byte

	if previous == nil {
		prevX, prevY = bytes[3+offset], bytes[4+offset]
	} else {
		prevX, prevY = previous.rightStick.X, previous.rightStick.Y
	}

	if math.Abs(float64(bytes[3+offset])-float64(prevX)) >= float64(analogSticksSmoothing) ||
		math.Abs(float64(bytes[4+offset])-float64(prevY)) >= float64(analogSticksSmoothing) {
		return Stick{X: bytes[3+offset], Y: bytes[4+offset]}
	}

	return Stick{X: prevX, Y: prevY}
}

func accelerometerState(bytes []byte, offset uint) Accelerometer {
	a := Accelerometer{
		X: int16(binary.LittleEndian.Uint16(bytes[13+offset:])),
		Y: -int16(binary.LittleEndian.Uint16(bytes[15+offset:])),
		Z: -int16(binary.LittleEndian.Uint16(bytes[17+offset:])),
	}

	return a
}

func gyroscopeState(bytes []byte, offset uint) Gyroscope {
	g := Gyroscope{
		Roll:  -int16(binary.LittleEndian.Uint16(bytes[19+offset:])),
		Yaw:   int16(binary.LittleEndian.Uint16(bytes[21+offset:])),
		Pitch: int16(binary.LittleEndian.Uint16(bytes[23+offset:])),
	}

	return g
}

func touchpadState(bytes []byte, offset uint) Touchpad {
	var (
		touches     []Touch
		touchOffset uint
	)

	for range 2 {
		base := 35 + touchOffset + offset

		var active = uint8(0)
		if (bytes[base] >> 7) == 0 {
			active = 1
		}

		touch := Touch{
			IsActive: active,
			X:        uint16(bytes[base+2]&0x0F)<<8 | uint16(bytes[base+1]),
			Y:        uint16(bytes[base+3])<<4 | uint16((bytes[base+2]&0xF0)>>4),
		}

		touches = append(touches, touch)
		touchOffset += 4
	}

	var press = uint8(0)
	if bytes[7+offset]&2 != 0 {
		press = 1
	}

	t := Touchpad{
		Press: press,
		Swipe: touches,
	}

	return t
}

func batteryState(bytes []byte, offset uint) Battery {
	var (
		charging    uint8
		maxCapacity byte
	)

	capacity := bytes[30+offset] & 0x0F
	isCableConnected := ((bytes[30+offset] >> 4) & 0x01) == 1

	if !isCableConnected || capacity > 10 {
		charging = 0
	} else {
		charging = 1
	}

	if isCableConnected {
		maxCapacity = 10
	} else {
		maxCapacity = 9
	}

	if capacity > maxCapacity {
		capacity = maxCapacity
	}

	capacity = byte(float64(float64(capacity) / float64(maxCapacity) * 100))

	var connected = uint8(0)
	if isCableConnected {
		connected = 1
	}

	battery := Battery{
		Capacity:         capacity,
		IsCharging:       charging,
		IsCableConnected: connected,
	}

	return battery
}
