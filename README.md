<p align="center">
  <a href="https://github.com/pashkov256/deletor"><img src="https://raw.githubusercontent.com/pashkov256/media/refs/heads/main/deletor/logo_v3.png" alt="deletor"></a>
</p>

<p align="center">
          <a><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT"></a>
        <a><img src="https://codecov.io/gh/pashkov256/deletor/graph/badge.svg?token=AGOWZDF04Y" alt="codecov"></a>
         <a href="https://goreportcard.com/report/github.com/pashkov256/deletor"> <img src="https://goreportcard.com/badge/github.com/pashkov256/deletor"/></a>

<p align="center">
    <em>Manage and delete files efficiently with an interactive TUI and scriptable CLI.</em>
</p>

<hr>
</p>



<a href="https://code2tutorial.com/tutorial/3aac813f-99c2-453f-819f-c80e4322e068/index.md"><b>Deletor</b></a> is a handy file deletion tool that combines a powerful text interface (**TUI**) with visual directory navigation, and classic command line mode (**CLI**). With it, you can quickly find and delete files by filters, send them to the trash or completely erase them, as well as clear the cache, both interactively and through scripts.

## Features
- 🖥️ **Interactive TUI**: Modern text-based user interface for easy file navigation and management
- ♻️ **Safe Deletion: Files**: Are moved to the system trash/recycle bin instead of permanent deletion
- 🧹 **OS Cache Cleaner**: Free up space by deleting temporary system cache
- 🛠️ **Deep Customization** Shape the tool to behave exactly how you need
- 🧠 **Rules System**: Save your filter settings and preferences for quick access
- 📖 **Log Operations**: Log the various fields and look at the tui table, or parse the file  
- ⏳ **Modification Time Filter**: Delete files older,newer than X days/hours/minutes
- 📏 **Size Filter**: Deletes only files larger than the specified size
- 🗑️ **Extensions Filter**: Deletes files with specified extensions
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

### ⚙️ CLI Flags

| Flags           | Description                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `-e`           | Comma-separated list of extensions (e.g., `mp4,zip,jpg`).                   |
| `-d`           | Path to the file search directory.                                          |
| `--min-size`   | Minimum file size to delete (e.g., `10kb`, `1mb`, `1gb`).                   |
| `--max-size`   | Maximum file size to delete (e.g., `10kb`, `1mb`, `1gb`).                   |
| `--older`      | Modification time older than (e.g., `1sec`, `2min`, `3hour`, `4day`).       |
| `--newer`      | Modification time newer than (e.g., `1sec`, `2min`, `3hour`, `4day`).       |
| `--exclude`    | Exclude specific files/paths (e.g., `data`, `backup`).                      |
| `-subdirs`     | Include subdirectories in scan. Default is false.                           |
| `-prune-empty` | Delete empty folders after scan.                                            |
| `-progress`    | Display a progress bar during file scanning.                                |
| `-skip-confirm`| Skip the confirmation of deletion.                                          |


## ✨ The Power of Dual Modes: TUI and CLI

- TUI mode provides a user-friendly way to navigate and manage files visually, ideal for manual cleanups and exploration.

- CLI mode is perfect for automation, scripting, and quick one-liners. It's essential for server environments, cron jobs, and integrating into larger toolchains.

Unlike many traditional disk usage tools that focus only on visualizing disk space (like *ncdu*, *gdu*, *dua-cli*), Deletor is optimized specifically for fast and targeted file removal.
It offers advanced filtering options by file extension, size, and custom exclusions, making it a powerful tool for real-world file management — not just analysis.


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
<a href="https://code2tutorial.com/tutorial/3aac813f-99c2-453f-819f-c80e4322e068/index.md">https://code2tutorial.com/tutorial/3aac813f-99c2-453f-819f-c80e4322e068/index.md</a>



## 📜 License
This project is distributed under the **MIT** license.

--- 
### Thank you to these wonderful people for their contributions!

<a href="https://github.com/pashkov256/deletor/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pashkov256/deletor" />
</a>
