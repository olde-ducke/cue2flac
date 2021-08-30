package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	
	"github.com/olde-ducke/cue2flac/parsecue"
)

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	album, tracks, err := parsecue.ParseCue()
	errorCheck(err)

	// since album name can contain illegal characters, it's better to get rid of them
	replacer := strings.NewReplacer(
									"<",  "",
									">",  "",
									":",  "",
									"\"", "",
									"/",  "",
									"\\", "",
									"|",  "",
									"?",  "",
									"*",  "")
	outputFolder := strings.Join([]string{album["date"], "-", replacer.Replace(album["album"])}, " ")
	err = os.Mkdir(outputFolder, os.ModePerm)
	log.Printf("output folder: %s%s", string(os.PathSeparator), outputFolder)
	errorCheck(err)
	
	// collecting ffmpeg arguments for each track
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
		
		// file name, same as album name above - can contain illegal characters
		var b strings.Builder
		fmt.Fprint(&b , outputFolder, string(os.PathSeparator), tracks[i]["track"], " - ",
					replacer.Replace(tracks[i]["title"]), ".flac" )
		args = append(args, b.String())
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		log.Println("waiting for ffmpeg..., command:", strings.Join(args, " "))
		errorCheck(err)
	}
}
