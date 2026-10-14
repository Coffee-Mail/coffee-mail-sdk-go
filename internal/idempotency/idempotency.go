package idempotency

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	uuidBytes       = 16
	versionByte     = 6
	variantByte     = 8
	versionMask     = 0x0f
	versionFour     = 0x40
	variantMask     = 0x3f
	variantRFC4122  = 0x80
	firstDashIndex  = 8
	secondDashIndex = 12
	thirdDashIndex  = 16
	fourthDashIndex = 20
)

func NewKey() string {
	raw := make([]byte, uuidBytes)
	_, _ = rand.Read(raw)
	raw[versionByte] = raw[versionByte]&versionMask | versionFour
	raw[variantByte] = raw[variantByte]&variantMask | variantRFC4122
	encoded := hex.EncodeToString(raw)
	return encoded[:firstDashIndex] + "-" +
		encoded[firstDashIndex:secondDashIndex] + "-" +
		encoded[secondDashIndex:thirdDashIndex] + "-" +
		encoded[thirdDashIndex:fourthDashIndex] + "-" +
		encoded[fourthDashIndex:]
}
