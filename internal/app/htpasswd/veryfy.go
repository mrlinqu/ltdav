package htpasswd

import (
	"bufio"
	"io"
	"strings"

	"github.com/go-crypt/crypt"
	"github.com/mrlinqu/ltdav/internal/app/htpasswd/config"
	"github.com/pkg/errors"
)

func veryfy(cfg config.Config, in io.Reader) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()

		if !isUserLine(line, cfg.Username) {
			continue
		}

		token := strings.SplitN(line, ":", 2)

		valid, err := crypt.CheckPasswordWithPlainText(cfg.Password, token[1])
		if err != nil {
			return errors.Wrap(err, "passwod check error")
		}

		if !valid {
			return errors.New("password not vald")
		}

		return nil
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "read passwords file error")
	}

	return nil
}
