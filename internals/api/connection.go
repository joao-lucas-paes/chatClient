package api

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Msg struct {
	Status 	bool
	Text 		string
}

type Channel struct {
	ChannelName string
	Connection 	net.Conn
	Reader      *bufio.Reader
}

type ForceReload struct {}

func ConnectTo(addr string, portServer string) (Channel, bool) {
	addr = addr + ":"
	conn, err := net.Dial("tcp", addr + portServer)
	if err != nil {
		return Channel{}, false
	}
	_ = conn.SetDeadline(time.Now().Add(30*time.Second))
	r := bufio.NewReader(conn)
	conn.Write([]byte("hello\n"))
	port, errPort := r.ReadString(byte('\n'))

	if errPort != nil {
		return Channel{}, false
	}
	_, errPort = conn.Write([]byte("ok\n"))
	
	if errPort != nil {
		return Channel{}, false
	}

	connChannel, errNewConn := net.Dial("tcp", addr + strings.TrimSpace(port))

	if errNewConn != nil {
		return Channel{}, false
	}

	return Channel{ChannelName: "<login/>", Connection: connChannel}, true
}


func Login(user string, channelName string, chn *Channel) bool {
	chn.ChannelName = channelName

	response, err := chn.Reader.ReadString(byte('\n'))

	if err != nil {
		return false
	}
	
	_, err = chn.Connection.Write([]byte(fmt.Sprintf("%s %s\n", user, channelName)))
	if err != nil {
		return false
	}

	response, err = chn.Reader.ReadString('>')

	if err != nil {
		return false
	}

	return response == "<ok>"
}

func sendMsg(chn *Channel, msg string) bool {
	_, err := chn.Connection.Write([]byte(msg))
	return err == nil	
}

func RoutineSendMsg(chn *Channel, msgs chan Msg) {
	for {
		for msg2 := range msgs {
			sendMsg(chn, msg2.Text + "\n")
		}
	}
}

func RoutineReadMsg(chn *Channel, responseMsgs *[]Msg, m *sync.Mutex, p *tea.Program) {
	for {
		body, err := chn.Reader.ReadString(byte('\n'))
		body = body[:len(body)-1] // remove a quebra de linha
		if err != nil {
			continue
		}
		m.Lock()
		*responseMsgs = append(*responseMsgs, Msg{Text: body, Status: true})
		m.Unlock()
		p.Send(ForceReload{})
	}
}