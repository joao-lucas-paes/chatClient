package api

import (
	"bufio"
	"net"
	"time"
	"fmt"
)

type Msg struct {
	status 	bool
	body 		string
}

type Channel struct {
	channelName string
	connection 	net.Conn
	reader      bufio.Reader
}

func ConnectTo(addr string, portServer string) (Channel, bool) {
	conn, err := net.Dial("tcp", addr + portServer)
	if err != nil {
		return Channel{}, false
	}
	_ = conn.SetDeadline(time.Now().Add(30*time.Second))
	r := bufio.NewReader(conn)
	conn.Write([]byte("hello"))
	port, errPort := r.ReadString(byte('\n'))

	if errPort != nil {
		return Channel{}, false
	}
	_, errPort = conn.Write([]byte("ok"))
	
	if errPort != nil {
		return Channel{}, false
	}

	connChannel, errNewConn := net.Dial("tcp", addr + port)

	if errNewConn != nil {
		return Channel{}, false
	}

	return Channel{channelName: "<login/>", connection: connChannel}, true
}

func Login(user string, channelName string, chn *Channel) bool {
	chn.channelName = channelName

	_, err := chn.connection.Write([]byte(fmt.Sprintf("%s %s", user, channelName)))
	if err != nil {
		return false
	}

	response, err2 := chn.reader.ReadString(byte('\n'))

	if err2 != nil {
		return false
	}

	if response == "ok" {
		return true
	}

	return false
}

func sendMsg(chn *Channel, msg string) bool {
	_, err := chn.connection.Write([]byte(msg))
	if err != nil {
		return false
	}

	response, _ := chn.reader.ReadString(byte('\n'))

	return response == "<ok>Message sent</ok>"
}

func RoutineSendMsg(chn *Channel, msgs chan Msg, responseMsgs chan Msg) {
	for {
		for msg2 := range msgs {
			msg2.status = sendMsg(chn, msg2.body)
			responseMsgs <- msg2
		}
	}
}

func RoutineReadMsg(chn *Channel, responseMsgs chan Msg) {
	for {
		body, err := chn.reader.ReadString(byte('\n'))
		if err != nil {
			responseMsgs <- Msg{body:"", status:false}
			continue
		}
		responseMsgs <- Msg{body:body, status:true}
	}
}