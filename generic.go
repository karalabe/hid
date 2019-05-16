// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2019 Péter Szilágyi, Guillaume Ballet. All rights reserved.

package hid

import (
	"C"
	"sync"
)
import "fmt"

type GenericEndpointDirection uint8

// List of endpoint direction types
const (
	GenericEndpointDirectionOut = 0x00
	GenericEndpointDirectionIn  = 0x80
)

// List of endpoint attributes
const (
	GenericEndpointAttributeInterrupt = 3
)

// GenericEndpoint represents a USB endpoint
type GenericEndpoint struct {
	Address    uint8
	Direction  GenericEndpointDirection
	Attributes uint8
}

type GenericDeviceInfo struct {
	Path      string // Platform-specific device path
	VendorID  uint16 // Device Vendor ID
	ProductID uint16 // Device Product ID

	Device GenericLibUsbDevice

	Interface int

	Endpoints []GenericEndpoint
}

func (gdi *GenericDeviceInfo) Type() DeviceType {
	return DeviceTypeGeneric
}

// Platform-specific device path
func (gdi *GenericDeviceInfo) GetPath() string {
	return gdi.Path
}

// IDs returns the vendor and product IDs for the device
func (gdi *GenericDeviceInfo) IDs() (uint16, uint16, int, uint16) {
	return gdi.VendorID, gdi.ProductID, gdi.Interface, 0
}

// Open tries to open the USB device represented by the current DeviceInfo
func (gdi *GenericDeviceInfo) Open() (Device, error) {
	device, err := GenericDeviceOpen(gdi.Device)
	if err != nil {
		return nil, err
	}

	newDev := &GenericDevice{
		GenericDeviceInfo: gdi,
		device:            device,
	}

	for _, endpoint := range gdi.Endpoints {
		switch {
		case endpoint.Direction == GenericEndpointDirectionOut && endpoint.Attributes == GenericEndpointAttributeInterrupt:
			newDev.WEndpoint = endpoint.Address
		case endpoint.Direction == GenericEndpointDirectionIn && endpoint.Attributes == GenericEndpointAttributeInterrupt:
			newDev.REndpoint = endpoint.Address
		}
	}

	if newDev.REndpoint == 0 || newDev.WEndpoint == 0 {
		return nil, fmt.Errorf("Missing endpoint in device %#x:%#x:%d", gdi.VendorID, gdi.ProductID, gdi.Interface)
	}

	return newDev, nil
}

// GenericDevice represents a generic USB device
type GenericDevice struct {
	*GenericDeviceInfo // Embed the infos for easier access

	REndpoint uint8
	WEndpoint uint8

	device GenericDeviceHandle
	lock   sync.Mutex
}

// Close a previously opened generic USB device
func (gd *GenericDevice) Close() error {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	if gd.device != nil {
		GenericDeviceClose(gd.device)
		gd.device = nil
	}

	return nil
}

// Write implements io.ReaderWriter
func (gd *GenericDevice) Write(b []byte) (int, error) {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	out, err := InterruptTransfer(gd.device, gd.WEndpoint, b, 0)
	return len(out), err
}

// Read implements io.ReaderWriter
func (gd *GenericDevice) Read(b []byte) (int, error) {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	out, err := InterruptTransfer(gd.device, gd.REndpoint, b, 0)
	return len(out), err
}

// Type identify the device as a HID device
func (gd *GenericDevice) Type() DeviceType {
	return gd.GenericDeviceInfo.Type()
}
