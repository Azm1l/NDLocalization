# NDLocGo

**NDLocGo** is a Go port of the Naughty Dog localization export/import workflow. It is designed for Windows drag-and-drop use: drop a file onto `NDLocGo.exe` and the tool automatically chooses export or import based on the file extension.

> The application has a distinct name so it can coexist with the original C# tool.

## Features

- Export `.subtitles`, `.common`, `.subtitles-systemic`, and `.web` files to `.txt` + `.ids`.
- Import exported TXT files back into a binary `.new` file.
- Supports localization table format versions 0, 1, and 2.
- Validates TXT/IDS line counts before import.
- Does not overwrite the original input file.
- No third-party dependencies.

## Download

Get `NDLocGo-windows-amd64.zip` from the Releases page. It contains `NDLocGo.exe` and this README.

## Usage (Windows)

### Export

1. Back up the original game file.
2. Drag a localization file such as `2.subtitles` onto `NDLocGo.exe`.
3. The tool creates two files in the same folder:
   - `2.subtitles.txt` — the text to translate.
   - `2.subtitles.ids` — metadata containing string IDs and format version. Keep this file unchanged.

### Import

1. Edit `2.subtitles.txt` using a text editor that saves as UTF-8.
2. Keep `2.subtitles.ids` in the same folder. It must come from the same export.
3. Drag `2.subtitles.txt` onto `NDLocGo.exe`.
4. The tool creates `2.subtitles.new`.
5. Validate the output before using a copy to replace any game asset. Do not overwrite the original file until you have a verified backup.

Example files:
```text
2.subtitles
2.subtitles.txt
2.subtitles.ids
2.subtitles.new
```

## Important notes

- Do not delete, edit, or mix `.ids` files from different exports.
- TXT and IDS must have the same number of entries. If they do not, import is cancelled.
- Keep one string per line. Literal newline characters inside a string are not supported by the line-based TXT format.
- Do not rename `.new` to `.txt` to inspect it. `.new` is a binary output file.
- Version 0 import offsets are adjusted to account for the `UNKNOWN STRING!!!\0` marker written before the string table.
- This port has not been verified with every game/platform variant. Always test on copies and validate by exporting the result again before using it in-game.


## Build from source

Requires Go 1.22 or newer. No external dependencies are required.

```bash
go test ./...
go build -trimpath -ldflags="-s -w" -o NDLocGo.exe .
```

To build for Linux:
```bash
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o NDLocGo .
```

## Command-line usage

The same operations can be run from a terminal:

```powershell
.\NDLocGo.exe "D:\Game\2.subtitles"
.\NDLocGo.exe "D:\Game\2.subtitles.txt"
```

## Input/output reference

| Input | Action | Output |
|---|---|---|
| `file.subtitles` | Export | `file.subtitles.txt`, `file.subtitles.ids` |
| `file.common` | Export | `file.common.txt`, `file.common.ids` |
| `file.subtitles-systemic` | Export | `file.subtitles-systemic.txt`, `file.subtitles-systemic.ids` |
| `file.web` | Export | `file.web.txt`, `file.web.ids` |
| `file.subtitles.txt` | Import | `file.subtitles.new` |

## Attribution and licensing

NDLocGo is an independent Go implementation based on the format and workflow of [NaughtyDogLocalizationTool](https://github.com/rm-NoobInCoding/NaughtyDogLocalizationTool) by NoobInCoding. 

NDLocGo is not an official product of Naughty Dog or Sony Interactive Entertainment. All trademarks belong to their respective owners.
