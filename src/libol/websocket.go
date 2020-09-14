package libol

import (
	"crypto/tls"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"time"
)

type wsConn struct {
	*websocket.Conn
}

func (ws *wsConn) RemoteAddr() net.Addr {
	req := ws.Request()
	if req == nil {
		return ws.RemoteAddr()
	}
	addr := req.RemoteAddr
	if ret, err := net.ResolveTCPAddr("tcp", addr); err == nil {
		return ret
	}
	return nil
}

type WebCa struct {
	CaKey string
	CaCrt string
}

type WebConfig struct {
	Ca      *WebCa
	Block   kcp.BlockCrypt
	Timeout time.Duration // ns
	RdQus   int           // per frames
	WrQus   int           // per frames
}

// Server Implement

type WebServer struct {
	*SocketServerImpl
	webCfg   *WebConfig
	listener *http.Server
}

func NewWebServer(listen string, cfg *WebConfig) *WebServer {
	t := &WebServer{
		webCfg:           cfg,
		SocketServerImpl: NewSocketServer(listen),
	}
	t.WrQus = cfg.WrQus
	t.close = t.Close
	return t
}

func (t *WebServer) Listen() (err error) {
	if t.webCfg.Ca != nil {
		Info("WebServer.Listen: wss://%s", t.address)
	} else {
		Info("WebServer.Listen: ws://%s", t.address)
	}
	t.listener = &http.Server{
		Addr: t.address,
	}
	return nil
}

func (t *WebServer) Close() {
	if t.listener != nil {
		_ = t.listener.Close()
		Info("WebServer.Close: %s", t.address)
		t.listener = nil
	}
}

func (t *WebServer) Accept() {
	Debug("WebServer.Accept")

	_ = t.Listen()
	defer t.Close()
	t.listener.Handler = websocket.Handler(func(ws *websocket.Conn) {
		if !t.preAccept(ws) {
			return
		}
		defer ws.Close()
		ws.PayloadType = websocket.BinaryFrame
		wws := &wsConn{ws}
		client := NewWebClientFromConn(wws, t.webCfg)
		t.onClients <- client
		<-client.done
		Info("WebServer.Accept: %s exit", ws.RemoteAddr())
	})
	promise := Promise{
		First:  2 * time.Second,
		MinInt: 5 * time.Second,
		MaxInt: 30 * time.Second,
	}
	promise.Done(func() error {
		if t.webCfg.Ca == nil {
			if err := t.listener.ListenAndServe(); err != nil {
				Error("WebServer.Accept on %s: %s", t.address, err)
				return err
			}
		} else {
			ca := t.webCfg.Ca
			if err := t.listener.ListenAndServeTLS(ca.CaCrt, ca.CaKey); err != nil {
				Error("WebServer.Accept on %s: %s", t.address, err)
				return err
			}
		}
		return nil
	})
}

// Client Implement

type WebClient struct {
	*SocketClientImpl
	webCfg *WebConfig
	done   chan bool
	RdBuf  int // per frames
	WrBuf  int // per frames
}

func NewWebClient(addr string, cfg *WebConfig) *WebClient {
	t := &WebClient{
		webCfg: cfg,
		SocketClientImpl: NewSocketClient(addr, &StreamMessagerImpl{
			block:   cfg.Block,
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
		done: make(chan bool, 2),
	}
	return t
}

func NewWebClientFromConn(conn net.Conn, cfg *WebConfig) *WebClient {
	addr := conn.RemoteAddr().String()
	t := &WebClient{
		webCfg: cfg,
		SocketClientImpl: NewSocketClient(addr, &StreamMessagerImpl{
			block:   cfg.Block,
			timeout: cfg.Timeout,
			bufSize: cfg.RdQus * MaxFrame,
		}),
		done: make(chan bool, 2),
	}
	t.updateConn(conn)
	return t
}

func (t *WebClient) Connect() error {
	if !t.Retry() {
		return nil
	}
	var url string
	if t.webCfg.Ca != nil {
		t.out.Info("WebClient.Connect: wss://%s", t.address)
		url = "wss://" + t.address
	} else {
		t.out.Info("WebClient.Connect: ws://%s", t.address)
		url = "ws://" + t.address
	}
	config, err := websocket.NewConfig(url, url)
	if err != nil {
		return err
	}
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	t.SetConnection(conn)
	if t.listener.OnConnected != nil {
		_ = t.listener.OnConnected(t)
	}
	return nil
}

func (t *WebClient) Close() {
	t.out.Info("WebClient.Close: %v", t.IsOk())
	t.lock.Lock()
	if t.connection != nil {
		if t.status != ClTerminal {
			t.status = ClClosed
		}
		t.updateConn(nil)
		t.done <- true
		t.private = nil
		t.lock.Unlock()
		if t.listener.OnClose != nil {
			_ = t.listener.OnClose(t)
		}
		t.out.Info("WebClient.Close: %d", t.status)
	} else {
		t.lock.Unlock()
	}
}

func (t *WebClient) Terminal() {
	t.SetStatus(ClTerminal)
	t.Close()
}

func (t *WebClient) SetStatus(v SocketStatus) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != v {
		if t.listener.OnStatus != nil {
			t.listener.OnStatus(t, t.status, v)
		}
		t.status = v
	}
}
