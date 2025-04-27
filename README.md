

<p align="center">
  <a href="https://github.com/pashkov256/deletor"><img src="https://raw.githubusercontent.com/pashkov256/media/refs/heads/main/deletor/logo.png" alt="deletor"></a>
</p>
<p align="center">
    <em>Manage and delete files efficiently with an interactive TUI and scriptable CLI.</em>
</p>
<p align="center">
  <a href="https://pkg.go.dev/github.com/pashkov256/deletor"><img src="https://pkg.go.dev/badge/github.com/pashkov256/deletor/v1.svg" alt="deletor"></a>
  <a><img src="https://img.shields.io/github/issues/pashkov256/deletor" alt="deletor"></a>
  <a><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT"></a>
<hr>
</p>



**Deletor** is a handy file deletion tool that combines a powerful text interface (**TUI**) with visual directory navigation, and classic command line mode (**CLI**). It allows you to quickly find and delete files by extension and size, both interactively and through scripts.

## Features
- 🖥️ **Interactive TUI**: Modern text-based user interface for easy file navigation and management
- 🗑️ **Delete by Extension**: Deletes files with specified extensions (e.g., .mp4, .zip)
- 📏 **Size Filter**: Deletes only files larger than the specified size (e.g., 10mb, 1gb)
- 📂 **Directory Navigation**: Easy navigation through directories with arrow keys
- 🎯 **Quick Selection**: Select and delete files with keyboard shortcuts
- ⚙️ **Customizable Options**: Toggle hidden files and confirmation prompts
- 🛠️ **Confirmation Prompt**: Optional confirmation before deleting files
- 🧠 **Rules System**: Create and manage deletion presets for repeated use
- 📊 **Formatted Output**: Clean, aligned display of file information


---
<p align="center">
  <img src="https://raw.githubusercontent.com/pashkov256/media/refs/heads/main/deletor.gif" alt="Project Banner" />
</p>

## 📦 Installation
```bash
go install github.com/pashkov256/deletor
```



## 🛠 Usage

### TUI Mode (default):

```bash
deletor -d ~/Downloads/
```
### CLI Mode (with filters):
```bash



deletor -cli -e mp4,zip -d ~/Downloads/ -s 10mb
```
### Arguments:
`-e, --extensions` — comma-separated list of extensions (for example, mp4,zip,jpg).

`-d, --directory` — the path to the file search directory.

`-s, --size` — minimum file size to delete (for example, 10 kb, 1mb, 1gb).


## ✨ The Power of Dual Modes: TUI and CLI

- TUI mode provides a user-friendly way to navigate and manage files visually, ideal for manual cleanups and exploration.

- CLI mode is perfect for automation, scripting, and quick one-liners. It's essential for server environments, cron jobs, and integrating into larger toolchains.

Unlike many traditional disk usage tools that focus only on visualizing disk space (like *ncdu*, *gdu*, *dua-cli*), Deletor is optimized specifically for fast and targeted file removal.
It offers advanced filtering options by file extension, size, and custom exclusions, making it a powerful tool for real-world file management — not just analysis.


## 📋 Rules System
Deletor supports rule-based file operations through JSON configuration:

1. **Rule Location**:
Automatically stored in `~/.config/deletor/rule.json` (Linux/macOS) or `%APPDATA%\deletor\rule.json` (Windows)

2. **Rule Format** (clean_logs.json example):
```json
{
  "path": "C:\Users\pashkov\Downloads\gws",
  "extensions": [".log", ".tmp"],
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


## Web docs
[https://pashkov256.github.io/deletor-doc](https://pashkov256.github.io/deletor-doc)

## 📜 License
This project is distributed under the **MIT** license.


