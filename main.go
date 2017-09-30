package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/jlubawy/go-id3v2"
	"github.com/jlubawy/go-id3v2/id3v230"
	"github.com/jlubawy/isolatin1"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: id3conv <source file> [destination file]")
		os.Exit(1)
	}

	// Read source file into a buffer
	sb, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	sourceBuffer := bytes.NewReader(sb)

	// Read the source ID3 tags
	sourceID3Tags, sourceID3TagsVersion, err := id3v2.Decode(sourceBuffer)
	if err != nil {
		log.Fatalln(err)
	}
	if sourceID3TagsVersion != id3v230.VersionString {
		log.Fatalf("expected ID3 version '%s' but got '%s'\n", id3v230.VersionString, sourceID3TagsVersion)
	}
	sourceID3TagsSize := sourceID3Tags.Size()
	sourceDataLen := int64(sourceBuffer.Len()) - int64(sourceID3TagsSize)

	// Convert the source tags' charset
	enc := isolatin1.ISOLatin1(isolatin1.InvalidSkip).NewEncoder()
	newFrames := make(map[string][]byte)
	for id, data := range sourceID3Tags.Frames() {
		if id == "TSSE" {
			continue
		}

		isoData, err := enc.Bytes(data[1:])
		if err != nil {
			log.Fatalf("%s (%s)\n", os.Args[1], err)
		}

		newFrames[id] = make([]byte, len(isoData)+1)
		newFrames[id][0] = 0
		copy(newFrames[id][1:], isoData)
	}
	sourceID3Tags.SetFrames(newFrames)
	destID3TagsSize := sourceID3Tags.Size()

	// Create a destination buffer the size of the new ID3 tags and the original data length
	destBuf := bytes.NewBuffer(make([]byte, 0, int64(destID3TagsSize)+int64(sourceDataLen)))

	// Write the new ID3 tags
	if err := id3v230.Encode(destBuf, sourceID3Tags); err != nil {
		log.Fatalln(err)
	}

	// Write the original data
	sourceBuffer.Seek(int64(sourceID3TagsSize), 0)
	if _, err := io.Copy(destBuf, sourceBuffer); err != nil {
		log.Fatalln(err)

	}

	if len(os.Args) == 3 {
		// If a destination file was provided use it
		if err := ioutil.WriteFile(os.Args[2], destBuf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}
	} else {
		// Else truncate the source file
		if err := ioutil.WriteFile(os.Args[1], destBuf.Bytes(), 0666); err != nil {
			log.Fatal(err)
		}
	}
}
