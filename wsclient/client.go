package wsclient

import (
	"golang.org/x/net/websocket"
)

type WsClient interface {
	Send(msg []byte) error
	Receive() (string, error)
	Shutdown() error
}

func GetConnection(host string) (*websocket.Conn, error) {
	ws, err := websocket.Dial(getURL(host), "", getOrigin(host))
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func getURL(host string) string {
	return "wss://" + host + "/ws"
}

func getOrigin(host string) string {
	return "https://" + host
}
