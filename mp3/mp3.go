package mp3

import (
	"io"
	"os"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

// Run wird gebraucht um dateien abzuspielen
func Run(name string) { // func mp3run() error {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		panic(err)
	}
	defer d.Close()

	// (sampleRate, channelNum, bytesPerSample, bufferSizeInBytes int)
	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 44100)
	if err != nil {
		panic(err)
	}
	defer p.Close()

	// fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(p, d); err != nil {
		panic(err)
	}
}
