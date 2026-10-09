package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"
	"time"

	"github.com/nobonobo/irpc-demo/rtc"
	"github.com/nobonobo/irpc-demo/services/rtcmsg"
)

type session struct{}

func (s *session) Send(msg rtcmsg.Message) error {
	log.Println(msg.Name, " -> ", msg.Text)
	if js.Global().Get("onMessageReceived").Type() == js.TypeFunction {
		js.Global().Call("onMessageReceived", msg.Name, msg.Text)
	}
	return nil
}

func runHost(this js.Value, args []js.Value) any {
	id := rtc.HostID
	if len(args) > 0 && args[0].Type() == js.TypeString && args[0].String() != "" {
		id = args[0].String()
	}
	go rtc.RunHost(&session{}, id)
	return nil
}

var (
	msgChan    chan rtcmsg.Message
	nodeCancel context.CancelFunc
)

func runNode(this js.Value, args []js.Value) any {
	if nodeCancel != nil {
		nodeCancel()
	}

	host := rtc.HostID
	id := fmt.Sprintf("node-%d", time.Now().UnixNano()%10000)
	name := "Guest"

	if len(args) == 1 {
		if args[0].Type() == js.TypeString && args[0].String() != "" {
			name = args[0].String()
		}
	} else if len(args) >= 3 {
		if args[0].Type() == js.TypeString && args[0].String() != "" {
			host = args[0].String()
		}
		if args[1].Type() == js.TypeString && args[1].String() != "" {
			id = args[1].String()
		}
		if args[2].Type() == js.TypeString && args[2].String() != "" {
			name = args[2].String()
		}
	} else if len(args) == 2 {
		if args[0].Type() == js.TypeString && args[0].String() != "" {
			name = args[0].String()
		}
		if args[1].Type() == js.TypeString && args[1].String() != "" {
			host = args[1].String()
		}
	}

	var ctx context.Context
	ctx, nodeCancel = context.WithCancel(context.Background())
	msgChan = make(chan rtcmsg.Message, 20)

	go rtc.RunNode(ctx, host, id, name, msgChan)
	return nil
}

func sendMessage(this js.Value, args []js.Value) any {
	if msgChan != nil && len(args) > 0 && args[0].Type() == js.TypeString {
		text := args[0].String()
		name := "Guest"
		if len(args) > 1 && args[1].Type() == js.TypeString {
			name = args[1].String()
		}
		select {
		case msgChan <- rtcmsg.Message{Name: name, Text: text}:
			if js.Global().Get("onMessageSent").Type() == js.TypeFunction {
				js.Global().Call("onMessageSent", name, text)
			}
		default:
		}
	}
	return nil
}

func main() {
	js.Global().Set("RunHost", js.FuncOf(runHost))
	js.Global().Set("RunNode", js.FuncOf(runNode))
	js.Global().Set("SendMessage", js.FuncOf(sendMessage))

	select {}
}

