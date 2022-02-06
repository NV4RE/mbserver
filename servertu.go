package mbserver

import (
	"io"
	"log"
	"time"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	serialConfig.Timeout = 100 * time.Millisecond
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Fatalf("failed to open %s: %v\n", serialConfig.Address, err)
	}
	s.ports = append(s.ports, port)

	s.portsWG.Add(1)
	go func() {
		defer s.portsWG.Done()
		s.acceptSerialRequests(port)
	}()

	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
	leftover := make([]byte, 0)

SkipFrameError:
	for {
		select {
		case <-s.portsCloseChan:
			return
		default:
		}

		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)

		switch err {
		case io.EOF, nil:
		// do nothing
		case serial.ErrTimeout:
			leftover = make([]byte, 0)
			continue SkipFrameError
		default:
			log.Printf("serial read error %v\n", err)
			return
		}

		if bytesRead != 0 {
			//fmt.Printf("leftover: %x\n", leftover)
			leftover = append(leftover, buffer[:bytesRead]...)
			// Set the length of the packet to the number of read bytes.
			//packet := buffer[:bytesRead]
			//fmt.Printf("packet: %x, bytesRead: %d\n", packet, bytesRead)
			//fmt.Printf("data: %x\n", leftover)
			frame, lo, err := NewRTUFrame(leftover)
			if err != nil {
				//log.Printf("bad serial frame error %v\n", err)
				continue SkipFrameError
			}
			leftover = lo

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}
