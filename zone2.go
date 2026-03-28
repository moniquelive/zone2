package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/moniquelive/zone2/internal/protocol"
)

func main() {
	host := flag.String("host", "", "AVR host/IP")
	mode := flag.String("mode", "toggle", "on|off|toggle|status")
	timeout := flag.Duration("timeout", 4*time.Second, "Socket timeout")
	verifyAttempts := flag.Int("verify", 20, "Verification attempts after a write")
	verbose := flag.Bool("verbose", false, "Print raw RX/TX frames")
	flag.Parse()

	if strings.TrimSpace(*host) == "" {
		log.Fatal("-host is required (AVR host/IP)")
	}

	u := url.URL{Scheme: "ws", Host: wsHost(*host), Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := protocol.NewClient(conn, *verbose)

	model, err := client.QueryZone2Model(*timeout)
	if err != nil {
		log.Fatal(err)
	}

	if *mode == "status" {
		fmt.Println(protocol.Zone2State(model[1]))
		return
	}

	var target byte
	switch *mode {
	case "on":
		target = 1
	case "off":
		target = 0
	case "toggle":
		if model[1] == 0 {
			target = 1
		} else {
			target = 0
		}
	default:
		log.Fatal("mode must be on|off|toggle|status")
	}

	updated, err := client.SetZone2Status(model, target, *timeout, *verifyAttempts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(protocol.Zone2State(updated[1]))
}

func wsHost(host string) string {
	if strings.Contains(host, ":") {
		return host
	}

	return host + ":50001"
}
