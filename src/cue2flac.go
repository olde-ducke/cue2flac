package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	
	"github.com/olde-ducke/cue2flac/src/parsecue"
)

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	album, tracks := parsecue.ParseCue()
	
	outputFolder := strings.Join([]string{album["date"], "-", album["album"]}, " ")
	err := os.Mkdir(outputFolder, os.ModePerm)
	log.Printf("output folder: %s%s", string(os.PathSeparator), outputFolder)
	errorCheck(err)
	
	for i, _ := range tracks {
		// ffmpeg options
		args := []string{"-i", album["file"]}
		if i != 0 {
			args = append(args, "-ss", tracks[i]["start"])
		}
		if i != len(tracks) - 1 {
			args = append(args, "-to", tracks[i + 1]["start"])
		}
		
		// add album and track metadata to ffmpeg command
		for k, v := range album {
			if k == "file" { continue }
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", k, v))
		}
		for k, v := range tracks[i] {
			if k == "start" { continue }
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", k, v))
		}
		// TODO: terrible way of doing that through sprintf
		// FIXME: filenames can contain ? and other symbols
		// that are not allowed
		args = append(args, fmt.Sprintf("%s%s%s%s%s%s", outputFolder, string(os.PathSeparator), tracks[i]["track"], " - ", tracks[i]["title"], ".flac" ))
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		log.Println("waiting for ffmpeg..., command:", strings.Join(args, " "))
		errorCheck(err)
	}
}
