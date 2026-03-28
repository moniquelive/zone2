package protocol

import (
	"bytes"
	"strings"
	"testing"
)

func makeFrame(command, status byte, payload []byte) []byte {
	frame := make([]byte, 0, 6+len(payload))
	frame = append(frame, 0x21, 0x01, command, status, byte(len(payload)))
	frame = append(frame, payload...)
	frame = append(frame, 0x0D)
	return frame
}

func TestSplitFramesExtractsValidFrames(t *testing.T) {
	frameA := makeFrame(cmdZone2, 0x00, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06})
	frameB := makeFrame(cmdMenu, 0x00, []byte{0x10})

	msg := []byte{0x00, 0x99}
	msg = append(msg, frameA...)
	msg = append(msg, 0x22, 0x23)
	msg = append(msg, frameB...)

	frames := splitFrames(msg)
	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	if !bytes.Equal(frames[0], frameA) {
		t.Fatalf("first frame mismatch: got % X want % X", frames[0], frameA)
	}

	if !bytes.Equal(frames[1], frameB) {
		t.Fatalf("second frame mismatch: got % X want % X", frames[1], frameB)
	}
}

func TestSplitFramesStopsOnTruncatedFrame(t *testing.T) {
	frameA := makeFrame(cmdZone2, 0x00, []byte{0x01})
	truncated := []byte{0x21, 0x01, cmdZone2, 0x00, 0x06, 0x01, 0x02}

	msg := append([]byte{}, frameA...)
	msg = append(msg, truncated...)

	frames := splitFrames(msg)
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	if !bytes.Equal(frames[0], frameA) {
		t.Fatalf("frame mismatch: got % X want % X", frames[0], frameA)
	}
}

func TestParseResponseSuccess(t *testing.T) {
	msg := makeFrame(cmdZone2, 0x00, []byte{0x11, 0x22})

	status, payload, err := parseResponse(msg, cmdZone2)
	if err != nil {
		t.Fatalf("parseResponse returned error: %v", err)
	}

	if status != 0x00 {
		t.Fatalf("unexpected status: got 0x%02X", status)
	}

	if !bytes.Equal(payload, []byte{0x11, 0x22}) {
		t.Fatalf("payload mismatch: got % X", payload)
	}
}

func TestParseResponseErrors(t *testing.T) {
	tests := []struct {
		name    string
		msg     []byte
		expects string
	}{
		{name: "too short", msg: []byte{0x21, 0x01, cmdZone2}, expects: "response too short"},
		{name: "bad framing", msg: []byte{0x20, 0x01, cmdZone2, 0x00, 0x00, 0x0D}, expects: "bad framing"},
		{name: "unexpected command", msg: makeFrame(cmdMenu, 0x00, []byte{0x01}), expects: "unexpected command"},
		{name: "missing terminator", msg: []byte{0x21, 0x01, cmdZone2, 0x00, 0x01, 0x99, 0x00}, expects: "missing terminator"},
		{name: "length mismatch", msg: []byte{0x21, 0x01, cmdZone2, 0x00, 0x02, 0x99, 0x0D}, expects: "length mismatch"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseResponse(tc.msg, cmdZone2)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.expects) {
				t.Fatalf("unexpected error: got %q want substring %q", err.Error(), tc.expects)
			}
		})
	}
}
