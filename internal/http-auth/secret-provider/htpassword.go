package secret_provider

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-crypt/crypt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type HtpasswordProvider struct {
	filePath string
}

func NewHtpasswordProvider(filePath string) (*HtpasswordProvider, error) {
	return &HtpasswordProvider{
		filePath: filePath,
	}, nil
}

func (p *HtpasswordProvider) Match(username string, password string) bool {
	hash, err := p.getHash(username)
	if err != nil {
		log.Error().Err(err).Msg("[HtpasswordProvider][Match] getHash error")
		return false
	}

	valid, err := crypt.CheckPasswordWithPlainText(password, hash)
	if err != nil {
		log.Error().Err(err).Msg("[HtpasswordProvider][Match] CheckPasswordWithPlainText error")
		return false
	}

	return valid
}

func (p *HtpasswordProvider) getHash(username string) (string, error) {
	f, err := os.Open(filepath.Clean(p.filePath))
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}

	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		str := sc.Text()

		parts := strings.Split(str, ":")
		if len(parts) < 2 {
			return "", errors.New("htpassword incorrect line \"" + str + "\"")
		}

		if parts[0] != username {
			continue
		}

		return parts[1], nil
	}

	return "", nil
}
