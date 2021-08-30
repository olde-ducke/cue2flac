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
	data, err := parsecue.ParseCue()
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
	outputFolder := strings.Join([]string{data.Album["date"], "-", replacer.Replace(data.Album["album"])}, " ")
	err = os.Mkdir(outputFolder, os.ModePerm)
	errorCheck(err)
	
	// collecting ffmpeg arguments for each track
	for i, _ := range data.Track {
		// ffmpeg options
		args := []string{"-i", data.Album["file"]}
		if i != 0 {
			args = append(args, "-ss", data.Track[i]["start"])
		}
		if i != len(data.Track) - 1 {
			args = append(args, "-to", data.Track[i + 1]["start"])
		}
		
		// add album and track metadata to ffmpeg command
		for k, v := range data.Album {
			if k == "file" { continue }
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", k, v))
		}
		for k, v := range data.Track[i] {
			if k == "start" { continue }
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", k, v))
		}
		
		// file name, same as album name above - can contain illegal characters
		var b strings.Builder
		fmt.Fprint(&b , outputFolder, string(os.PathSeparator), data.Track[i]["track"], " - ",
					replacer.Replace(data.Track[i]["title"]), ".flac" )
		args = append(args, b.String())
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		log.Println("waiting for ffmpeg..., command:", strings.Join(args, " "))
		errorCheck(err)
	}
}
