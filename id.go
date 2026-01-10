package toolbelt

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-chi/chi/v5"
	"github.com/rzajac/zflake"
	"github.com/zeebo/xxh3"
)

var flake *zflake.Gen

func NextID() int64 {
	if flake == nil {
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

	return flake.NextFID()
}

func NextEncodedID() string {
	return EncodeID(NextID())
}

func EncodeID(id int64) string {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(id))
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func EncodedIDToInt64(s string) (int64, error) {
	trimmed := strings.TrimRight(s, "=")
	buf, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(trimmed)
	if err != nil {
		return 0, err
	}
	if len(buf) != 8 {
		return 0, fmt.Errorf("encoded id must decode to 8 bytes, got %d", len(buf))
	}
	return int64(binary.LittleEndian.Uint64(buf)), nil
}

func ChiParamInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func ChiParamEncodedID(r *http.Request, name string) (int64, error) {
	return EncodedIDToInt64(chi.URLParam(r, name))
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
