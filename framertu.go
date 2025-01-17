package mbserver

import (
	"encoding/binary"
	"errors"
)

var (
	ErrPacketTooShort = errors.New("packet too short")
	ErrCrcNotMatch    = errors.New("crc not match")
)

// RTUFrame is the Modbus TCP frame.
type RTUFrame struct {
	Address  uint8
	Function uint8
	Data     []byte
	CRC      uint16
}

// NewRTUFrame converts a packet to a Modbus TCP frame.
func NewRTUFrame(packet []byte) (*RTUFrame, []byte, error) {
	// Check the that the packet length.

	if len(packet) < 5 {
		return nil, packet, ErrPacketTooShort
	}

	leftOver := make([]byte, 0)
	// Case of read always 8 byte
	if packet[1] <= 5 {
		if len(packet) > 8 {
			leftOver = packet[8:]
			packet = packet[:8]
		} else if len(packet) < 8 {
			return nil, packet, ErrPacketTooShort
		}
	}

	// Check the CRC.
	pLen := len(packet)
	crcExpect := binary.LittleEndian.Uint16(packet[pLen-2 : pLen])
	crcCalc := crcModbus(packet[0 : pLen-2])
	if crcCalc != crcExpect {
		return nil, make([]byte, 0), ErrCrcNotMatch
	}

	frame := &RTUFrame{
		Address:  packet[0],
		Function: packet[1],
		Data:     packet[2 : pLen-2],
	}

	return frame, leftOver, nil
}

// Copy the RTUFrame.
func (frame *RTUFrame) Copy() Framer {
	cf := *frame
	return &cf
}

// Bytes returns the Modbus byte stream based on the RTUFrame fields
func (frame *RTUFrame) Bytes() []byte {
	bytes := make([]byte, 2)

	bytes[0] = frame.Address
	bytes[1] = frame.Function
	bytes = append(bytes, frame.Data...)

	// Calculate the CRC.
	pLen := len(bytes)
	crc := crcModbus(bytes[0:pLen])

	// Add the CRC.
	bytes = append(bytes, []byte{0, 0}...)
	binary.LittleEndian.PutUint16(bytes[pLen:pLen+2], crc)

	return bytes
}

// GetFunction returns the Modbus function code.
func (frame *RTUFrame) GetFunction() uint8 {
	return frame.Function
}

// GetData returns the RTUFrame Data byte field.
func (frame *RTUFrame) GetData() []byte {
	return frame.Data
}

// SetData sets the RTUFrame Data byte field and updates the frame length
// accordingly.
func (frame *RTUFrame) SetData(data []byte) {
	frame.Data = data
}

// SetException sets the Modbus exception code in the frame.
func (frame *RTUFrame) SetException(exception *Exception) {
	frame.Function = frame.Function | 0x80
	frame.Data = []byte{byte(*exception)}
}
