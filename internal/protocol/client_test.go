package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Len(t, frames, 2)
	assert.Equal(t, frameA, frames[0])
	assert.Equal(t, frameB, frames[1])
}

func TestSplitFramesStopsOnTruncatedFrame(t *testing.T) {
	frameA := makeFrame(cmdZone2, 0x00, []byte{0x01})
	truncated := []byte{0x21, 0x01, cmdZone2, 0x00, 0x06, 0x01, 0x02}

	msg := append([]byte{}, frameA...)
	msg = append(msg, truncated...)

	frames := splitFrames(msg)
	require.Len(t, frames, 1)
	assert.Equal(t, frameA, frames[0])
}

func TestParseResponseSuccess(t *testing.T) {
	msg := makeFrame(cmdZone2, 0x00, []byte{0x11, 0x22})

	status, payload, err := parseResponse(msg, cmdZone2)
	require.NoError(t, err)
	assert.Equal(t, byte(0x00), status)
	assert.Equal(t, []byte{0x11, 0x22}, payload)
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
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expects)
		})
	}
}
