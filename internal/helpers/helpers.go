package helpers

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

func ReadModel(r io.Reader, model any) {

	err := binary.Read(r, binary.BigEndian, model)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
}

func WriteModel(w io.Writer, model any) {
	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, model)

	_, err := w.Write(bin_buf.Bytes())

	if err != nil {
		log.Fatal(err)
	}
	bin_buf.Reset()

}
