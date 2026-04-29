package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

func hashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(encoded string, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid password hash format")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported password hash algorithm")
	}
	if parts[2] != "v=19" {
		return false, errors.New("unsupported password hash version")
	}
	var memory uint32
	var iterations uint32
	var threads uint8
	params := strings.Split(parts[3], ",")
	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "m="):
			value, err := strconv.ParseUint(strings.TrimPrefix(param, "m="), 10, 32)
			if err != nil {
				return false, err
			}
			memory = uint32(value)
		case strings.HasPrefix(param, "t="):
			value, err := strconv.ParseUint(strings.TrimPrefix(param, "t="), 10, 32)
			if err != nil {
				return false, err
			}
			iterations = uint32(value)
		case strings.HasPrefix(param, "p="):
			value, err := strconv.ParseUint(strings.TrimPrefix(param, "p="), 10, 8)
			if err != nil {
				return false, err
			}
			threads = uint8(value)
		}
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	computed := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, computed) == 1, nil
}
