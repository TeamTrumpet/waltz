package waltz

import (
	"errors"
	"image"
	"io"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var (
	ErrInvalidInputFormat = errors.New("invalid input format")
)

// Do performs the resize
func Do(r io.Reader, w io.Writer, crop *image.Rectangle, width, height int) error {
	// read it
	img, err := imaging.Decode(r)
	if err != nil {
		return err
	}

	// if crop isn't nil
	if crop != nil {
		// then crop it
		img = imaging.Crop(img, *crop)
	}

	// resize it
	img = imaging.Resize(img, width, height, imaging.MitchellNetravali)

	// write it
	if err := imaging.Encode(w, img, imaging.PNG); err != nil {
		return err
	}

	return nil
}

func ParseResize(resize string) (int, int, error) {
	var resizeX, resizeY int
	var err error

	resize1 := strings.Split(resize, "x")

	if len(resize1) < 1 {
		return 0, 0, ErrInvalidInputFormat
	}

	if resizeX, err = strconv.Atoi(resize1[0]); err != nil {
		return 0, 0, ErrInvalidInputFormat
	}

	if len(resize1) == 2 {
		if resizeY, err = strconv.Atoi(resize1[1]); err != nil {
			return 0, 0, ErrInvalidInputFormat
		}
	}

	return resizeX, resizeY, nil
}
