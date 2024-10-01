package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Rating int32

var ErrInvalidRatingFormat = errors.New("Invalid rating format")

var maxRating = 5

func (r Rating) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d из %d", r, maxRating)

	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

func (r *Rating) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRatingFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")

	ratingValue, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	if len(parts) != 3 || parts[1] != "из" || ratingValue > maxRating {
		return ErrInvalidRatingFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRatingFormat
	}

	*r = Rating(i)
	return nil

}
