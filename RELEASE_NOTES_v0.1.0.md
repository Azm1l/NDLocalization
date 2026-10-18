# NDLocGo v0.1.0

Initial Go port of the Naughty Dog localization export/import workflow.

## Highlights
- Windows drag-and-drop: drop a localization binary or exported TXT onto `NDLocGo.exe`.
- Exports `.subtitles`, `.common`, `.subtitles-systemic`, and `.web` to `.txt` + `.ids`.
- Imports paired TXT/IDS files to `.new` output.
- Validates TXT/IDS line counts before import.
- Fixes version 0 import offsets to account for the `UNKNOWN STRING!!!\0` marker.

## Important
This release has not been verified against every game/platform build. Keep backups and validate output before replacing any game asset.
