package htpasswd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"syscall"

	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/go-crypt/crypt/algorithm/md5crypt"
	"github.com/go-crypt/crypt/algorithm/plaintext"
	"github.com/go-crypt/crypt/algorithm/sha1crypt"
	"github.com/mrlinqu/ltdav/internal/app/htpasswd/config"
	"github.com/pkg/errors"
	"golang.org/x/term"
)

func update(cfg config.Config, in io.Reader, out io.Writer) error {
	hasher, err := getHasher(cfg)
	if err != nil {
		return err
	}

	passwd, err := getPassword(cfg)
	if err != nil {
		return err
	}

	hash, err := hasher.Hash(passwd)
	if err != nil {
		return err
	}

	hashString := hash.String()

	err = updateOrAddLines(cfg.Username, hashString, in, out)
	if err != nil {
		return err
	}

	return nil
}

func getHasher(cfg config.Config) (algorithm.Hash, error) {
	if cfg.Bcrypt {
		return bcrypt.New(
			bcrypt.WithIterations(cfg.BcryptCost),
		)
	}

	if cfg.Sha1 {
		return sha1crypt.New()
	}

	if cfg.Plaintext {
		return plaintext.New()
	}

	return md5crypt.New()
}

func getPassword(cfg config.Config) (string, error) {
	if cfg.BatchMode {
		return cfg.Password, nil
	}

	fmt.Print("Enter password: ")
	input, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()

	fmt.Print("Re-type new password: ")
	retry, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println()

	if !bytes.Equal(input, retry) {
		return "", errors.New("password is not equal")
	}

	return string(input), nil
}

func updateOrAddLines(username string, hash string, in io.Reader, out io.Writer) error {
	found, err := updateLines(username, hash, in, out)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	err = addLine(username, hash, out)

	return err
}

func updateLines(username string, hash string, in io.Reader, out io.Writer) (bool, error) {
	if in == nil {
		return false, nil
	}

	found := false

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()

		newLine, ok := getNewLine(line, username, hash)
		found = found || ok

		_, err := out.Write([]byte(newLine + "\n"))
		if err != nil {
			return false, err
		}
	}

	if err := scanner.Err(); err != nil {
		return false, errors.Wrap(err, "read passwords file error")
	}

	return found, nil
}

func getNewLine(line string, username string, hash string) (string, bool) {
	if !isUserLine(line, username) {
		return line, false
	}

	return username + ":" + hash, true
}

func addLine(username string, hash string, out io.Writer) error {
	_, err := out.Write([]byte(username + ":" + hash + "\n"))

	return err
}
