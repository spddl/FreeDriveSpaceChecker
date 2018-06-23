package mp3

import (
	"io"
	"os"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

// Run wird gebraucht um die Dateien abzuspielen
func Run(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}
	defer d.Close()

	// (sampleRate, channelNum, bytesPerSample, bufferSizeInBytes int)
	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 44100)
	if err != nil {
		return err
	}
	defer p.Close()

	// fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}
