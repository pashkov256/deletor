

# [Deletor](https://pashkov256.github.io/deletor-doc) - interactive and CLI file deletion tool
[![Go Reference](https://pkg.go.dev/badge/github.com/pashkov256/deletor/v1.svg)](https://pkg.go.dev/github.com/pashkov256/deletor)
![Open Issues](https://img.shields.io/github/issues/pashkov256/deletor)
![License](https://img.shields.io/badge/license-MIT-blue)


**Deletor** is a handy file deletion tool that combines a powerful text interface (**TUI**) with visual directory navigation, and classic command line mode (**CLI**). It allows you to quickly find and delete files by extension and size, both interactively and through scripts.

## Features
- 🖥️ **Interactive TUI**: Modern text-based user interface for easy file navigation and management
- 🗑️ **Delete by Extension**: Deletes files with specified extensions (e.g., .mp4, .zip)
- 📏 **Size Filter**: Deletes only files larger than the specified size (e.g., 10mb, 1gb)
- 📂 **Directory Navigation**: Easy navigation through directories with arrow keys
- 🎯 **Quick Selection**: Select and delete files with keyboard shortcuts
- ⚙️ **Customizable Options**: Toggle hidden files and confirmation prompts
- 🛠️ **Confirmation Prompt**: Optional confirmation before deleting files
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

deletor -e mp4,zip -d ~/Downloads/ -s 10mb
```
### Arguments:
`-e, --extensions` — comma-separated list of extensions (for example, mp4,zip,jpg).

`-d, --directory` — the path to the file search directory.

`-s, --size` — minimum file size to delete (for example, 10 kb, 1mb, 1gb).

## Web docs
[https://pashkov256.github.io/deletor-doc](https://pashkov256.github.io/deletor-doc)



## 🛠 Contribute
[CONTRIBUTING.md](https://github.com/pashkov256/deletor/blob/main/CONTRIBUTING.md)

## 📜 License
This project is distributed under the MIT license.


