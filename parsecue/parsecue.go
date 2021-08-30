package parsecue

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"os"
	"strconv"
	"strings"
)

type Data struct {
	Album	map[string]string
	Track []map[string]string
}

func extractData(input string, i int, f bool) string {
	input = strings.ReplaceAll(input, "\"", "")
	buf := strings.Fields(input)
	if f {
		return strings.Join(buf[i:len(buf) - 1], " ")
	}
	return strings.Join(buf[i:], " ")
}

// converts timestamp from mm:ss:ff to mm:ss.ms
func convertTime(input string) (string, error) {
	buf := strings.Fields(input)
	// assuming that wiki is right, and format is always:
	// INDEX <number> <mm:ss:ff> convert frames to ms for ffmpeg
	// take last part of input line (time code) and split it
	buf = strings.Split(buf[2], ":")
	// returns int 0 if there is an error, so we may proceed anyway
	i, err := strconv.Atoi(buf[2])
	// not dealing with floating point
	i = i * 1_000_000 / 75_000
	var b strings.Builder
	fmt.Fprint(&b, buf[0], ":", buf[1], ".")
	// cue specifies frames instead of ms, 1 frame = 1/75 s, which
	// is 0,01(3), so only one leading zero is possible
	if i < 100 {
		fmt.Fprint(&b, "0", strconv.Itoa(i))	
	} else { fmt.Fprint(&b, strconv.Itoa(i)) }
	return b.String(), err
}

// returns mapped metadata found in cue sheet for album and
// individual tracks
func ParseCue() (*Data, error) {
	// find *.cue files in current dir TODO: setting path from caller?
	matches, err := filepath.Glob(`*.[cC][uU][eE]`)
	// apparently doesn't throw any errors, except "bad pattern"
	if matches == nil {
		return nil, errors.New("no cue sheet file found")
	}
	
	// TODO: for now only first cue sheet file is parsed
	file, err := os.Open(matches[0])
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// start parsing metadata
	data := &Data{Album: make(map[string]string)}
	bufData := make(map[string]string)
	// album variable is used to differentiate between sections
	scanner, album := bufio.NewScanner(file), true
	// scanner will error with lines longer than 65536
	// characters, if line length is greater than 64K,
	// use the Buffer() method to increase the scanner's
	// capacity:
	for scanner.Scan() {
		currentLine := scanner.Text()
		
		// invert album flag when first audio track is encountered
		if album {
			album = !strings.Contains(currentLine,
										  "AUDIO")
		}
		// TODO: there are more tags than listed below
		// SONGWRITER for example
		switch {
		// TODO: figure out better cue-metadata key combination

		// should only encounter these in album section
		case strings.Contains(currentLine, "FILE"):
			data.Album["file"] = extractData(currentLine,
											1, true) 
		case strings.Contains(currentLine, "REM GENRE"):
			data.Album["genre"] = extractData(currentLine,
												 2, false)
		case strings.Contains(currentLine, "REM DATE"):
			data.Album["date"] = extractData(currentLine,
												2, false)
		/* ignore this tag for now
		case strings.Contains(currentLine, "REM DISCID"):
			data.Album["compilation"] = extractData(currentLine,
												  2, false)
		*/
		case strings.Contains(currentLine, "REM COMMENT"):
			data.Album["comment"] = extractData(currentLine, 2, false)
		
		// same for both, album and tracks
		case strings.Contains(currentLine, "TITLE"):
			if album {
				data.Album["album"] = extractData(currentLine, 1, false)
			} else {
				bufData["title"] = extractData(currentLine, 1, false)
			}
		case strings.Contains(currentLine, "PERFORMER"):
			if album {
				data.Album["album_artist"] = 
							extractData(currentLine, 1, false)
			} else {
				bufData["artist"] = extractData(currentLine, 1, false)
			}
		
		// should only encounter these in track section
		case strings.Contains(currentLine, "TRACK"):
			bufData["track"] = extractData(currentLine,
											1, true)
		case strings.Contains(currentLine, "INDEX 01"):
			bufData["start"], err = convertTime(currentLine)
			if err != nil {
				return nil, err
			}
			// add collected metadata to tracks map
			// and clear buffer map
			data.Track = append(data.Track, bufData)
			bufData = make(map[string]string)
		}
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	if album {
		return nil, errors.New("no audio tracks found in cue sheet")
	}

	data.Album["tracktotal"] = strconv.Itoa(len(data.Track))

	return data, nil
}

