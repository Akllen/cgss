package cg

import (
	"encoding/json"
	"errors"
	"sync"

	"cgss/ipc"
)

var _ ipc.Server = &CenterServer{} //确认实现了server接口

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

type CenterServer struct {
	servers map[string]ipc.Server
	players []*Player
	// rooms   []*Room
	mutex sync.RWMutex //读写锁
}

func NewCenterServer() *CenterServer { //初始化
	servers := make(map[string]ipc.Server)
	players := make([]*Player, 0)

	return &CenterServer{servers: servers, players: players}
}

func (server *CenterServer) addplayer(params string) error {
	player := NewPlayer()

	err := json.Unmarshal([]byte(params), &player)
	if err != nil {
		return err
	}
	server.mutex.Lock()         //写锁定
	defer server.mutex.Unlock() //解锁

	server.players = append(server.players, player)

	return nil
}

func (server *CenterServer) removeplayer(params string) error {
	server.mutex.Lock()         //写锁定
	defer server.mutex.Unlock() //解锁

	for i, v := range server.players {
		if v.Name == params {
			if len(server.players) == 1 {
				server.players = make([]*Player, 0)
			} else if i == len(server.players)-1 {
				server.players = server.players[:i]
			} else if i == 0 {
				server.players = server.players[1:]
			} else {
				server.players = append(server.players[:i-1], server.players[:i+1]...)
			}
			return nil
		}
	}
	return errors.New("找不到该玩家")
}

func (server *CenterServer) listplayer(param string) (players string, err error) {
	server.mutex.RLock() //读锁定
	defer server.mutex.RUnlock()

	if len(server.players) > 0 {
		b, _ := json.Marshal(server.players)
		players = string(b)
	} else {
		err = errors.New("没有玩家在线")
	}
	return
}
func (server *CenterServer) broadcast(params string) error {
	var message Message
	err := json.Unmarshal([]byte(params), &message)
	if err != nil {
		return err
	}
	server.mutex.Lock() //写锁
	defer server.mutex.Unlock()

	if len(server.players) > 0 {
		for _, player := range server.players {
			player.mq <- &message
		}
	} else {
		err = errors.New("没有玩家在线")
	}
	return err
}
func (server *CenterServer) Handle(method, params string) *ipc.Response {
	switch method {
	case "addplayer":
		err := server.addplayer(params)
		if err != nil {
			return &ipc.Response{Code: err.Error()}
		}
	case "removeplayer":
		err := server.removeplayer(params)
		if err != nil {
			return &ipc.Response{Code: err.Error()}
		}
	case "listplayer":
		players, err := server.listplayer(params)
		if err != nil {
			return &ipc.Response{Code: err.Error()}
		}
		return &ipc.Response{"200", players}
	case "broadcast":
		err := server.broadcast(params)
		if err != nil {
			return &ipc.Response{Code: err.Error()}
		}
		return &ipc.Response{Code: "200"}
	default:
		return &ipc.Response{Code: "404", Body: method + ":" + params}
	}
	return &ipc.Response{Code: "200"}
}

func (server *CenterServer) Name() string {
	return "CenterServer"
}
