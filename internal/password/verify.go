package password

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argon2Params struct {
	time    uint32
	memory  uint32
	threads uint8
}

func Verify(password, encodedHash string) bool {
	params, salt, expectedHash, err := parseHash(encodedHash)
	if err != nil {
		return false
	}

	actualHash := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

func parseHash(encodedHash string) (argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return argon2Params{}, nil, nil, errors.New("invalid password hash format")
	}

	params, err := parseParams(parts[3])
	if err != nil {
		return argon2Params{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != saltLen {
		return argon2Params{}, nil, nil, errors.New("invalid password hash salt")
	}

	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(hash) != int(argon2KeyLen) {
		return argon2Params{}, nil, nil, errors.New("invalid password hash")
	}

	return params, salt, hash, nil
}

func parseParams(encoded string) (argon2Params, error) {
	parts := strings.Split(encoded, ",")
	if len(parts) != 3 {
		return argon2Params{}, errors.New("invalid password hash parameters")
	}

	memory, err := parseUintParam(parts[0], "m", 32)
	if err != nil {
		return argon2Params{}, err
	}
	time, err := parseUintParam(parts[1], "t", 32)
	if err != nil {
		return argon2Params{}, err
	}
	threads, err := parseUintParam(parts[2], "p", 8)
	if err != nil {
		return argon2Params{}, err
	}

	if uint32(memory) != argon2Memory || uint32(time) != argon2Time || uint8(threads) != argon2Threads {
		return argon2Params{}, errors.New("unsupported password hash parameters")
	}

	return argon2Params{memory: uint32(memory), time: uint32(time), threads: uint8(threads)}, nil
}

func parseUintParam(encoded, name string, bitSize int) (uint64, error) {
	prefix := name + "="
	if !strings.HasPrefix(encoded, prefix) {
		return 0, errors.New("invalid password hash parameters")
	}

	value, err := strconv.ParseUint(strings.TrimPrefix(encoded, prefix), 10, bitSize)
	if err != nil {
		return 0, errors.New("invalid password hash parameters")
	}

	return value, nil
}
