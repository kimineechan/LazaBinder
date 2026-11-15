# Changelog

## 1.1.0 - 2026-11-15

A new window, in the style of LazaReplay and LazaScrub. Bundles are made exactly as before.

- The window is black and see-through, with what is behind it blurred, and has rounded buttons, switches in place of tick boxes, and its own list of icon groups, which scrolls when a program holds many.
- Keyboard use throughout: Tab and Shift + Tab move between controls with a white ring to show where you are, Enter or Space presses, and Ctrl + Tab changes section.
- The button for the thing to do next is the light one.
- The window is drawn by LazaBinder itself, only where something changed. No web view, no third-party modules, CGO still off.
- The launcher that goes into every bundle no longer carries any of the window code.

## 1.0.1 — 2026-11-09

Renamed the project from LazarusBinder to LazaBinder.

- Rename the application, Go module, build and publish scripts, documentation, GitHub metadata, and release assets to LazaBinder. No functional changes.

## 1.0.0 — 2026-10-14

First public release.

- Bundle two standalone Windows EXEs into one combined EXE.
- Start both programs together or launch the second after the first exits.
- Make the final combined EXE request administrator rights through Windows UAC.
- Launch the programs without an extra status window; exit the wrapper after startup.
- Apply a custom ICO during bundle creation or save a bundle copy with a new icon.
- Extract complete, multi-resolution icon groups from EXE/DLL resources.
- Preserve original input files and verify extracted payloads using size and SHA-256 checks.
- Include Windows build/release automation, source tests, and checksum files.

