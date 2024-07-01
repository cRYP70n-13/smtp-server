package server

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	Addr     string
	listener net.Listener
}

type message struct {
	body         string
	from         string
	to           string
	subject      string
	date         string
	clientDomain string
	atmHeaders   map[string]string
	smtpCommands map[string]string
}

type connection struct {
	conn net.Conn
	buf  []byte
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func (s *Server) ListenAndAccept() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.listener = ln

	go s.acceptLoop()

	log.Println("server is up and running on: ", ln.Addr().String())

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Listener closed, stopping accept loop:", err)
				return
			}
			log.Println("Failed to accept connection:", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("Failed to close connection:", err)
		}
	}()

	for {
		c := connection{conn, nil}
		err := c.writeLine("220")
		if err != nil {
			logError(err)
			return
		}

		logInfo("Awaiting EHLO")

		line, err := c.readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logInfo("Connection is closed by the client")
				break
			}
			logError(err)
			return
		}

		if !strings.HasPrefix(line, "EHLO") {
			logError(errors.New("expected EHLO got: " + line))
			return
		}

		msg := message{
			smtpCommands: map[string]string{},
			atmHeaders:   map[string]string{},
		}
		msg.clientDomain = line[len("EHLO "):]
		logInfo("Received EHLO")

		err = c.writeLine("250 ")
		if err != nil {
			logError(err)
			return
		}
		logInfo("Done EHLO")

		for line != "" {
			line, err = c.readline()
			if err != nil {
				logError(err)
				return
			}

			pieces := strings.SplitN(line, ":", 2)
			smtpCommand := strings.ToUpper(pieces[0])

			// Special command without a value
			if smtpCommand == "DATA" {
				err = c.writeLine("354")
				if err != nil {
					logError(err)
					return
				}

				break
			}

			smtpValue := pieces[1]
			msg.smtpCommands[smtpCommand] = smtpValue

			logInfo("Got command: " + line)

			err = c.writeLine("250 OK")
			if err != nil {
				logError(err)
				return
			}
		}

		logInfo("Done SMTP commands, reading ARPA text message headers")

		for {
			line, err = c.readMultiLine()
			if err != nil {
				logError(err)
				return
			}

			if strings.TrimSpace(line) == "" {
				break
			}

			pieces := strings.SplitN(line, ":", 2)
			atmHeader := strings.ToUpper(pieces[0])
			atmValue := pieces[1]
			msg.atmHeaders[atmHeader] = atmValue

			if atmHeader == "SUBJECT" {
				msg.subject = atmValue
			}
			if atmHeader == "TO" {
				msg.to = atmValue
			}
			if atmHeader == "FROM" {
				msg.from = atmValue
			}
			if atmHeader == "DATE" {
				msg.date = atmValue
			}
		}

		logInfo("DONE ARPA text message headers, reading body")

		msg.body, err = c.readTillEndOfBody()
		if err != nil {
			logError(err)
			return
		}

		c.logInfo("Got body (%d bytes)", len(msg.body))

		err = c.writeLine("250 OK")
		if err != nil {
			logError(err)
			return
		}

		c.logInfo("Message:\n%s\n", msg.body)
		c.logInfo("Connection closed")
	}
}
