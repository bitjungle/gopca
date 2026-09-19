# GoPCA Suite — Official Downloads

**Professional-grade Principal Component Analysis. Free, offline, and yours.**

This repository distributes the official binary releases of GoPCA Suite. It is
the only place these binaries are published by the author.

📥 **[Download the latest release](https://github.com/bitjungle/GoPCA-Releases/releases/latest)**
&nbsp;•&nbsp; 📖 **[Documentation](https://bitjungle.github.io/GoPCA-Releases/)**

---

## What you get

Three tools that work together, in one download:

| Tool | What it does |
|---|---|
| **GoPCA Desktop** | Interactive visual analysis and exploration |
| **GoCSV Desktop** | Data preparation with a spreadsheet-like interface |
| **`pca` CLI** | Scriptable command line for automation |

All processing happens on your machine. No cloud, no telemetry, no network
dependencies — GoPCA works fully offline.

## Download

| Platform | File | Contents |
|---|---|---|
| **macOS** (Intel + Apple Silicon) | `gopca-macos-universal.zip` | `GoPCA.app`, `GoCSV.app` (signed & notarized), `pca-intel`, `pca-arm64` |
| **Windows** (installer) | `GoPCA-Setup-vX.X.X.exe` | Guided installation |
| **Windows** (portable) | `gopca-windows-x64.zip` | `GoPCA.exe`, `GoCSV.exe`, `pca.exe` |
| **Linux** | `gopca-linux-x64.tar.gz` | `GoPCA`, `GoCSV`, `pca-x64`, `pca-arm64` |
| **Linux** (AppImage) | `GoPCA-x86_64.AppImage`, `GoCSV-x86_64.AppImage` | Standalone, all distributions |

**Windows users:** installing from the
**[Microsoft Store](https://apps.microsoft.com/detail/9n8hcxgrjzt5)** is
recommended — Microsoft signs the package, so you get no SmartScreen warnings
and automatic updates.

## Verifying your download

Every release includes `checksums.txt` with SHA-256 hashes for all artifacts.

```bash
# macOS / Linux
shasum -a 256 -c checksums.txt

# Windows (PowerShell)
Get-FileHash .\gopca-windows-x64.zip -Algorithm SHA256
```

Only download from this repository. Binaries offered anywhere else are not
published by the author and have not been verified.

## Installation notes

**macOS** — Gatekeeper may block a freshly downloaded app. Move both `GoPCA.app`
and `GoCSV.app` into your Applications folder before launching, or right-click
the app and choose **Open**. The macOS builds are signed and notarized by the
author.

**Windows** — the standalone installer and portable ZIP are **not code-signed**,
so SmartScreen will warn on first run. Click **More info → Run anyway**, or
install from the [Microsoft Store](https://apps.microsoft.com/detail/9n8hcxgrjzt5)
to avoid the warning entirely.

**Linux** — make the AppImage executable before running it:

```bash
chmod +x GoPCA-x86_64.AppImage
./GoPCA-x86_64.AppImage
```

These warnings are standard operating-system behaviour for independently
published software, not a sign that anything is wrong. Verify the checksum if in
doubt.

## Documentation

Full documentation is published at
**[bitjungle.github.io/GoPCA-Releases](https://bitjungle.github.io/GoPCA-Releases/)**:

- **[Introduction to PCA](docs/intro_to_pca.md)** — the method, from first principles
- **[Introduction to Data Preparation](docs/intro_to_data_prep.md)** — getting your data ready
- **[CLI Reference](docs/cli_reference.md)** — every `pca` command and flag
- **[Data Format Guide](docs/data-format.md)** — how GoPCA reads your CSV files
- **[Windows Installation Guide](docs/windows-installation.md)** — Store, installer, and portable

## Privacy

GoPCA Suite performs all computation locally:

- **No telemetry** — no analytics, tracking, or data collection
- **No network calls** — the applications work entirely offline
- **Your data stays put** — nothing is uploaded anywhere

This makes GoPCA suitable for use under GDPR, HIPAA, and strict corporate data
policies.

**You can check this yourself, without the source code.** Block GoPCA at your
firewall, or disconnect from the network entirely — every feature still works,
because there is nothing to connect to. See **[PRIVACY.md](PRIVACY.md)** for
packet-capture and DNS-monitoring instructions, and for the full privacy policy.

## License

The binaries in this repository are distributed free of charge under the
**[GoPCA Suite Binary Distribution License](LICENSE)**.

In short: you may use and redistribute the official binaries at no cost, provided
they are unmodified, the license travels with them, you charge nothing for the
software itself, and you do not present yourself as the author. The license also
prohibits military, weapons, intelligence, and surveillance applications. Please
read the full text — the summary above is not the license.

## Source code

The source code for GoPCA Suite is **not public**. Access may be granted at the
author's discretion and requires a written agreement.

If you need source access — for security review, research, interoperability, or
evaluation — write to **devel@bitjungle.com** describing what you need and why.

## Problems and questions

Please open an issue in
**[this repository's issue tracker](https://github.com/bitjungle/GoPCA-Releases/issues)**.
Include your platform, the GoPCA version, and what you expected to happen.

---

GoPCA Suite © 2025-2026 Rune Mathisen. All rights reserved except as granted in
the [LICENSE](LICENSE).
