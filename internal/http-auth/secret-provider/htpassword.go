package secret_provider

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type HtpasswordProvider struct {
	passwordHashes map[string]passwordHash
}

func NewHtpasswordProvider(filePath string) (*HtpasswordProvider, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	hases := map[string]passwordHash{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		str := sc.Text()

		parts := strings.Split(str, ":")
		if len(parts) < 2 {
			return nil, errors.New("htpassword incorrect line \"" + str + "\"")
		}

		switch {
		case strings.HasPrefix(parts[1], "$2y$"):
			hases[parts[0]] = bcryptHash([]byte(parts[1]))
		case strings.HasPrefix(parts[1], "$apr1$"):
			hases[parts[0]] = md5Hash([]byte(parts[1]))
		case strings.HasPrefix(parts[1], "{SHA}"):
			hases[parts[0]] = shaHash([]byte(parts[1][5:]))
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return &HtpasswordProvider{
		passwordHashes: hases,
	}, nil
}

func (p *HtpasswordProvider) Match(username string, password string) bool {
	hash, ok := p.passwordHashes[username]
	if !ok {
		return false
	}

	return hash.IsValid([]byte(password))
}
