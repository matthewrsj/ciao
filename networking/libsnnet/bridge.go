//
// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package libsnnet

import (
	"net"
	"fmt"

	"github.com/vishvananda/netlink"
	"os/exec"
	"github.com/golang/glog"
)

// NewBridge is used to initialize the bridge properties
// This has to be called prior to Create() or GetDevice()
func NewBridge(id string, mode NetworkMode) (*Bridge, error) {
	bridge := &Bridge{}
	bridge.Link = &netlink.Bridge{}
	bridge.GlobalID = id //TODO: Add other parameters
	bridge.Mode = mode
	return bridge, nil
}

// GetDevice associates the bridge with an existing bridge with that GlobalId.
// If there are multiple bridges incorrectly created with the same id, it will
// associate the bridge with the first
func (b *Bridge) GetDevice() error {

	if b.GlobalID == "" {
		return netError(b, "GetDevice: unnamed bridge")
	}

	link, err := netlink.LinkByAlias(b.GlobalID)

	if err != nil {
		return netError(b, "GetDevice: link by alias %v", err)
	}

	brl, ok := link.(*netlink.Bridge)
	if !ok {
		return netError(b, "GetDevice: incorrect interface type %v %v", b.GlobalID, link.Type())
	}

	b.Link = brl
	b.LinkName = brl.Name
	return nil
}

// Create instantiates a new bridge.
func (b *Bridge) Create() error {

	if b.GlobalID == "" {
		return netError(b, "create an unnamed bridge")
	}

	var err error

	if b.LinkName == "" {
		if b.LinkName, err = genIface(b, true); err != nil {
			return netError(b, "create %v", err)
		}

		if _, err := netlink.LinkByAlias(b.GlobalID); err == nil {
			return netError(b, "create interface exists: %v %v", b.GlobalID, b.LinkName)
		}
	}

	switch b.Mode {
	case GreTunnel:
		bridge := &netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: b.LinkName}}

		if err := netlink.LinkAdd(bridge); err != nil {
			return netError(b, "create link add %v %v", b.GlobalID, err)
		}

		link, err := netlink.LinkByName(b.LinkName)
		if err != nil {
			return netError(b, "create LinkByName %v %v", b.GlobalID, err)
		}

		brl, ok := link.(*netlink.Bridge)
		if !ok {
			return netError(b, "create incorrect interface type %v %v", b.GlobalID, link)
		}

		b.Link = brl
		if err := b.setAlias(b.GlobalID); err != nil {
			err1 := b.Destroy()
			return netError(b, "create set alias [%v] [%v]", err, err1)
		}

		return nil
		break
	case OvsGreTunnel:
		if err := createOvsBridge(b.GlobalID); err != nil {
			return err
		}
		return nil
		break
	default:
		return netError(b, "Create Bridge: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}

	return nil

}

// Destroy an existing bridge
func (b *Bridge) Destroy() error {
	switch b.Mode {
	case GreTunnel:
		if b.Link == nil || b.Link.Index == 0 {
			return netError(b, "destroy bridge unnitialized")
		}

		if err := netlink.LinkDel(b.Link); err != nil {
			return netError(b, "destroy bridge %v", err)
		}
		return nil
		break
	case OvsGreTunnel:
		glog.Info("destroying ovs bridge")
		if err := destroyOvsBridge(b.GlobalID); err != nil {
			return netError(b, "destroy bridge %v", err)
		}
		return nil
		break
	default:
		return netError(b, "Destroy bridge: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}
	return nil

}

// Enable the bridge
func (b *Bridge) Enable() error {
	switch b.Mode {
	case GreTunnel:
		if b.Link == nil || b.Link.Index == 0 {
			return netError(b, "enable bridge unnitialized")
		}

		if err := netlink.LinkSetUp(b.Link); err != nil {
			return netError(b, "enable link set up", err)
		}

		return nil
		break
	case OvsGreTunnel:
		args := []string{b.GlobalID, "up"}
		if _, err := exec.Command("ifconfig", args...).Output(); err != nil {
			return err
		}
		return nil
		break
	default:
		return netError(b, "Enable bridge: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}
	return nil

}

// Disable the bridge
func (b *Bridge) Disable() error {
	switch b.Mode {
	case GreTunnel:
		if b.Link == nil || b.Link.Index == 0 {
			return netError(b, "disable bridge unnitialized")
		}

		if err := netlink.LinkSetDown(b.Link); err != nil {
			return netError(b, "disable link set down %v", err)
		}

		return nil
	case OvsGreTunnel:
		args := []string{b.GlobalID, "down"}
		if _, err := exec.Command("ifconfig", args...).Output(); err != nil {
			return err
		}
		return nil
	default:
		return netError(b, "Disable bridge: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}
	return nil

}

// AddIP Adds an IP Address to the bridge
func (b *Bridge) AddIP(ip *net.IPNet) error {
	switch b.Mode {
	case GreTunnel:
		if b.Link == nil || b.Link.Index == 0 {
			return netError(b, "add ip bridge unnitialized")
		}

		addr := &netlink.Addr{IPNet: ip}

		if err := netlink.AddrAdd(b.Link, addr); err != nil {
			return netError(b, "assigning IP address to bridge %v %v", addr.String(), err)
		}

		return nil
		break
	case OvsGreTunnel:
		glog.Info("ifconfig ip bridge.go")
		args := []string{b.GlobalID, ip.IP.String()}
		if _, err := exec.Command("ifconfig", args...).Output(); err != nil {
			return err
		}
		return nil
		break
	default:
		return netError(b, "Assign IP: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}
	return nil

}

// DelIP Deletes an IP Address assigned to the bridge
func (b *Bridge) DelIP(ip *net.IPNet) error {
	switch b.Mode {
	case GreTunnel:
		if b.Link == nil || b.Link.Index == 0 {
			return netError(b, "del ip bridge unnitialized")
		}

		addr := &netlink.Addr{IPNet: ip}

		if err := netlink.AddrDel(b.Link, addr); err != nil {
			return netError(b, "deleting IP address from bridge %v %v", addr.String(), err)
		}

		return nil
	case OvsGreTunnel:
		glog.Info("ifconfig del bridge.go")
		args := []string{b.GlobalID, "del", ip.IP.String()}
		if _, err := exec.Command("ifconfig", args...).Output(); err != nil {
			return err
		}
		return nil
		break
	default:
		return netError(b, "Delete IP: unknown network mode %v, bridge %v", b.Mode, b.GlobalID)
	}
	return nil
}

// setAlias sets up the alias on the device
func (b *Bridge) setAlias(alias string) error {

	if b.Link == nil || b.Link.Index == 0 {
		return netError(b, "set alias bridge unnitialized")
	}

	if err := netlink.LinkSetAlias(b.Link, alias); err != nil {
		return netError(b, "setting alias on bridge %v %v", alias, err)
	}

	return nil
}
