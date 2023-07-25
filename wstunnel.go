package shieldoo_lighthouse

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK!")
	log.Debug("Endpoint Hit: /api/health")
}

func WSTunnelRun() error {
	myRouter := mux.NewRouter()
	//myRouter.HandleFunc("/wstunnel/udp/{upn}/{accessid}", wstunnelHandler)
	myRouter.HandleFunc("/wstunnel/udp/{upn}/{accessid}", wstunnelHandler)
	myRouter.HandleFunc("/api/health", homePage)
	log.Info("wstunnel start listening on: ", myconfig.WebSocketPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", myconfig.WebSocketPort), myRouter)
	return err
}

func wstunnelHandler(w http.ResponseWriter, r *http.Request) {
	session := WSSession{}
	vars := mux.Vars(r)
	log.Debug("wstunnel: upn: ", vars["upn"])
	log.Debug("wstunnel: accessid: ", vars["accessid"])

	var err error
	session.WSConn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("wstunnel error:", err)
		return
	}
	defer func() {
		if session.WSConn != nil {
			session.WSConn.Close()
		}
		if session.Conn != nil {
			session.Conn.Close()
		}
	}()
	go session.udpRun("127.0.0.1", myconfig.UdpPort)
	for {
		mt, message, err := session.WSConn.ReadMessage()
		if err != nil {
			// client disconnected
			log.Info("wstunnel read error - connection closed:", err)
			break
		}
		//log.Debug("wstunnel rec: ", len(message))
		if mt == websocket.BinaryMessage {
			go session.udpWrite(message)
		}
	}
}

type WSSession struct {
	RemoteAddr string
	RemotePort int
	LocalAddr  string
	Conn       *net.UDPConn
	WSConn     *websocket.Conn
	udplock    sync.Mutex
}

func (t *WSSession) udpWrite(buf []byte) error {
	if t.Conn == nil {
		return errors.New("UDP connection not initialized")
	}
	//log.Debug("wstunnel send: ", len(buf))
	_, err := t.Conn.Write(buf)
	return err
}

func (t *WSSession) udpReceive(buf []byte) {
	t.udplock.Lock()
	defer t.udplock.Unlock()
	if t.WSConn != nil {
		err := t.WSConn.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			log.Info("wstunnel write error:", err)
			t.WSConn.Close()
			t.Conn.Close()
		}
	}
}

func (t *WSSession) udpRun(RemoteAddr string, RemotePort int) error {
	t.RemoteAddr = RemoteAddr
	t.RemotePort = RemotePort
	t.LocalAddr = "127.0.0.1"
	server_addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", t.RemoteAddr, t.RemotePort))
	if err != nil {
		return err
	}
	log.Debug("UDP will connect to server: ", server_addr)

	log.Debug("UDP dialing connection ..")
	t.Conn, err = net.DialUDP("udp", nil, server_addr)
	if err != nil {
		return err
	}
	log.Debug("UDP connection listen: ", t.Conn.LocalAddr())
	log.Debug("UDP connection ready")

	defer t.Conn.Close()

	for {
		rxbuf := make([]byte, 2048)
		n, _, err := t.Conn.ReadFromUDP(rxbuf)
		if err != nil {
			log.Debug("UDP conn closed: ", err)
			return nil
		}
		go t.udpReceive(rxbuf[:n])
	}
}
