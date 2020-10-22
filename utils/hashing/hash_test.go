package hashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMd5(t *testing.T) {
	//values given by https://md5calc.com/hash/md5/test
	assert.Equal(t, "098f6bcd4621d373cade4e832627b4f6", CalculateMD5Hash("test"))
	assert.Equal(t, "c61d595888f85f6d30e99ef6cacfcb7d", CalculateMD5Hash("CMSIS"))
}
