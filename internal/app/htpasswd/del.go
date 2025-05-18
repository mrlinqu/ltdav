package htpasswd

import (
	"bufio"
	"io"

	"github.com/mrlinqu/ltdav/internal/app/htpasswd/config"
	"github.com/pkg/errors"
)

func del(cfg config.Config, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()

		if !isUserLine(line, cfg.Username) {
			_, err := out.Write([]byte(line + "\n"))
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "read passwords file error")
	}

	return nil
}
