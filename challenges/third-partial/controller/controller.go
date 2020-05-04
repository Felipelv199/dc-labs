package controller

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nanomsg/mangos/protocol/surveyor"
	"go.nanomsg.org/mangos"

	// register transports

	_ "go.nanomsg.org/mangos/transport/all"
)

var controllerAddress = "tcp://localhost:40899"

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

func Start() {
	var sock mangos.Socket
	var err error
	var msg []byte
	if sock, err = surveyor.NewSocket(); err != nil {
		die("can't get new surveyor socket: %s", err)
	}
	if err = sock.Listen(controllerAddress); err != nil {
		die("can't listen on surveyor socket: %s", err.Error())
	}
	err = sock.SetOption(mangos.OptionSurveyTime, time.Second/2)
	if err != nil {
		die("SetOption(): %s", err.Error())
	}
	for {
		time.Sleep(time.Second)
		fmt.Println("SERVER: SENDING DATE SURVEY REQUEST")
		if err = sock.Send([]byte("DATE")); err != nil {
			die("Failed sending survey: %s", err.Error())
		}
		for {
			if msg, err = sock.Recv(); err != nil {
				break
			}
			msgSplitted := strings.Split(string(msg), "|")
			fmt.Printf("SERVER: RECEIVED CLIENT(%s) MESSAGE: \"%s\" SURVEY RESPONSE\n",
				msgSplitted[0], msgSplitted[1])
		}
		fmt.Println("SERVER: SURVEY OVER")
	}
}

