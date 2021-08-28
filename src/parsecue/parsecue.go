package parsecue

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"log"
	"os"
	"strconv"
	"strings"
)

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func extractData(input string, i int, f bool) string {
	input = strings.ReplaceAll(input, "\"", "")
	buf := strings.Fields(input)
	if f {
		return strings.Join(buf[i:len(buf) - 1], " ")
	}
	return strings.Join(buf[i:], " ")
}

func extractTime(input string) string {
	buf := strings.Fields(input)
	// assuming that wiki is right, and format is always:
	// INDEX <number> <mm:ss:ff> convert frames to ms for ffmpeg
	// take last part of input line (time code) and split it
	buf = strings.Split(buf[2], ":")
	i, err := strconv.Atoi(buf[2])
	errorCheck(err)
	// not dealing with floating point
	i = i * 1_000_000 / 75_000
	// probably not rational to use builder, i don't know
	var b strings.Builder
	fmt.Fprint(&b, buf[0], ":", buf[1], ".")
	// cue specifies frames instead of ms, 1 frame = 1/75 s, which
	// is 0,01(3), so only one leading zero is possible
	if i < 100 {
		fmt.Fprint(&b, "0", strconv.Itoa(i))	
	} else { fmt.Fprint(&b, strconv.Itoa(i)) }
	return b.String()
}

// returns mapped metadata found in cue sheet for album and
// individual tracks
func ParseCue() (map[string]string, []map[string]string) {
	// Glob expects it's own pattern as input, not regex
	matches, err := filepath.Glob(`*.[cC][uU][eE]`)
	// apparently doesn't throw any errors, except "bad pattern"
	if matches == nil {
		errorCheck(errors.New("no cue sheet file found"))
	}
	
	// TODO: for now only first cue sheet file is parsed
	file, err := os.Open(matches[0])
	errorCheck(err)
	defer file.Close()
	
	albumData := make(map[string]string)
	var trackData []map[string]string
	tempData := make(map[string]string)
	// set album flag to differentiate between sections
	scanner, album := bufio.NewScanner(file), true
	// scanner will error with lines longer than 65536
	// characters, if line length is greater than 64K,
	// use the Buffer() method to increase the scanner's
	// capacity:
	for scanner.Scan() {
		currentLine := scanner.Text()
		
		// invert album flag when first audio track is
		// encountered album metadata SHOULD be over
		if album {
			album = !strings.Contains(currentLine,
										  "AUDIO")
		}
		// TODO: there are more tags than listed below
		// SONGWRITER or whatever, for example,
		// pregap is ignored entirely
		switch {
		// TODO: figure out better cue-metadata key combination
		// should only encounter these in album section
		case strings.Contains(currentLine, "FILE"):
			albumData["file"] = extractData(currentLine,
											1, true) 
		case strings.Contains(currentLine, "REM GENRE"):
			albumData["genre"] = extractData(currentLine,
												 2, false)
		case strings.Contains(currentLine, "REM DATE"):
			albumData["date"] = extractData(currentLine,
												2, false)
		/* ignore this tag
		case strings.Contains(currentLine, "REM DISCID"):
			albumData["compilation"] = extractData(currentLine,
												  2, false)
		*/
		case strings.Contains(currentLine, "REM COMMENT"):
			albumData["comment"] = extractData(currentLine,
												   2, false)
		
		// same for both, album and tracks
		case strings.Contains(currentLine, "TITLE"):
			if album {
				albumData["album"] = extractData(currentLine,
												 1, false)
			} else {
				tempData["title"] = extractData(currentLine,
												 1, false)
			}
		case strings.Contains(currentLine, "PERFORMER"):
			if album {
				albumData["album_artist"] = 
							extractData(currentLine, 1, false)
			} else {
				tempData["artist"] = extractData(currentLine,
													 1, false)
			}
		
		// should only encounter these in track section, same
		// goes for AUDIO
		// section INDEX 01 should be last (ignoring postgap)?
		case strings.Contains(currentLine, "TRACK"):
			tempData["track"] = extractData(currentLine,
											1, true)
		case strings.Contains(currentLine, "INDEX 01"):
			tempData["start"] = extractTime(currentLine)
			// add collected metadata to tracks map
			// and clear buffer map
			trackData = append(trackData, tempData)
			tempData = make(map[string]string)
		}
	}
	errorCheck(scanner.Err())
	
	if album {
		errorCheck(errors.New(
			"no audio tracks found in cue sheet"))
	}
	
	albumData["tracktotal"] = strconv.Itoa(len(trackData))
	
	return albumData, trackData
}

