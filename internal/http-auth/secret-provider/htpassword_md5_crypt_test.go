package secret_provider

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type md5entry struct {
	Magic, Salt, Hash []byte
}

func newEntry(e string) *md5entry {
	parts := strings.SplitN(e, "$", 4)
	if len(parts) != 4 {
		return nil
	}

	return &md5entry{
		Magic: []byte("$" + parts[1] + "$"),
		Salt:  []byte(parts[2]),
		Hash:  []byte(parts[3]),
	}
}

func Test_MD5Crypt(t *testing.T) {
	t.Parallel()

	testCases := [][]string{
		{"apache", "$apr1$J.w5a/..$IW9y6DR0oO/ADuhlMF5/X1"},
		{"pass", "$1$YeNsbWdH$wvOF8JdqsoiLix754LTW90"},
		{"pass", "$apr1$CNF9TtqB$/vuyA6fK.syXqMQsnhBsF."},
		{"topsecret", "$apr1$JI4wh3am$AmhephVqLTUyAVpFQeHZC0"},
		{"topsecret", "$apr1$WocJ4HFf$Pzywmvs0YMdnIIoRJt7Cr0"},
		{"topsecret", "$apr1$7iV4q0x8$HflR8CVQ28hhs4G5BCdup."},
		{"", "$apr1$XVZzfg/M$49LIgq2ajJ/ZcT2BWPF/G/"},
		{"", "$apr1$mTS2ggJm$OV8/p71yNkoHZ6koZDITU."},
		{"", "$apr1$KKlLaWRx$NNQQarBOcF.yj1oUfHQow."},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase[0], func(t *testing.T) {
			t.Parallel()

			e := newEntry(testCase[1])
			result := md5Crypt([]byte(testCase[0]), e.Salt, e.Magic)

			assert.Equal(t, string(result), testCase[1])
		})
	}
}
