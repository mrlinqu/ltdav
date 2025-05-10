package htpasswd

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mrlinqu/ltdav/internal/app/htpasswd/config"
	"github.com/rs/zerolog/log"
)

type App struct{}

func New() App {
	return App{}
}

func (a App) Run(cfg config.Config) error {
	in, err := openInFile(cfg)
	if err != nil {
		//return errors.Wrap(err, "Can't open file")
		return err
	}

	if cfg.Verify {
		err := veryfy(cfg, in)
		if err != nil {
			return err
		}

		return nil
	}

	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)

	if cfg.Del {
		err = del(cfg, in, buf)
	} else {
		err = update(cfg, in, buf)
	}

	if err != nil {
		return err
	}

	err = in.Close()
	if err != nil && err != os.ErrInvalid && err != os.ErrClosed {
		return err
	}

	out, err := openOutFile(cfg)
	if err != nil {
		return err
	}

	_, err = out.Write(buf.Bytes())
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}

func openInFile(cfg config.Config) (io.ReadCloser, error) {
	if cfg.Create {
		return io.NopCloser(nil), nil
	}

	if cfg.FileName == "" {
		return nil, nil
	}

	ret, err := os.Open(cfg.FileName)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("err", err.Error()).Msg("ddddddddddddddd")

	return ret, nil
}

func openOutFile(cfg config.Config) (io.WriteCloser, error) {
	if cfg.FileName == "" {
		return os.Stdout, nil
	}

	flags := os.O_WRONLY | os.O_TRUNC
	if cfg.Create {
		flags = flags | os.O_CREATE
	}

	return os.OpenFile(cfg.FileName, flags, 0644)
}

func isUserLine(line string, username string) bool {
	token := strings.SplitN(line, ":", 2)
	if len(token) != 2 {
		return false
	}

	return token[0] == username
}
