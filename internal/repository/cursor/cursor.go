package cursor

import (
	"encoding/base64"
	"strconv"
)

func Encode(n int64) string {
	str := strconv.FormatInt(n, 10)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Decode(cursor *string) (*int64, error) {
	if cursor == nil {
		return nil, nil
	}
	var val int64

	decoded, err := base64.StdEncoding.DecodeString(*cursor)
	if err != nil {
		return &val, err
	}
	val, err = strconv.ParseInt(string(decoded), 10, 64)
	return &val, err
}
