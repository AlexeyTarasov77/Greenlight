package fields

import (
	"fmt"
	"strconv"
)

type MovieRuntime int32

func (m MovieRuntime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprintf("%d mins", m))), nil
}