package proto

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataParse(t *testing.T) {
	asrt, rq := assert.New(t), require.New(t)

	f, err := os.ReadFile("./dataupdate.bin")
	rq.NoError(err)

	p, err := ReadPacket(bytes.NewReader(f))
	if err == io.EOF {
		err = nil
	}
	rq.NoError(err, "should not error reading packet")

	de, err := UnmarshalDataEntries(p.Body)
	rq.NoError(err)
	// asrt.Equal(len(f), n, "should read all bytes")
	rq.NotNil(de, "should not be nil")
	asrt.Len(de, 2, "should have 2 entries")
}
