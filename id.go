package toolbelt

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/rzajac/zflake"
	"github.com/zeebo/xxh3"
)

var flake *zflake.Gen

func init() {
	id, err := machineid.ID()
	if err != nil {
		id = time.Now().Format(time.RFC3339Nano)
	}
	h := xxh3.HashString(id) % (1 << zflake.BitLenGID)
	h16 := uint16(h)

	flake = zflake.NewGen(
		zflake.Epoch(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
		zflake.GID(h16),
	)
}

func NextID() int64 {
	return flake.NextFID()
}

func NextEncodedID() string {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(NextID()))
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func AliasHash(alias string) int64 {
	return int64(xxh3.HashString(alias) & 0x7fffffffffffffff)
}

func AliasHashf(format string, args ...interface{}) int64 {
	return AliasHash(fmt.Sprintf(format, args...))
}

func AliasHashEncoded(alias string) string {
	h := AliasHash(alias)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(h))

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func AliasHashEncodedf(format string, args ...interface{}) string {
	return AliasHashEncoded(fmt.Sprintf(format, args...))
}
