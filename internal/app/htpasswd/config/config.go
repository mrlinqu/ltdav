package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type Config struct {
	fs *flag.FlagSet

	Create      bool
	NotUpdate   bool
	BatchMode   bool
	Interactive bool
	Md5         bool
	Sha1        bool
	Bcrypt      bool
	BcryptCost  int
	Plaintext   bool
	Del         bool
	Verify      bool

	FileName string
	Username string
	Password string

	//Args []string
}

func New(allArgs []string) (Config, error) {
	ret := Config{}

	ret.fs = flag.NewFlagSet("", flag.ContinueOnError)
	ret.fs.BoolVar(&ret.Create, "c", false, "Create a new file")
	ret.fs.BoolVar(&ret.NotUpdate, "n", false, "Don't update file; display results on stdout")
	ret.fs.BoolVar(&ret.BatchMode, "b", false, "Use the password from the command line rather than prompting for it.")
	ret.fs.BoolVar(&ret.Interactive, "i", false, "Read password from stdin without verification (for script usage)")
	ret.fs.BoolVar(&ret.Md5, "m", false, "Use MD5 hashing for passwords (default)")
	ret.fs.BoolVar(&ret.Sha1, "s", false, "Use SHA-1 hashing for password (insecure)")
	ret.fs.BoolVar(&ret.Bcrypt, "B", false, "Use bcrypt hashing for password (currently considered to be very secure)")
	ret.fs.IntVar(&ret.BcryptCost, "C", 10, "Set the computing time used for the bcrypt algorithm (higher is more secure but slower, default: 10, valid: 4 to 31)")
	ret.fs.BoolVar(&ret.Plaintext, "p", false, "Do not encrypt the password (plaintext, insecure)")
	ret.fs.BoolVar(&ret.Del, "D", false, "Delete the specified user.")
	ret.fs.BoolVar(&ret.Verify, "v", false, "Verify password for the specified user.")

	ret.fs.Usage = ret.Usage

	err := ret.fs.Parse(allArgs)
	if err != nil {
		return ret, err
	}

	args := ret.fs.Args()

	log.Debug().Str("args", strings.Join(args, ",")).Msg("vvvvvvv")

	if len(args) < 1 {
		return ret, errors.New(ret.usage())
	}

	if !ret.NotUpdate {
		if len(args) < 2 {
			return ret, errors.New(ret.usage())
		}

		ret.FileName = args[0]

		args = args[1:]
	}

	ret.Username = args[0]

	if len(args) > 1 {
		ret.Password = args[1]
	}

	return ret, nil
}

func (c Config) usage() string {
	return `Usage:
	htpasswd [-cimBdpsDv] [-C cost] passwordfile username
	htpasswd -b[cmBdpsDv] [-C cost] passwordfile username password
	
	htpasswd -n[imBdps] [-C cost] username
	htpasswd -nb[mBdps] [-C cost] username password
	
	`
}

func (c Config) Usage() {
	fmt.Fprint(os.Stderr, `Usage:
htpasswd [-cimBdpsDv] [-C cost] passwordfile username
htpasswd -b[cmBdpsDv] [-C cost] passwordfile username password

htpasswd -n[imBdps] [-C cost] username
htpasswd -nb[mBdps] [-C cost] username password

`)

	c.fs.PrintDefaults()
}
