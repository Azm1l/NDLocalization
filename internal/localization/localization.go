// Package localization reads and writes Naughty Dog localization table files.
package localization

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var unknownString = []byte("UNKNOWN STRING!!!\x00")

var allowedExtensions = map[string]bool{
	".subtitles":          true,
	".common":             true,
	".subtitles-systemic": true,
	".web":                true,
}

type ExportResult struct {
	Version  int
	Count    int
	TextPath string
	IDsPath  string
}

type ImportResult struct {
	Version    int
	Count      int
	OutputPath string
}

type table struct {
	Version int
	Ext     string
	IDs     []uint64
	Strings []string
}

func detectVersion(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.New("file is too small to identify")
	}
	if data[0] == 0 {
		return 0, nil
	}
	if len(data) < 20 {
		return 0, errors.New("file is too small for version 1/2 format")
	}
	if int32(binary.LittleEndian.Uint32(data[16:20])) == 0 {
		return 2, nil
	}
	return 1, nil
}

func readCString(data []byte, offset int) (string, error) {
	if offset < 0 || offset >= len(data) {
		return "", fmt.Errorf("string offset is outside the file: %d", offset)
	}
	rel := bytes.IndexByte(data[offset:], 0)
	if rel < 0 {
		return "", fmt.Errorf("string at offset %d is not NUL-terminated", offset)
	}
	return string(data[offset : offset+rel]), nil
}

func parseBinary(inputPath string) (*table, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("unsupported extension %q", ext)
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	version, err := detectVersion(data)
	if err != nil {
		return nil, err
	}
	t := &table{Version: version, Ext: ext[1:]}
	var count uint32
	cursor := 4

	switch version {
	case 0:
		count = binary.BigEndian.Uint32(data[:4])
	case 1, 2:
		count = binary.LittleEndian.Uint32(data[:4])
	}
	if count > 2_000_000 {
		return nil, fmt.Errorf("unreasonable string count: %d", count)
	}
	entrySize := 8
	if version == 2 {
		entrySize = 16
	}
	headerEnd := int64(4) + int64(count)*int64(entrySize)
	if headerEnd > int64(len(data)) {
		return nil, errors.New("header/offset table extends past the end of the file")
	}

	t.IDs = make([]uint64, int(count))
	offsets := make([]int64, int(count))
	for i := 0; i < int(count); i++ {
		if version == 0 {
			t.IDs[i] = uint64(binary.BigEndian.Uint32(data[cursor : cursor+4]))
			offsets[i] = int64(int32(binary.BigEndian.Uint32(data[cursor+4 : cursor+8])))
		} else if version == 1 {
			t.IDs[i] = uint64(binary.LittleEndian.Uint32(data[cursor : cursor+4]))
			offsets[i] = int64(binary.LittleEndian.Uint32(data[cursor+4 : cursor+8]))
		} else {
			t.IDs[i] = binary.LittleEndian.Uint64(data[cursor : cursor+8])
			offsets[i] = int64(binary.LittleEndian.Uint64(data[cursor+8 : cursor+16]))
		}
		if offsets[i] < 0 || offsets[i] > int64(len(data)) {
			return nil, fmt.Errorf("invalid string offset %d: %d", i+1, offsets[i])
		}
		cursor += entrySize
	}
	stringTableStart := cursor
	t.Strings = make([]string, int(count))
	for i, offset := range offsets {
		absolute := int64(stringTableStart) + offset
		if absolute < 0 || absolute >= int64(len(data)) {
			return nil, fmt.Errorf("string offset %d is outside the file", i+1)
		}
		s, err := readCString(data, int(absolute))
		if err != nil {
			return nil, fmt.Errorf("string %d: %w", i+1, err)
		}
		t.Strings[i] = s
	}
	return t, nil
}

func Export(inputPath string) (*ExportResult, error) {
	t, err := parseBinary(inputPath)
	if err != nil {
		return nil, err
	}
	textPath := inputPath + ".txt"
	idsPath := inputPath + ".ids"

	if err := writeLines(textPath, t.Strings); err != nil {
		return nil, fmt.Errorf("failed to write TXT: %w", err)
	}
	idLines := make([]string, 0, len(t.IDs)+1)
	idLines = append(idLines, fmt.Sprintf("%d|%s", t.Version, t.Ext))
	for _, id := range t.IDs {
		idLines = append(idLines, strconv.FormatUint(id, 10))
	}
	if err := writeLines(idsPath, idLines); err != nil {
		return nil, fmt.Errorf("gagal menulis IDS: %w", err)
	}
	return &ExportResult{Version: t.Version, Count: len(t.Strings), TextPath: textPath, IDsPath: idsPath}, nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSuffix(scanner.Text(), "\r"))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, line := range lines {
		if strings.ContainsRune(line, '\x00') {
			_ = f.Close()
			return errors.New("text contains a NUL character")
		}
		if _, err := w.WriteString(line); err != nil {
			_ = f.Close()
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func Import(txtPath string) (*ImportResult, error) {
	if strings.ToLower(filepath.Ext(txtPath)) != ".txt" {
		return nil, errors.New("import requires an exported .txt file")
	}
	// Pairing is based on the exported base name: 2.subtitles.txt -> 2.subtitles.ids.
	base := strings.TrimSuffix(txtPath, filepath.Ext(txtPath))
	idsPath := base + ".ids"
	idLines, err := readLines(idsPath)
	if err != nil {
		return nil, fmt.Errorf("matching .ids file could not be read (%s): %w", idsPath, err)
	}
	textLines, err := readLines(txtPath)
	if err != nil {
		return nil, fmt.Errorf("could not read TXT: %w", err)
	}
	if len(idLines) == 0 {
		return nil, errors.New("the .ids file is empty")
	}

	header := strings.SplitN(idLines[0], "|", 2)
	if len(header) != 2 {
		return nil, errors.New("the .ids header must use the version|extension format")
	}
	version, err := strconv.Atoi(header[0])
	if err != nil || version < 0 || version > 2 {
		return nil, fmt.Errorf("unsupported version in .ids: %q", header[0])
	}
	ext := "." + strings.ToLower(header[1])
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("unsupported extension in .ids: %q", header[1])
	}
	idStrings := idLines[1:]
	if len(idStrings) != len(textLines) {
		return nil, fmt.Errorf("line count mismatch: TXT=%d, IDS=%d. Import cancelled to prevent an invalid output", len(textLines), len(idStrings))
	}

	ids := make([]uint64, len(idStrings))
	for i, s := range idStrings {
		ids[i], err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID on line %d: %q", i+2, s)
		}
	}
	encoded := make([][]byte, len(textLines))
	offsets := make([]uint64, len(textLines))
	// Offsets are relative to the start of the string table. The marker string
	// is stored first, so the first actual entry starts after its NUL terminator.
	offset := uint64(len(unknownString))
	for i, s := range textLines {
		if strings.ContainsRune(s, '\x00') {
			return nil, fmt.Errorf("text on line %d contains a NUL character", i+1)
		}
		encoded[i] = append([]byte(s), 0)
		offsets[i] = offset
		offset += uint64(len(encoded[i]))
		if offset > math.MaxInt32 && version == 0 {
			return nil, errors.New("data is too large for version 0 offsets")
		}
	}

	entrySize := 8
	if version == 2 {
		entrySize = 16
	}
	headerSize := 4 + len(ids)*entrySize
	stringBytes := len(unknownString)
	for _, b := range encoded {
		stringBytes += len(b)
	}
	out := make([]byte, headerSize+stringBytes)
	cursor := 0
	if version == 0 {
		binary.BigEndian.PutUint32(out[cursor:cursor+4], uint32(len(ids)))
	} else {
		binary.LittleEndian.PutUint32(out[cursor:cursor+4], uint32(len(ids)))
	}
	cursor += 4
	for i, id := range ids {
		switch version {
		case 0:
			if id > math.MaxUint32 {
				return nil, fmt.Errorf("version 0 ID is too large on line %d", i+2)
			}
			binary.BigEndian.PutUint32(out[cursor:cursor+4], uint32(id))
			binary.BigEndian.PutUint32(out[cursor+4:cursor+8], uint32(offsets[i]))
		case 1:
			if id > math.MaxUint32 || offsets[i] > math.MaxUint32 {
				return nil, fmt.Errorf("version 1 ID/offset is too large on line %d", i+2)
			}
			binary.LittleEndian.PutUint32(out[cursor:cursor+4], uint32(id))
			binary.LittleEndian.PutUint32(out[cursor+4:cursor+8], uint32(offsets[i]))
		case 2:
			binary.LittleEndian.PutUint64(out[cursor:cursor+8], id)
			binary.LittleEndian.PutUint64(out[cursor+8:cursor+16], offsets[i])
		}
		cursor += entrySize
	}
	copy(out[cursor:], unknownString)
	cursor += len(unknownString)
	for _, b := range encoded {
		copy(out[cursor:], b)
		cursor += len(b)
	}
	if cursor != len(out) {
		return nil, io.ErrShortWrite
	}

	outputPath := base + ".new"
	if err := os.WriteFile(outputPath, out, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write output: %w", err)
	}
	return &ImportResult{Version: version, Count: len(textLines), OutputPath: outputPath}, nil
}
