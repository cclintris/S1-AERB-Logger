package util

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer/constant"
)

var (
	ErrUnitUndefined = errors.New("undefined unit")
)

func ParseUnit(bufSize string) (int, error) {
	s := strings.Split(bufSize, " ")
	quantity, _ := strconv.Atoi(s[0])
	unit := strings.ToUpper(s[1])

	switch unit {
	case "B":
		return quantity * int(math.Round(B)), nil
	case "KB":
		return quantity * int(math.Round(KB)), nil
	case "MB":
		return quantity * int(math.Round(MB)), nil
	case "GB":
		return quantity * int(math.Round(GB)), nil
	case "TB":
		return quantity * int(math.Round(TB)), nil
	case "PB":
		return quantity * int(math.Round(PB)), nil
	case "EB":
		return quantity * int(math.Round(EB)), nil
	case "ZB":
		return quantity * int(math.Round(ZB)), nil
	case "YB":
		return quantity * int(math.Round(YB)), nil
	default:
		fmt.Println("Undefined size unit")
	}

	return -1, ErrUnitUndefined
}
