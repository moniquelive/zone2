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
	mode := flag.String("mode", "toggle", "on|off|toggle|status|main-status|decode-on|decode-off|decode-status")
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
	operation := strings.ToLower(strings.TrimSpace(*mode))

	switch operation {
	case "main-status":
		state, err := client.QueryMainPower(*timeout)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(protocol.PowerState(state))
		return
	case "decode-status":
		state, err := queryDecodeSwitchState(client, *timeout)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(state)
		return
	case "decode-on", "decode-off":
		if err := ensureStereoInput(client, *timeout); err != nil {
			log.Fatal(err)
		}

		current, err := client.QueryStereoDecodeMode(*timeout)
		if err != nil {
			log.Fatal(err)
		}

		target := byte(protocol.DecodeFiveSevenChStereo)
		if operation == "decode-off" {
			target = protocol.DecodeStereo
		}

		updated, err := client.SetStereoDecodeMode(current, target, *timeout, *verifyAttempts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(protocol.DecodeModeState(updated))
		return
	}

	model, err := client.QueryZone2Model(*timeout)
	if err != nil {
		log.Fatal(err)
	}

	if operation == "status" {
		fmt.Println(protocol.Zone2State(model[1]))
		return
	}

	var target byte
	switch operation {
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
		log.Fatal("mode must be on|off|toggle|status|main-status|decode-on|decode-off|decode-status")
	}

	updated, err := client.SetZone2Status(model, target, *timeout, *verifyAttempts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(protocol.Zone2State(updated[1]))
}

func queryDecodeSwitchState(client *protocol.Client, timeout time.Duration) (string, error) {
	_, channels, err := client.QueryIncomingAudioFormat(timeout)
	if err != nil {
		return "", err
	}

	if !protocol.IsStereoChannelConfig(channels) {
		return "off", nil
	}

	mode, err := client.QueryStereoDecodeMode(timeout)
	if err != nil {
		return "", err
	}
	return protocol.DecodeModeState(mode), nil
}

func ensureStereoInput(client *protocol.Client, timeout time.Duration) error {
	_, channels, err := client.QueryIncomingAudioFormat(timeout)
	if err != nil {
		return err
	}

	if !protocol.IsStereoChannelConfig(channels) {
		return fmt.Errorf("current incoming audio is not stereo; receiver displays multi-channel decode mode instead")
	}

	return nil
}

func wsHost(host string) string {
	if strings.Contains(host, ":") {
		return host
	}

	return host + ":50001"
}
