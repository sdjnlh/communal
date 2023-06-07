package password

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestEncrypt(t *testing.T) {
	t.Log("test encrypter")

	encrypted := "sha512:15000:EiU5S1wZMl7UerVKNmlHIMEHkpMrBAAc9S7mMadJd7rC9rGrJ3b8+0W1vUjX2JpY8YDPFdxsH6JmUD6ZfuD8Hw==:VavlnOebygAjlsLniO0jwmxsr/lvaurpXVszOXb6hKFafmD2hKbhtfD2BbFxDSJ+0Cj7F+nTbqKLY74j0Z8IGA=="
	origin := "23333"
	assert.Equal(t, Validate(origin, encrypted), true)
}
