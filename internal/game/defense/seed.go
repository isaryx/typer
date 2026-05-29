package defense

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"time"
)

// resolveSeed returns explicit when non-zero; otherwise a random seed for this run.
func resolveSeed(explicit uint64) uint64 {
	if explicit != 0 {
		return explicit
	}
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err == nil {
		return binary.LittleEndian.Uint64(b[:])
	}
	return uint64(time.Now().UnixNano())
}
