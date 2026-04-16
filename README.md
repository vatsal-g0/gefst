# GEFST

**GEFST (GEnerate File STructure)** is a CLI tool that generates file and folder structures from a JSON configuration.

---

## Features

- Create nested folders and files
- Write file contents directly from config
- Dry-run mode (preview without creating)
- Overwrite control
- Validation mode
- Colored and structured output

---

## Installation

### Using Go

```bash
go install github.com/vatsal-g0/gefst@latest
```
Make sure your `$GOPATH/bin` is in your PATH.

## Usage
```
gefst [options] <config.json>
```

### Example Config
```
{
  "type": "folder",
  "name": "project",
  "children": [
    {
      "type": "file",
      "name": "main.go",
      "content": "package main\n\nfunc main() {}"
    },
    {
      "type": "folder",
      "name": "data",
      "children": []
    }
  ]
}
```

### Example
```
gefst --root ./output example/basic.json
```

**Output:**

```
output/
└── project/
    ├── main.go
    └── data/
```

## Options
- `--root <dir>` → output directory (default: .)
- `--dry-run` → preview only
- `--overwrite` → overwrite existing files
- `--verbose` → detailed logs
- `--validate` → validate config only
- `--quiet` → no output
- `--no-color` → disable colors

## Notes
- Errors are intentionally minimal.
- Input is expected to be valid.

- Please give your feedback, I would really appreciate it.

## License
MIT License
