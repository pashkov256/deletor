<p align="center">
  <a href="https://github.com/pashkov256/deletor"><img src="https://raw.githubusercontent.com/pashkov256/media/refs/heads/main/deletor/logo_v3.png" alt="deletor"></a>
</p>

<p align="center">
        <a href="https://img.shields.io/github/stars/pashkov256/deletor?style=flat"><img src="https://img.shields.io/github/stars/pashkov256/deletor?style=flat"></a>
        <a><img src="https://codecov.io/gh/pashkov256/deletor/graph/badge.svg?token=AGOWZDF04Y" alt="codecov"></a>
  <br/>
        <a href="https://img.shields.io/github/issues-raw/pashkov256/deletor?style=flat-square"><img src="https://img.shields.io/github/issues-raw/pashkov256/deletor?style=flat-square"/></a>
         <a href="https://goreportcard.com/report/github.com/pashkov256/deletor"> <img src="https://goreportcard.com/badge/github.com/pashkov256/deletor"/></a>
        <a><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT"></a>

<p align="center">
    <em>Manage and delete files efficiently with an interactive TUI and scriptable CLI.</em>
</p>

<hr>
</p>



<a href="https://code2tutorial.com/tutorial/dcba8e56-33cd-4d67-b9ee-0c9f3c276b6e/index.md"><b>Deletor</b></a> is a handy file deletion tool that combines a powerful text interface (**TUI**) with visual directory navigation, and classic command line mode (**CLI**). It allows you to quickly find and delete files by extension and size, both interactively and through scripts.

## Features
- 🖥️ **Interactive TUI**: Modern text-based user interface for easy file navigation and management
- ♻️ **Safe Deletion: Files**: Are moved to the system trash/recycle bin instead of permanent deletion
- 🧹 **OS Cache Cleaner**: Free up space by deleting temporary system cache
- 🛠️ **Deep Customization** Shape the tool to behave exactly how you need
- 🧠 **Rules System**: Create and manage deletion presets for repeated use
- 📖 **Log Operations**: Log the various fields and look at the tui table, or parse the file  
- ⏳ **Modification Time Filter**: Delete files older,newer than X days/hours/minutes
- 📏 **Size Filter**: Deletes only files larger than the specified size (e.g., 10mb, 1gb)
- 🗑️ **Extensions Filter**: Deletes files with specified extensions (e.g., .mp4, .zip)
- 📂 **Directory Navigation**: Easy navigation through directories with arrow keys
- 🎯 **Quick Selection**: Select and delete files with keyboard shortcuts
- ✅ **Confirmation Prompt**: Optional confirmation before deleting files

---
<p align="center">
  <img src="https://raw.githubusercontent.com/pashkov256/media/refs/heads/main/deletor2.gif" alt="Project Banner" />
</p>

## 📦 Installation

### Using Go
```bash
go install github.com/pashkov256/deletor@latest
```

## 🛠 Usage

### TUI Mode (default):

```bash
deletor
```
### CLI Mode (with filters):
```bash
deletor -cli -d ~/Downloads -e mp4,zip  --min-size 10mb -subdirs --exclude data,backup
```
### Dev launch:
```bash
go run . -cli -d ~/Downloads -e mp4,zip  --min-size 10mb -subdirs --exclude data,backup
```

### Arguments:
`-e, --extensions` — comma-separated list of extensions (for example, mp4,zip,jpg).

`-d, --directory` — the path to the file search directory.

`--min-size` — minimum file size to delete (for example, 10 kb, 1mb, 1gb).

`--max-size` — maximum file size to delete (for example, 10kb, 1mb, 1gb).

`--exclude` - exclude specific files/paths (e.g. data,backup)

`-subdirs` - include subdirectories in scan, default false

`-progress` - display a progress bar during file scanning

`-confirm-delete` - confirmation before deleting files


## ✨ The Power of Dual Modes: TUI and CLI

- TUI mode provides a user-friendly way to navigate and manage files visually, ideal for manual cleanups and exploration.

- CLI mode is perfect for automation, scripting, and quick one-liners. It's essential for server environments, cron jobs, and integrating into larger toolchains.

Unlike many traditional disk usage tools that focus only on visualizing disk space (like *ncdu*, *gdu*, *dua-cli*), Deletor is optimized specifically for fast and targeted file removal.
It offers advanced filtering options by file extension, size, and custom exclusions, making it a powerful tool for real-world file management — not just analysis.


## 📋 Rules System
Deletor supports rule-based file operations through JSON configuration:

1. **Rule Location**:
Automatically stored in `~/.config/deletor/rule.json` (Linux/macOS) or `%APPDATA%\deletor\rule.json` (Windows)

2. **Rule Format** (rule.json example):
```json
{
  "path": "C:\Users\pashkov\Downloads\gws",
  "extensions": [".log", ".tmp"],
  "exclude": ["backup", "important"],
  "min_size": "10mb"
}
```
3.  **Key Features**:
- Create/edit rules via TUI or manual JSON editing

- Combine multiple filters (extension + size + exclusions)

- Share rules between machines



## 🛠 Contributing
We welcome and appreciate any contributions to Deletor!
There are many ways you can help us grow and improve:

- **🐛 Report Bugs** — Found an issue? Let us know by opening an issue.
- **💡 Suggest Features** — Got an idea for a new feature? We'd love to hear it!
- **📚 Improve Documentation** — Help us make the docs even clearer and easier to use.
- **💻 Submit Code** — Fix a bug, refactor code, or add new functionality by submitting a pull request.

Before contributing, please take a moment to read our [CONTRIBUTING.md](https://github.com/pashkov256/deletor/blob/main/CONTRIBUTING.md) guide.
It explains how to set up the project, coding standards, and the process for submitting contributions. 

Together, we can make Deletor even better! 🚀


## AI docs
<a href="https://code2tutorial.com/tutorial/dcba8e56-33cd-4d67-b9ee-0c9f3c276b6e/index.md">https://code2tutorial.com/tutorial/dcba8e56-33cd-4d67-b9ee-0c9f3c276b6e/index.md</a>



## 📜 License
This project is distributed under the **MIT** license.

--- 
### Thank you to these wonderful people for their contributions!

<a href="https://github.com/pashkov256/deletor/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pashkov256/deletor" />
</a>
