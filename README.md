
简体中文 | [English](./README_EN.md)

# 概述 
[![Build Status](https://travis-ci.org/danieldin95/openlan-go.svg?branch=master)](https://travis-ci.org/danieldin95/openlan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/danieldin95/openlan-go)](https://goreportcard.com/report/lightstar-dev/openlan-go)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

OpenLAN提供一种局域网数据报文在广域网的传输实现，并能够建立多个用户空间的虚拟以太网络。

## 功能清单

* 支持基于用户名密码的接入认证；
* 支持多个网络空间划分，为不同的业务提供逻辑网络隔离；
* 支持TCP/TLS，UDP/KCP，WS/WSS等多种传输协议实现；
* 支持HTTP/HTTPS，以及SOCKS5等HTTP的正向代理技术；
* 支持基于TCP的端口转发，为防火墙下的主机提供TCP端口代理。

## 缩略语

* OLSW: OpenLAN Switch，开放局域网交换机
* OLAP: OpenLAN Access Point，开放局域网接入点
* NAT: Network Address Translation, 网络地址转换

## 分支接入

                                       OLSW(企业中心) - 10.16.1.10/24
                                                ^
                                                |
                                             Wifi(DNAT)
                                                |
                                                |
                       ----------------------Internet-------------------------
                       ^                        ^                           ^
                       |                        |                           |
                     分支1                    分支2                        分支3     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                 10.16.1.11/24             10.16.1.12/24                10.16.1.13/24
                 

## 区域互联

                   192.168.1.20/24                                 192.168.1.21/24
                         |                                                 |
                       OLAP -- 酒店 Wifi --> OLSW(南京) <--- 其他 Wifi --- OLAP
                                                |
                                                |
                                             互联网
                                                |
                                                |
                                             OLSW(上海) - 192.168.1.10/24
                                                |
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   办公 Wifi               家庭 Wifi                 酒店 Wifi     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                192.168.1.11/24           192.168.1.12/24             192.168.1.13/24

 
# 客户端用接入
客户端接入软件(OLAP)工作在用户或者企业一侧，每个OLAP通过拨入公网侧的虚拟交换机(OLSW)可以实现多点间的跨互联网互通。 

# 服务端虚拟交换
每个拨入虚拟交换机(OLSW)的客户点(OLAP)就像工作在一个物理的交换机下的主机，多个虚拟交换之间通过配置Link(服务端之间主动拨入对方)可以实现跨区域互通。

## 在Windows系统中
### 首先安装虚拟网卡驱动 tap-windows6

下载资源 [tap-windows-9](https://github.com/danieldin95/openlan-go/releases/download/v4.0/tap-windows-9.21.2.exe), 然后点击安装它。

### 然后配置接入认证
  使用notepad++新建一个文件：

    {
      "network": "default",
      "vs.addr": "www.openlan.xx",
      "vs.auth": "hi:123456",
      "if.addr": "192.168.1.20/24",
      "vs.tls": true
    }

 把它保存在文件`point.json`中，并与程序`point.windows.x86_64.exe`在同一个目录下。 

### 点击OLAP程序执行

  在打开的console终端中看到`login: success`字样，代表登录成功。如下：
  
    2020/04/15 00:19:19 INFO Config: version is 4.3.16
    2020/04/15 00:19:19 INFO Config: built on 2020-04-14T07:40:42-0400
    2020/04/15 00:19:19 INFO Config: commit at 1562b95686c195959ae4b8dca43094bf2b034710
    2020/04/15 00:19:19 INFO Point.Start Windows.
    2020/04/15 00:19:19 INFO TapWorker.Open >>>> Ethernet 2 <<<<
    2020/04/15 00:19:19 INFO TcpClient.Connect tls://www.openlan.xx:10002
    2020/04/15 00:19:19 INFO TapWorker.Read
    2020/04/15 00:19:19 INFO TapWorker.Loop
    2020/04/15 00:19:19 INFO TcpWorker.Read true
    2020/04/15 00:19:19 INFO Worker.OnSuccess
    2020/04/15 00:19:19 INFO TcpWorker.onInstruct.login: success

 *说明*
 
    vs.addr    虚拟交换的地址或者域名
    vs.auth    接入虚拟交换的认证信息，如：user:password
    if.addr    配置本地虚拟网卡地址
    vs.tls     是否启用TLS加密信道

### 添加新的Tap设备  
  打开设备管理器
  
    Control Panel\Hardware and Sound\Device Manager

  添加新的网卡

    1. 选择Network adapter；
    2. 点击Action选择Add legacy hardware；
    3. 找到Tap-Windows-9添加即可；
    4. 回到Control Panel\Network and Internet\Network Connections；
    5. 为新增的网卡重命名，如 Ethernet 3。
  
  启用新的网卡
  
    1. 使用notepad++打开point.json；
    2. 配置if.name为网卡名称，如：Ethernet 3。
    3. 运行point程序即可。

## 在Linux系统中
### 安装OLSW并运行

    [root@office ~]# wget https://github.com/danieldin95/openlan-go/releases/download/v4.3.16/openlan-vswitch-4.3.16-1.el7.x86_64.rpm
    [root@office ~]# yum install ./openlan-vswitch-4.3.16-1.el7.x86_64.rpm
    [root@office ~]# cat /etc/vswitch/vswitch.json
    {
      "crt.dir": "/var/openlan/ca",
      "log.file": "/var/log/vswitch.log",
      "http.dir": "/var/openlan/public",
      "bridge": [
        {
            "network": "default",
            "if.addr": "192.168.1.10/24"
        },
      ]
    }

 *说明*

    if.addr    配置本地网桥的地址
    bridge     配置租户的网桥，实现网络隔离
    crt.dir    存放信道加密证书的目录
    log.file   配置日志输出文件

  配置租户网络的认证信息

    [root@office ~]# cat /etc/vswitch/password/default.json
    [
      { "name": "hi", "password": "123456" },
      { "name": "hei", "password": "123456" }
    ]

  使能服务并启动

    [root@office ~]# systemctl enable vswitch
    [root@office ~]# systemctl start vswitch

### 安装OLAP并运行

    [root@home ~]# wget https://github.com/danieldin95/openlan-go/releases/download/v4.3.16/openlan-point-4.3.16-1.el7.x86_64.rpm
    [root@home ~]# yum install ./openlan-point-4.3.16-1.el7.x86_64.rpm
    [root@home ~]# cat /etc/point/point.json
    {
      "network": "default",
      "vs.tls": true,
      "vs.addr": "www.openlan.xx",
      "vs.auth": "hi:123456",
      "if.addr": "192.168.1.21/24",
      "log.file": "/var/log/point.log"
    }

  使能服务并启动
    
    [root@home ~]# systemctl enable point
    [root@home ~]# systemctl start point
  
  测试网络可达
  
    [root@home ~]# ping 192.168.1.20

## 在MacOS系统中

  在终端中运行OLAP

    admindeMac:~ admin$ cat ./point.json
    {
      "network": "default",
      "vs.addr": "www.openlan.xx",
      "vs.auth": "hi:123456",
      "vs.tls": true,
      "if.addr": "192.168.1.22/192.168.1.10"
    }
    admindeMac:~ admin$ 
    admindeMac:~ admin$ sudo ./point.darwin.x86_64

  测试网络可达

    admindeMac:~ admin$ ping 192.168.1.10


  测试与同网段其他主机的可达

    admindeMac:~ admin$ ping 192.168.1.20
    admindeMac:~ admin$ ping 192.168.1.21
 
# 获取源码

    [root@localhost ~]# go get -u -v github.com/danieldin95/openlan-go  

## 编译程序

   只编译程序
    
    [root@localhost openlan-go]# make all
   
   编译并打包
   
    [root@localhost openlan-go]# make all/pkg

   单独编译
   
    [root@localhost openlan-go]# make linux
    [root@localhost openlan-go]# make windows
    [root@localhost openlan-go]# make darwin
    

# 欢迎捐赠

欢迎使用支付宝手扫描下面的二维码，对该项目进行捐赠。

<img src="https://raw.githubusercontent.com/danieldin95/openlan-go/master/packaging/resource/donation.jpg" width="46%">

## 欢迎关注

微信号: DanielDin

邮件地址: danieldin95@163.com

