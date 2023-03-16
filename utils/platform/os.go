/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package platform

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var (
	errNotSupportedByWindows = errors.New("not supported by windows")
)

// ConvertError converts a platform error into a commonerrors
func ConvertError(err error) error {
	switch {
	case err == nil:
		return err
	case commonerrors.Any(err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported):
		return err
	case IsWindows() && commonerrors.Any(err, errNotSupportedByWindows):
		return fmt.Errorf("%w: %v", commonerrors.ErrUnsupported, err.Error())
	case commonerrors.CorrespondTo(err, "not supported"):
		return fmt.Errorf("%w: %v", commonerrors.ErrUnsupported, err.Error())
	default:
		return err
		// TODO extend with more platform specific errors
	}
}

// IsWindows checks whether we are running on Windows or not.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// LineSeparator returns the line separator.
func LineSeparator() string {
	if IsWindows() {
		return "\r\n"
	}
	return UnixLineSeparator()
}

// UnixLineSeparator returns the line separator on Unix platform.
func UnixLineSeparator() string {
	return "\n"
}

// Hostname returns the hostname.
func Hostname() (string, error) {
	return os.Hostname()
}

// UpTime returns system uptime.
func UpTime() (uptime time.Duration, err error) {
	_uptime, err := host.Uptime()
	if err != nil {
		return
	}
	uptime = time.Duration(_uptime) * time.Second
	return
}

// BootTime returns system uptime.
func BootTime() (bootime time.Time, err error) {
	_bootime, err := host.BootTime()
	if err != nil {
		return
	}
	bootime = time.Unix(int64(_bootime), 0)
	return

}

// NodeName returns the system node name (equivalent to uname -n).
func NodeName() (nodename string, err error) {
	info, err := host.Info()
	if err != nil {
		return
	}
	nodename = fmt.Sprintf("%v (%v)", info.Hostname, info.HostID)
	return
}

// PlatformInformation returns the platform information (equivalent to uname -s).
func PlatformInformation() (information string, err error) {
	platform, family, version, err := host.PlatformInformation()
	if err != nil {
		return
	}
	information = fmt.Sprintf("%v (%v/%v)", platform, family, version)
	return
}

// SystemInformation returns the system information (equivalent to uname -a)
func SystemInformation() (information string, err error) {
	hostname, err := Hostname()
	if err != nil {
		return
	}
	nodename, err := NodeName()
	if err != nil {
		return
	}
	platform, err := PlatformInformation()
	if err != nil {
		return
	}
	uptime, err := UpTime()
	if err != nil {
		return
	}
	bootime, err := BootTime()
	if err != nil {
		return
	}
	information = fmt.Sprintf("Host: %v, Node: %v, Platform: %v, Up time: %v, Boot time: %v", hostname, nodename, platform, uptime, bootime)
	return
}

func Uname() (string, error) {
	return SystemInformation()
}

type RAM interface {
	// GetTotal returns total amount of RAM on this system
	GetTotal() uint64
	// GetAvailable returns RAM available for programs to allocate
	GetAvailable() uint64
	// GetUsed returns RAM used by programs
	GetUsed() uint64
	// GetUsedPercent returns Percentage of RAM used by programs
	GetUsedPercent() float64
	// GetFree returns kernel's notion of free memory
	GetFree() uint64
}

type VirtualMemory struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Free        uint64
}

func (m *VirtualMemory) GetTotal() uint64        { return m.Total }
func (m *VirtualMemory) GetAvailable() uint64    { return m.Available }
func (m *VirtualMemory) GetUsed() uint64         { return m.Used }
func (m *VirtualMemory) GetUsedPercent() float64 { return m.UsedPercent }
func (m *VirtualMemory) GetFree() uint64         { return m.Free }

func GetRAM() (ram RAM, err error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	ram = &VirtualMemory{
		Total:       vm.Total,
		Available:   vm.Available,
		Used:        vm.Used,
		UsedPercent: vm.UsedPercent,
		Free:        vm.Free,
	}
	return
}
