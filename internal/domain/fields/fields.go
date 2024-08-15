package fields

import (
	"fmt"
	"strconv"
	"strings"
)

type MovieRuntime int32

func (m MovieRuntime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprintf("%d mins", m))), nil
}

func (m *MovieRuntime) UnmarshalJSON(b []byte) error {
	invalidFormatErr := fmt.Errorf(
		"invalid runtime format, must be '<number> mins' (e.g. '120 mins'). Got %s",
		string(b),
	)
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return invalidFormatErr
	}
	parts := strings.Split(s, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return invalidFormatErr
	}
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	*m = MovieRuntime(i)
	return nil
}