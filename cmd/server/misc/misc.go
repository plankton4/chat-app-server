package misc

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/plankton4/chat-app-server/cmd/server/config"
)

// StrToUInt32V конвертирует string с uint32
func StrToUInt32V(s string) (value uint32) {
	v, err := strconv.ParseUint(strings.Trim(s, " "), 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}

// StrToUInt16 конвертирует string с uint16.
// При ошибке возвращается defaultValue
func StrToUInt16(s string, defaultValue uint16) (value uint16, err error) {
	v, err := strconv.ParseUint(strings.Trim(s, " "), 10, 16)
	if err != nil {
		value = defaultValue
	} else {
		value = uint16(v)
	}
	return
}

func GetNewUniqueKey() (string, error) {
	c := 100
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(b)), nil
}

func IsConnectedToRemoteServer() bool {
	isRemote := false
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		if strings.Contains(addr.String(), config.RemoteServerAddress) {
			isRemote = true
		}
	}

	return isRemote
}

func RemoveFromSlice[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}
