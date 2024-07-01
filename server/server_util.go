package server

import "log"

func logError(err error) {
	log.Printf("[ERROR] %s\n", err)
}

func logInfo(msg string) {
	log.Printf("[INFO] %s\n", msg)
}

func (c *connection) logInfo(msg string, args ...interface{}) {
	args = append([]interface{}{c.conn.RemoteAddr().String()}, args...)
	log.Printf("[INFO] [%s] "+msg+"\n", args...)
}

func (c *connection) isBodyClose(i int) bool {
	return i > 4 &&
		c.buf[i-4] == '\r' &&
		c.buf[i-3] == '\n' &&
		c.buf[i-2] == '.' &&
		c.buf[i-1] == '\r' &&
		c.buf[i-0] == '\n'
}

func (c *connection) writeLine(msg string) error {
	msg += "\r\n"
	for len(msg) > 0 {
		n, err := c.conn.Write([]byte(msg))
		if err != nil {
			return err
		}

		msg = msg[n:]
	}

	return nil
}

func (c *connection) readline() (string, error) {
	for {
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			return "", err
		}

		c.buf = append(c.buf, buf[:n]...)
		for i, b := range c.buf {
			// If end of line
			if b == '\n' && i > 0 && c.buf[i-1] == '\r' {
				line := string(c.buf[:i-1])
				c.buf = c.buf[i+1:]
				return line, nil
			}
		}
	}
}

func (c *connection) readTillEndOfBody() (string, error) {
    for {
        for i := range c.buf {
            if c.isBodyClose(i) {
                return string(c.buf[:i-4]), nil
            }
        }

        b := make([]byte, 1024)
        n, err := c.conn.Read(b)
        if err != nil {
            log.Println("error in readTillEndOfBody", err)
            return "", err
        }

        c.buf = append(c.buf, b[:n]...)
    }
}

func (c *connection) readMultiLine() (string, error) {
	for {
		noMoreReads := false
		for i, b := range c.buf {
			if i > 1 && b != ' ' && b != '\t' && c.buf[i-2] == '\r' && c.buf[i-1] == '\n' {
				// i-2 because drop the CRLF, nobody cares after this
				line := string(c.buf[:i-2])
				c.buf = c.buf[i:]
				return line, nil
			}

			noMoreReads = c.isBodyClose(i)
		}

		if !noMoreReads {
			b := make([]byte, 1024)
			n, err := c.conn.Read(b)
			if err != nil {
				return "", err
			}
			c.buf = append(c.buf, b[:n]...)

			// If this gets here more than once it's going to be an infinite loop
		}
	}
}

