package network

import (
	"github.com/danieldin95/openlan-go/src/libol"
	"github.com/vishvananda/netlink"
)

type LinuxBridge struct {
	address *netlink.Addr
	ifMtu   int
	name    string
	device  netlink.Link
	ctl     *libol.BrCtl
}

func NewLinuxBridge(name string, mtu int) *LinuxBridge {
	b := &LinuxBridge{
		name:  name,
		ifMtu: mtu,
		ctl:   libol.NewBrCtl(name),
	}
	Bridges.Add(b)
	return b
}

func (b *LinuxBridge) Kernel() string {
	return b.name
}

func (b *LinuxBridge) Open(addr string) {
	libol.Debug("LinuxBridge.Open: %s", b.name)

	link, _ := netlink.LinkByName(b.name)
	if link == nil {
		br := &netlink.Bridge{
			LinkAttrs: netlink.LinkAttrs{
				TxQLen: -1,
				Name:   b.name,
			},
		}
		err := netlink.LinkAdd(br)
		if err != nil {
			libol.Error("LinuxBridge.Open: %s", err)
			return
		}
		link, err = netlink.LinkByName(b.name)
		if link == nil {
			libol.Error("LinuxBridge.Open: %s", err)
			return
		}
	}
	if err := netlink.LinkSetUp(link); err != nil {
		libol.Error("LinuxBridge.Open: %s", err)
	}
	libol.Info("LinuxBridge.Open %s", b.name)
	if addr != "" {
		ipAddr, err := netlink.ParseAddr(addr)
		if err != nil {
			libol.Error("LinuxBridge.Open: ParseCIDR %s : %s", addr, err)
		}
		if err := netlink.AddrAdd(link, ipAddr); err != nil {
			libol.Error("LinuxBridge.Open: SetLinkIp %s : %s", b.name, err)
		}
		b.address = ipAddr
	}
	b.device = link
}

func (b *LinuxBridge) Close() error {
	var err error
	if b.device != nil && b.address != nil {
		if err = netlink.AddrDel(b.device, b.address); err != nil {
			libol.Error("LinuxBridge.Close: UnsetLinkIp %s : %s", b.name, err)
		}
	}
	return err
}

func (b *LinuxBridge) AddSlave(name string) error {
	if err := b.ctl.AddPort(name); err != nil {
		libol.Error("LinuxBridge.AddSlave: %s %s", name, b.name)
		return err
	}
	libol.Info("LinuxBridge.AddSlave: %s %s", name, b.name)
	return nil
}

func (b *LinuxBridge) DelSlave(name string) error {
	if err := b.ctl.DelPort(name); err != nil {
		libol.Error("LinuxBridge.DelSlave: %s %s", name, b.name)
		return err
	}
	libol.Info("LinuxBridge.DelSlave: %s %s", name, b.name)
	return nil
}

func (b *LinuxBridge) Type() string {
	return "linux"
}

func (b *LinuxBridge) Name() string {
	return b.name
}

func (b *LinuxBridge) SetName(value string) {
	b.name = value
}

func (b *LinuxBridge) Input(m *Framer) error {
	return nil
}

func (b *LinuxBridge) SetTimeout(value int) {
	//TODO
}

func (b *LinuxBridge) Mtu() int {
	return b.ifMtu
}

func (b *LinuxBridge) Stp(enable bool) error {
	if err := b.ctl.Stp(enable); err != nil {
		return err
	}
	return nil
}

func (b *LinuxBridge) Delay(value int) error {
	if err := b.ctl.Delay(value); err != nil {
		return err
	}
	return nil
}
