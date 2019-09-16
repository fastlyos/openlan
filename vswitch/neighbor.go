package vswitch

import (
    "fmt"
    "net"
    "log"
    "sync"
    "time"
    "strings"
    
    "github.com/danieldin95/openlan-go/libol"
)

type Neighbor struct {
    Client *libol.TcpClient  `json:"Client"`
    HwAddr net.HardwareAddr `json:"HwAddr"`
    IpAddr net.IP `json:"IpAddr"`
    NewTime int64 `json:"NewTime"`
    HitTime int64 `json:"HitTime"`
}

func (this *Neighbor) String() string {
    return fmt.Sprintf("%s,%s,%s", this.HwAddr, this.IpAddr, this.Client)
}

func NewNeighbor(hwaddr net.HardwareAddr, ipaddr net.IP, client *libol.TcpClient) (this *Neighbor) {
    this = &Neighbor {
        HwAddr: hwaddr,
        IpAddr: ipaddr,
        Client: client,
        NewTime: time.Now().Unix(),
        HitTime: time.Now().Unix(),
    }

    return
}

func (this *Neighbor) UpTime() int64 {
    return time.Now().Unix() - this.NewTime
}

type Neighborer struct {
    lock sync.RWMutex
    neighbors map[string]*Neighbor
    verbose int
    wroker *VSwitchWroker
}

func NewNeighborer(wroker *VSwitchWroker, c *Config) (this *Neighborer) {
    this = &Neighborer {
        neighbors: make(map[string]*Neighbor, 1024*10),
        verbose: c.Verbose,
        wroker: wroker,
    }

    return
}

func (this *Neighborer) GetNeighbor(name string) *Neighbor {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if n, ok := this.neighbors[name]; ok {
        return n
    }

    return nil
}

func (this *Neighborer) ListNeighbor() chan *Neighbor {
    c := make(chan *Neighbor, 128)

    go func() {
        this.lock.RLock()
        defer this.lock.RUnlock()

        for _, u := range this.neighbors {
            c <- u
        }
        c <- nil //Finish channel by nil.
    }()

    return c
}

func (this *Neighborer) OnFrame(client *libol.TcpClient, frame *libol.Frame) error {
    if this.IsVerbose() {
        log.Printf("Debug| Neighborer.OnFrame % x.", frame.Data)
    }

    if libol.IsInst(frame.Data) {
        return nil
    }

    ethtype, ethdata := frame.EthParse()
    if ethtype != libol.ETH_P_ARP {
        if ethtype == libol.ETH_P_VLAN {
            //TODO
        }
        
        return nil
    }

    arp, err := libol.NewArpFromFrame(ethdata)
    if err != nil {
        log.Printf("Error| Neighborer.OnFrame %s.", err)
        return nil
    }

    if arp.ProCode == libol.ETH_P_IP4 {
        if arp.OpCode == libol.ARP_REQUEST ||
           arp.OpCode == libol.ARP_REPLY {
            n := NewNeighbor(net.HardwareAddr(arp.SHwAddr), net.IP(arp.SIpAddr), client)
            this.AddNeighbor(n)
        }
    }
    
    return nil
}

func (this *Neighborer) AddNeighbor(neb *Neighbor) {
    this.lock.Lock()
    defer this.lock.Unlock()
    
    if n, ok := this.neighbors[neb.HwAddr.String()]; ok {
        //TODO update.
        log.Printf("Info| Neighborer.AddNeighbor: update %s.", neb)
        n.IpAddr = neb.IpAddr
        n.Client = neb.Client
        n.HitTime = time.Now().Unix()
    } else {
        log.Printf("Info| Neighborer.AddNeighbor: new %s.", neb)
        n = neb
        this.neighbors[neb.HwAddr.String()] = n
    }

    this.PubNeighbor(neb)
}

func (this *Neighborer) DelNeighbor(hwaddr net.HardwareAddr) {
    this.lock.RLock()
    defer this.lock.RUnlock()
    
    log.Printf("Info| Neighborer.DelNeighbor %s.", hwaddr)
    if n := this.neighbors[hwaddr.String()]; n != nil {
        delete(this.neighbors, hwaddr.String())
    }
}

func (this *Neighborer) OnClientClose(client *libol.TcpClient) {
    //TODO
    log.Printf("Info| Neighborer.OnClientClose %s.", client)
}

func (this *Neighborer) IsVerbose() bool {
    return this.verbose != 0
}

func (this *Neighborer) PubNeighbor(neb *Neighbor) {
    key := fmt.Sprintf("neighbor:%s", strings.Replace(neb.HwAddr.String(), ":", "-", -1))
    value := map[string]interface{} {
        "hwaddr": neb.HwAddr.String(),
        "ipaddr": neb.IpAddr.String(),
        "remote": neb.Client.String(),
        "newtime": neb.NewTime,
        "hittime": neb.HitTime,
    }
    
    if _, err := this.wroker.Redis.Client.HMSet(key, value).Result(); err != nil {
        log.Printf("Error| Neighborer.PubNeighbor hset %s", err)
    }
}