package main

import (
	"flag"
	"image"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/TeamTrumpet/waltz/waltz"
)

// Usage: waltz --crop=0x0,32x32 --resize=16x16

func main() {
	// use all CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// crop points
	var cropTLX, cropTLY, cropBRX, cropBRY int

	// resize points
	var resizeX, resizeY int

	// string components
	var crop, resize string

	flag.StringVar(&crop, "crop", "", "crop dimensions as 0x0,32x32 to indicate bounds")
	flag.StringVar(&resize, "resize", "", "resize dimensions after the crop as 16x16")

	flag.Parse()

	crop1 := strings.Split(crop, ",")

	if len(crop1) != 2 {
		log.Fatalf("crop invalid, should be in the form of 0x0,32x32")
	}

	crop2TL := strings.Split(crop1[0], "x")

	if len(crop2TL) != 2 {
		log.Fatalf("crop invalid, should be in the form of 0x0,32x32")
	}

	crop2BR := strings.Split(crop1[1], "x")

	if len(crop2BR) != 2 {
		log.Fatalf("crop invalid, should be in the form of 0x0,32x32")
	}

	var err error

	if cropTLX, err = strconv.Atoi(crop2TL[0]); err != nil {
		log.Fatalln("crop invalid, should be in the form of 0x0,32x32")
	}

	if cropTLY, err = strconv.Atoi(crop2TL[1]); err != nil {
		log.Fatalln("crop invalid, should be in the form of 0x0,32x32")
	}

	if cropBRX, err = strconv.Atoi(crop2BR[0]); err != nil {
		log.Fatalln("crop invalid, should be in the form of 0x0,32x32")
	}

	if cropBRY, err = strconv.Atoi(crop2BR[1]); err != nil {
		log.Fatalln("crop invalid, should be in the form of 0x0,32x32")
	}

	cropRectangle := image.Rect(cropTLX, cropTLY, cropBRX, cropBRY)

	resize1 := strings.Split(resize, "x")

	if len(resize1) < 1 {
		log.Fatalln("resize invalid, should be in the form 16x16")
	}

	if resizeX, err = strconv.Atoi(resize1[0]); err != nil {
		log.Fatalln("resize invalid, should be in the form 16x16")
	}

	if len(resize1) == 2 {
		if resizeY, err = strconv.Atoi(resize1[1]); err != nil {
			log.Fatalln("resize invalid, should be in the form 16x16")
		}
	}

	if err := waltz.Do(os.Stdin, os.Stdout, &cropRectangle, resizeX, resizeY); err != nil {
		log.Fatalf("An error occured performing the resize: %s\n", err.Error())
	}
}
