package hid

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/Dowdow/deckline/control"
	"github.com/Dowdow/deckline/hid/dualshock4"
)

type HIDController struct {
	Path      string
	VendorId  uint16
	ProductId uint16
}

func (c *HIDController) Listen(events chan control.ControllerEvent) {
	// Maybe some generic data can be extracted here ?
}

type hidMonitor struct{}

func NewHIDMonitor() *hidMonitor {
	return &hidMonitor{}
}

func (m *hidMonitor) Scan() []control.Controller {
	entries, err := os.ReadDir("/sys/class/hidraw")
	if err != nil {
		return nil
	}

	controllers := make([]control.Controller, 0)

	for _, entry := range entries {
		name := entry.Name()

		controller := m.CreateController(filepath.Join("/dev", name))
		if controller != nil {
			controllers = append(controllers, controller)
		}
	}

	return controllers
}

func (m *hidMonitor) Watch(events chan control.ControllerMonitorEvent) {
	defer close(events)

	fd, err := syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_DGRAM,
		syscall.NETLINK_KOBJECT_UEVENT,
	)
	if err != nil {
		return
	}
	defer syscall.Close(fd)

	sa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: 1,
		Pid:    uint32(syscall.Getpid()),
	}

	if err := syscall.Bind(fd, sa); err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)

	for {
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			continue
		}

		parts := bytes.Split(buf[:n], []byte{0})

		env := make(map[string]string)

		for _, part := range parts {
			before, after, found := bytes.Cut(part, []byte{'='})
			if found {
				env[string(before)] = string(after)
			}
		}

		action := env["ACTION"]
		devPath := filepath.Join("/dev", env["DEVNAME"])

		if env["SUBSYSTEM"] == "hidraw" {
			switch action {
			case "add":
				controller := m.CreateController(devPath)
				if controller != nil {
					events <- control.ControllerMonitorEvent{
						Connected:  true,
						Controller: controller,
					}
				}
			case "remove":
				events <- control.ControllerMonitorEvent{
					Connected:  false,
					Controller: &HIDController{Path: devPath},
				}
			}
		}

		if env["SUBSYSTEM"] == "hidraw" {
			if env["ACTION"] == "add" || env["ACTION"] == "remove" {
				m.CreateController(filepath.Join("/dev", env["DEVNAME"]))
				continue
			}
		}
	}
}

func (m *hidMonitor) CreateController(path string) control.Controller {
	var f *os.File
	var err error

	for range 5 {
		f, err = os.OpenFile(path, os.O_RDWR, 0)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if err != nil {
		return nil
	}
	defer f.Close()

	var info struct {
		BusType uint32
		Vendor  int16
		Product int16
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		// ioctl number for HIDIOCGRAWINFO
		0x80084803,
		uintptr(unsafe.Pointer(&info)),
	)
	if errno != 0 {
		return nil
	}

	vendorId := uint16(info.Vendor)
	productId := uint16(info.Product)

	switch {
	case vendorId == 0x054c && productId == 0x09cc:
		return &dualshock4.Dualshock4{Path: path}
	default:
		return &HIDController{
			Path:      path,
			VendorId:  vendorId,
			ProductId: productId,
		}
	}
}
