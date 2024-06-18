package callbackcounter

import "encoding/base64"

// Binary is a thin wrapper around string that is using base64 encoding for []byte.
func (b *Binary) Unwrap() []byte {
	res, err := base64.StdEncoding.DecodeString(string(*b))
	if err != nil {
		panic(err)
	}
	return res
}

// Binary is a thin wrapper around string that is using base64 encoding for []byte.
func (b *Data_Result) Unwrap() []byte {
	res, err := base64.StdEncoding.DecodeString(string(*b))
	if err != nil {
		panic(err)
	}
	return res
}
