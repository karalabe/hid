// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2019 Péter Szilágyi, Guillaume Ballet. All rights reserved.

package hid

import (
	"C"
	"sync"
)

type GenericDeviceInfo struct {
	Path      string // Platform-specific device path
	VendorID  uint16 // Device Vendor ID
	ProductID uint16 // Device Product ID

	Device GenericLibUsbDevice

	WEndpoint uint8
	REndpoint uint8
}

func (gdi *GenericDeviceInfo) Type() DeviceType {
	return DeviceTypeGeneric
}

// Platform-specific device path
func (gdi *GenericDeviceInfo) GetPath() string {
	return gdi.Path
}

// IDs returns the vendor and product IDs for the device
func (gdi *GenericDeviceInfo) IDs() (uint16, uint16) {
	return gdi.VendorID, gdi.ProductID
}

// Open tries to open the USB device represented by the current DeviceInfo
func (gdi *GenericDeviceInfo) Open() (Device, error) {
	device, err := GenericDeviceOpen(gdi.Device)
	if err != nil {
		return nil, err
	}
	return &GenericDevice{
		GenericDeviceInfo: gdi,
		device:            device,
	}, nil
}

type GenericDevice struct {
	*GenericDeviceInfo // Embed the infos for easier access

	WEndpoint uint8
	REndpoint uint8

	device GenericDeviceHandle
	lock   sync.Mutex
}

func (gd *GenericDevice) Close() error {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	if gd.device != nil {
		GenericDeviceClose(gd.device)
		gd.device = nil
	}

	return nil
}

func (gd *GenericDevice) Write(b []byte) (int, error) {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	out, err := InterruptTransfer(gd.device, gd.WEndpoint, b, 0)
	return len(out), err
}

func (gd *GenericDevice) Read(b []byte) (int, error) {
	gd.lock.Lock()
	defer gd.lock.Unlock()

	out, err := InterruptTransfer(gd.device, gd.REndpoint, b, 0)
	return len(out), err
}

// SetEndpoint sets the endpoint number for either the Read or
// Write function.
func (gd *GenericDevice) SetEndpoint(write bool, endpoint uint8) {
	if write {
		gd.WEndpoint = endpoint
	} else {
		gd.REndpoint = endpoint
	}
}

// Type identify the device as a HID device
func (gd *GenericDevice) Type() DeviceType {
	return gd.GenericDeviceInfo.Type()
}
