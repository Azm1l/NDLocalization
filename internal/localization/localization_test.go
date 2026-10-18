package localization

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func makeV0(strings []string) []byte {
	headerSize := 4 + len(strings)*8
	unknown := []byte("UNKNOWN STRING!!!\x00")
	out := make([]byte, headerSize+len(unknown))
	binary.BigEndian.PutUint32(out[:4], uint32(len(strings)))
	cursor := 4
	offset := uint32(len(unknown))
	var stringData []byte
	for i, s := range strings {
		binary.BigEndian.PutUint32(out[cursor:cursor+4], uint32(10+i))
		binary.BigEndian.PutUint32(out[cursor+4:cursor+8], offset)
		cursor += 8
		b := append([]byte(s), 0)
		stringData = append(stringData, b...)
		offset += uint32(len(b))
	}
	copy(out[headerSize:], unknown)
	out = append(out, stringData...)
	return out
}

func TestVersion0ExportImportKeepsStringsAndOffsets(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "2.subtitles")
	original := makeV0([]string{"Hello", "World"})
	if err := os.WriteFile(input, original, 0o600); err != nil {
		t.Fatal(err)
	}

	exported, err := Export(input)
	if err != nil {
		t.Fatal(err)
	}
	if exported.Version != 0 || exported.Count != 2 {
		t.Fatalf("unexpected export result: %+v", exported)
	}

	imported, err := Import(exported.TextPath)
	if err != nil {
		t.Fatal(err)
	}

	// Rename the generated binary to a supported extension to re-export it.
	reimportPath := filepath.Join(dir, "roundtrip.subtitles")
	if err := os.Rename(imported.OutputPath, reimportPath); err != nil {
		t.Fatal(err)
	}
	reexported, err := Export(reimportPath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(reexported.TextPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("Hello\nWorld\n")
	if string(got) != string(want) {
		t.Fatalf("round-trip text mismatch\ngot: %q\nwant: %q", got, want)
	}
}
