package main

import (
	"io"
	"log"
	"os"

	"github.com/jlubawy/go-id3v2"
	"github.com/jlubawy/go-id3v2/id3v230"
	"github.com/jlubawy/isolatin1"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) != 3 {
		log.Fatalln("usage: id3conv <infile> <outfile>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	tag, version, err := id3v2.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}

	if version != id3v230.VersionString {
		log.Fatalf("expected version '%s' but got '%s'\n", id3v230.VersionString, version)
	}

	enc := isolatin1.ISOLatin1(isolatin1.InvalidSkip).NewEncoder()

	newFrames := make(map[string][]byte)
	for id, data := range tag.Frames() {
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

	oldSize := tag.Size()
	tag.SetFrames(newFrames)

	of, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	defer of.Close()

	zeros := make([]byte, tag.Size())
	written, err := of.WriteAt(zeros, 0)
	if err != nil {
		log.Fatalln(err)
	}
	if written != len(zeros) {
		log.Fatalln("error writing")
	}

	of.Seek(0, 0)

	if err := id3v230.Encode(of, tag); err != nil {
		log.Fatalln(err)
	}

	f.Seek(int64(oldSize), 0)

	if _, err := io.Copy(of, f); err != nil {

		log.Fatalln(err)

	}
}
