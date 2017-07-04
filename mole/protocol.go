package mole

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
)

var (
	HEADER = []byte("MOLE")
)

// TODO replace this struct by fixed-size [32]byte
// so we don't need to use gob to encode/decode the command
type command struct {
	Cmd      string // cmdJoin, cmdLeave, cmdNewWorker, cmdPing
	AgentID  string // require on cmdJoin / cmdLeave / cmdPing
	WorkerID string // require on cmdNewWorker
}

var (
	cmdJoin      = "join"  // agent -> master
	cmdLeave     = "leave" // agent ->  master
	cmdPing      = "ping"  // master -> agent
	cmdNewWorker = "new"   // master -> agent (with new workerID), agent -> master (notify back with the same workerID that conn established)
)

func (cmd *command) valid() error {
	switch cmd.Cmd {
	case cmdJoin, cmdLeave, cmdPing:
		if cmd.AgentID == "" {
			return errors.New("protocol: agent id required")
		}
	case cmdNewWorker:
		if cmd.WorkerID == "" {
			return errors.New("protocol: worker id required")
		}
	default:
		return errors.New("protocol: unknown command")
	}
	return nil
}

func Encode(msg []byte) []byte {
	ret := make([]byte, 0)

	ret = append(ret, HEADER...) // write header, 4 bytes

	lenBytes := int2bytes(len(msg))
	ret = append(ret, lenBytes...) // write length of msg body, 4 bytes

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, msg)
	ret = append(ret, buf.Bytes()...) // write msg body

	return ret
}

type Decoder struct {
	r     io.Reader // read from
	store []byte    // store the read out bytes
}

// NewDecoder returns a new protocol decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:     r,
		store: make([]byte, 0),
	}
}

func (d *Decoder) Buffered() []byte {
	return d.store
}

// Decode reads the next protocol-encoded command from reader
// Note that Decode() is not concurrency safe.
func (d *Decoder) Decode() (*command, error) {
	var (
		headerN = len(HEADER) + 4    // header + length
		buf     = make([]byte, 1024) // each piece read out
		n       int
		err     error
	)

	// if previous buffered data length over than one header+length
	// consume them firstly
	if len(d.Buffered()) >= headerN {
		goto READ_HEADER
	}

READ_MORE:
	// read one piece of data
	n, err = d.r.Read(buf)
	if err != nil {
		return nil, err
	}

	// accumulated append to p.store
	d.store = append(d.store, buf[:n]...)

	// read more if short than header
	if len(d.Buffered()) < headerN {
		goto READ_MORE
	}

READ_HEADER:
	// scan the read out buffered data...
	// read the header and lenghtN firstly
	var (
		hl      = d.store[:headerN] // header and length bytes
		header  = hl[:len(HEADER)]  // header bytes
		length  = hl[len(HEADER):]  // length bytes
		lengthN = bytes2int(length) // length int
	)

	// ensure protocol HEADER
	if string(header) != "MOLE" {
		// slice down the consumed abnormal data in d.store
		d.store = d.store[headerN:]
		return nil, errors.New("NOT MOLE PROTOCOL")
	}

	// if readout data not contains a full body, continue read
	if len(d.Buffered()) < headerN+lengthN {
		goto READ_MORE
	}

	// read the body out
	body := d.store[headerN : headerN+lengthN]

	// slice down the consumed data in d.store
	d.store = d.store[headerN+lengthN:]

	var (
		cmd    *command
		buffer = bytes.NewBuffer(body)
	)
	if err := gob.NewDecoder(buffer).Decode(&cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}

func newCmd(cmd, aid, wid string) []byte {
	buf := bytes.NewBuffer(nil)
	gob.NewEncoder(buf).Encode(command{cmd, aid, wid})
	return Encode(buf.Bytes())
}
