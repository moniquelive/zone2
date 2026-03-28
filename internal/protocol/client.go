package protocol

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	cmdZone2 = 0x2F
	cmdMenu  = 0x14
)

type Client struct {
	conn    *websocket.Conn
	verbose bool
}

func NewClient(conn *websocket.Conn, verbose bool) *Client {
	return &Client{conn: conn, verbose: verbose}
}

func Zone2State(status byte) string {
	if status == 1 {
		return "on"
	}

	return "off"
}

func (c *Client) QueryMenuState(timeout time.Duration) (byte, error) {
	if err := c.sendCommand(cmdMenu, []byte{0xF0}); err != nil {
		return 0, err
	}

	msg, err := c.readMessageForCommand(cmdMenu, timeout)
	if err != nil {
		return 0, err
	}

	status, payload, err := parseResponse(msg, cmdMenu)
	if err != nil {
		return 0, err
	}

	if status != 0x00 {
		return 0, fmt.Errorf("menu query failed: status=0x%02X", status)
	}

	if len(payload) < 1 {
		return 0, fmt.Errorf("menu payload too short")
	}

	return payload[0], nil
}

func (c *Client) QueryZone2Model(timeout time.Duration) ([6]byte, error) {
	var model [6]byte

	for attempt := 0; attempt < 5; attempt++ {
		if err := c.sendCommand(cmdZone2, []byte{0xF0}); err != nil {
			return model, err
		}

		msg, err := c.readMessageForCommand(cmdZone2, timeout)
		if err != nil {
			return model, err
		}

		status, payload, err := parseResponse(msg, cmdZone2)
		if err != nil {
			continue
		}

		if status == 0x85 {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		if status != 0x00 {
			return model, fmt.Errorf("zone2 query failed: status=0x%02X", status)
		}

		if len(payload) < 6 {
			return model, fmt.Errorf("zone2 payload too short: %d", len(payload))
		}

		copy(model[:], payload[:6])
		return model, nil
	}

	return model, errors.New("unable to read zone2 model")
}

func (c *Client) SetZone2Status(current [6]byte, target byte, timeout time.Duration, verifyAttempts int) ([6]byte, error) {
	if current[1] == target {
		return current, nil
	}

	desired := current
	desired[1] = target

	if err := c.sendCommand(cmdZone2, desired[:]); err != nil {
		return current, err
	}

	if ack, err := c.readMessageForCommand(cmdZone2, 400*time.Millisecond); err == nil {
		if status, _, parseErr := parseResponse(ack, cmdZone2); parseErr == nil && status != 0x00 {
			return current, fmt.Errorf("zone2 write failed: status=0x%02X", status)
		}
	}

	if verifyAttempts < 1 {
		verifyAttempts = 1
	}

	last := current
	for attempt := 0; attempt < verifyAttempts; attempt++ {
		time.Sleep(250 * time.Millisecond)
		updated, err := c.QueryZone2Model(timeout)
		if err != nil {
			continue
		}

		last = updated
		if updated[1] == target {
			return updated, nil
		}
	}

	return last, fmt.Errorf("zone2 write did not stick (expected=%s got=%s)", Zone2State(target), Zone2State(last[1]))
}

func (c *Client) sendRaw(payload []byte) error {
	if c.verbose {
		fmt.Printf("TX: % X\n", payload)
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, payload)
}

func (c *Client) sendCommand(command byte, data []byte) error {
	packet := make([]byte, 0, 5+len(data))
	packet = append(packet, 0x21, 0x01, command, byte(len(data)))
	packet = append(packet, data...)
	packet = append(packet, 0x0D)
	return c.sendRaw(packet)
}

func (c *Client) readMessageForCommand(command byte, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)

	for {
		_ = c.conn.SetReadDeadline(deadline)
		messageType, msg, err := c.conn.ReadMessage()
		if err != nil {
			if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
				return nil, fmt.Errorf("timed out waiting for command 0x%02X", command)
			}

			return nil, err
		}

		if messageType != websocket.BinaryMessage {
			continue
		}

		if c.verbose {
			fmt.Printf("RX: % X\n", msg)
		}

		frames := splitFrames(msg)
		if len(frames) == 0 {
			if len(msg) >= 3 && msg[2] == command {
				return msg, nil
			}

			continue
		}

		for _, frame := range frames {
			if len(frame) >= 3 && frame[2] == command {
				return frame, nil
			}
		}
	}
}

func splitFrames(msg []byte) [][]byte {
	var frames [][]byte

	for i := 0; i < len(msg); {
		if msg[i] != 0x21 {
			i++
			continue
		}

		if i+5 >= len(msg) {
			break
		}

		frameLen := int(msg[i+4]) + 6
		if frameLen <= 0 || i+frameLen > len(msg) {
			break
		}

		frame := msg[i : i+frameLen]
		if frame[len(frame)-1] != 0x0D {
			i++
			continue
		}

		frames = append(frames, frame)
		i += frameLen
	}

	return frames
}

func parseResponse(msg []byte, expectedCommand byte) (byte, []byte, error) {
	if len(msg) < 6 {
		return 0, nil, fmt.Errorf("response too short: %d", len(msg))
	}

	if msg[0] != 0x21 || msg[1] != 0x01 {
		return 0, nil, fmt.Errorf("bad framing: % X", msg)
	}

	if msg[2] != expectedCommand {
		return 0, nil, fmt.Errorf("unexpected command 0x%02X", msg[2])
	}

	if msg[len(msg)-1] != 0x0D {
		return 0, nil, fmt.Errorf("missing terminator")
	}

	status := msg[3]
	declaredLen := int(msg[4])
	if declaredLen+6 != len(msg) {
		return 0, nil, fmt.Errorf("length mismatch: declared=%d actual=%d", declaredLen, len(msg))
	}

	payload := msg[5 : len(msg)-1]
	return status, payload, nil
}
