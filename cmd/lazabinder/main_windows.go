//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"lazabinder/internal/bundle"
	"lazabinder/internal/icons"
	"lazabinder/internal/winui"
	"os"
	"path/filepath"
	"strings"
)

//go:embed launcher.exe
var launcher []byte

//go:embed launcher-admin.exe
var adminLauncher []byte

const exeFilter = "Windows programs (*.exe)\x00*.exe\x00\x00"
const iconFilter = "Windows icons (*.ico)\x00*.ico\x00\x00"
const resourceFilter = "Windows programs and libraries\x00*.exe;*.dll\x00\x00"

func main() {
	w, err := winui.New("LazaBinder | bundle programs and manage icons", "LazaBinderNativeWindow", 880, 620)
	if err != nil {
		winui.Fatal("LazaBinder", err)
		return
	}
	w.Grip(100)
	w.Frame()
	w.Title("LazaBinder", 32, 24, 430, 40)
	w.Quiet("Two programs. One EXE. Your icon.", 33, 66, 760, 22)
	w.Strip(32, 104, 816, 44)
	var nav []*winui.Control
	for i, label := range []string{"01  Bundle programs", "02  Set bundle icon", "03  Extract icons"} {
		nav = append(nav, w.Section(label, i, 36+i*271, 108, 266, 36))
	}
	w.Rule(32, 572, 816)
	status := w.Quiet("Ready. All work stays on this computer.", 32, 582, 816, 24)
	var jobs []*winui.Control
	busy := false
	run := func(message string, work func() error, done func()) {
		if busy {
			return
		}
		busy = true
		status.SetText(message)
		for _, c := range append(append([]*winui.Control{}, jobs...), nav...) {
			c.Enable(false)
		}
		go func() {
			e := work()
			w.Post(func() {
				busy = false
				for _, c := range append(append([]*winui.Control{}, jobs...), nav...) {
					c.Enable(true)
				}
				if e != nil {
					status.SetText("Operation failed. Your inputs were kept.")
					w.Alert("Could not complete operation", e.Error())
				} else {
					status.SetText("Done.")
					if done != nil {
						done()
					}
				}
			})
		}()
	}
	choose := func(title, filter, ext string) string {
		path, e := w.FileDialog(title, filter, ext, "", false)
		if e != nil {
			w.Alert("Choose a file", e.Error())
		}
		return path
	}
	save := func(title, filter, ext, name string) string {
		path, e := w.FileDialog(title, filter, ext, name, true)
		if e != nil {
			w.Alert("Save a file", e.Error())
		}
		return path
	}

	w.Page(0)
	w.Heading("Bundle two standalone EXEs", 32, 168, 816, 30)
	w.Label("The final combined EXE can request administrator access for both embedded programs.", 32, 202, 816, 24)
	w.Caption("Program 1", 32, 240, 110, 34)
	first := w.Path("Choose the program to start first", 146, 240, 558, 34)
	jobs = append(jobs, w.Button("Browse...", 714, 240, 134, 34, func() {
		if path := choose("Choose the first program", exeFilter, "exe"); path != "" {
			first.SetText(path)
		}
	}))
	w.Caption("Program 2", 32, 280, 110, 34)
	second := w.Path("Choose the second program", 146, 280, 558, 34)
	jobs = append(jobs, w.Button("Browse...", 714, 280, 134, 34, func() {
		if path := choose("Choose the second program", exeFilter, "exe"); path != "" {
			second.SetText(path)
		}
	}))
	w.Caption("Bundle name", 32, 320, 110, 34)
	title := w.Input("My program bundle", 146, 320, 702, 34, nil)
	w.Caption("Icon (optional)", 32, 360, 110, 34)
	icon := w.Path("The launcher's own icon is used", 146, 360, 440, 34)
	preview := w.Picture(618, 370, 64)
	jobs = append(jobs, w.Button("Choose ICO", 714, 360, 134, 34, func() {
		if path := choose("Choose the bundle icon", iconFilter, "ico"); path != "" {
			if _, e := icons.ReadICO(path); e != nil {
				w.Alert("Invalid icon", e.Error())
				return
			}
			icon.SetText(path)
			preview.ShowIcon(path)
		}
	}))
	jobs = append(jobs, w.Button("Clear", 714, 400, 134, 34, func() { icon.SetText(""); preview.ShowIcon("") }))
	requireAdmin := w.Switch("Require administrator for the combined EXE", 146, 408, 440, 26, nil)
	requireAdmin.Turn(true)
	wait := w.Switch("Start program 2 after program 1 exits", 146, 440, 440, 26, nil)
	jobs = append(jobs, requireAdmin, wait, title)
	w.Quiet("Use standalone EXEs. Nearby DLLs, folders and external runtimes are not bundled.", 32, 480, 816, 24)
	create := w.Button("Create combined EXE", 576, 520, 272, 40, nil)
	create.Lead(true)
	jobs = append(jobs, create)
	create.Do(func() {
		paths := [2]string{first.Text(), second.Text()}
		for _, p := range paths {
			if e := bundle.ValidateEXE(p); e != nil {
				w.Alert("Choose two programs", e.Error())
				return
			}
		}
		output := save("Save the combined EXE", exeFilter, "exe", "Combined.exe")
		if output == "" {
			return
		}
		bundleTitle, iconPath, sequence := title.Text(), icon.Text(), wait.On()
		options := bundle.Options{Title: bundleTitle, WaitForFirst: sequence, RequireAdmin: requireAdmin.On()}
		template := launcher
		if options.RequireAdmin {
			template = adminLauncher
		}
		var setter func(string) error
		if iconPath != "" {
			setter = func(path string) error { return icons.SetFile(path, iconPath) }
		}
		run("Building and compressing the two programs...", func() error { return bundle.Build(template, paths, options, output, setter) }, func() {
			status.SetText("Created " + output)
			w.Info("Bundle created", "Saved:\n"+output+"\n\nDouble-click the new EXE to start both programs.")
		})
	})

	w.Page(1)
	w.Heading("Give your bundle a new icon", 32, 168, 816, 30)
	w.Label("Choose a LazaBinder bundle and save a new copy with a different Windows icon.", 32, 202, 816, 24)
	w.Caption("Bundle EXE", 32, 240, 110, 34)
	bound := w.Path("Choose a bundle made by LazaBinder", 146, 240, 558, 34)
	jobs = append(jobs, w.Button("Browse...", 714, 240, 134, 34, func() {
		if path := choose("Choose a LazaBinder bundle", exeFilter, "exe"); path != "" {
			a, e := bundle.Open(path)
			if e != nil {
				w.Alert("Choose a LazaBinder bundle", e.Error())
				return
			}
			a.Close()
			bound.SetText(path)
		}
	}))
	w.Caption("New icon", 32, 280, 110, 34)
	newIcon := w.Path("Choose an ICO file", 146, 280, 558, 34)
	newPreview := w.Picture(152, 340, 64)
	jobs = append(jobs, w.Button("Choose ICO", 714, 280, 134, 34, func() {
		if path := choose("Choose the new icon", iconFilter, "ico"); path != "" {
			if _, e := icons.ReadICO(path); e != nil {
				w.Alert("Invalid icon", e.Error())
				return
			}
			newIcon.SetText(path)
			newPreview.ShowIcon(path)
		}
	}))
	w.Label("ICO files can include several sizes, so your icon stays sharp in Explorer.\nThe copy keeps both embedded programs intact.", 244, 348, 604, 48)
	w.Quiet("Extract an icon in the next section, then reuse that ICO file here.", 32, 480, 816, 24)
	retouch := w.Button("Save EXE with new icon", 576, 520, 272, 40, nil)
	retouch.Lead(true)
	jobs = append(jobs, retouch)
	retouch.Do(func() {
		source, ico := bound.Text(), newIcon.Text()
		if _, e := icons.ReadICO(ico); e != nil {
			w.Alert("Choose an icon", e.Error())
			return
		}
		output := save("Save the bundle with its new icon", exeFilter, "exe", "Combined-new-icon.exe")
		if output == "" {
			return
		}
		run("Updating the bundle icon...", func() error {
			return bundle.Retouch(source, output, func(path string) error { return icons.SetFile(path, ico) })
		}, func() {
			status.SetText("Created " + output)
			w.Info("Icon applied", "Saved:\n"+output+"\n\nBoth bundled programs were preserved.")
		})
	})

	w.Page(2)
	w.Heading("Extract icons from an EXE or DLL", 32, 168, 816, 30)
	w.Label("Reads embedded icon resources without running the selected program.", 32, 202, 816, 24)
	w.Caption("Source file", 32, 240, 110, 34)
	source := w.Path("Choose a program or a library", 146, 240, 558, 34)
	var groups []icons.Group
	var groupCombo, groupPreview, summary *winui.Control
	updatePreview := func() {
		i := groupCombo.Chosen()
		if i < 0 || i >= len(groups) {
			return
		}
		g := groups[i]
		summary.SetText(g.Label + "\nSizes: " + g.Sizes)
		temp, e := os.CreateTemp("", "lazabinder-preview-*.ico")
		if e != nil {
			return
		}
		path := temp.Name()
		_, e = temp.Write(g.Data)
		temp.Close()
		defer os.Remove(path)
		if e == nil {
			groupPreview.ShowIcon(path)
		}
	}
	jobs = append(jobs, w.Button("Load icons", 714, 240, 134, 34, func() {
		path := choose("Choose an EXE or DLL", resourceFilter, "exe")
		if path == "" {
			return
		}
		var found []icons.Group
		run("Reading embedded icon resources...", func() error { var e error; found, e = icons.Extract(path); return e }, func() {
			groups = found
			source.SetText(path)
			items := make([]string, len(groups))
			for i, g := range groups {
				items[i] = g.Label
			}
			groupCombo.SetEntries(items)
			groupCombo.Choose(0)
			updatePreview()
			status.SetText(fmt.Sprintf("Found %d icon group(s).", len(groups)))
		})
	}))
	w.Caption("Icon group", 32, 280, 110, 34)
	groupCombo = w.Choice(nil, 146, 280, 702, 34, updatePreview)
	jobs = append(jobs, groupCombo)
	groupPreview = w.Picture(152, 340, 64)
	summary = w.Label("Load a program to see its embedded icons.", 244, 348, 604, 80)
	w.Quiet("Exported ICO files keep all embedded resolutions and color depths.", 32, 480, 816, 24)
	jobs = append(jobs, w.Button("Export all groups", 292, 520, 272, 40, func() {
		if len(groups) == 0 {
			w.Alert("Load a program", "Choose an EXE or DLL with embedded icons first.")
			return
		}
		path := save("Choose a base name for the icon folder", iconFilter, "ico", "Extracted-icons.ico")
		if path == "" {
			return
		}
		folder := strings.TrimSuffix(path, filepath.Ext(path)) + "-groups"
		copyGroups := append([]icons.Group(nil), groups...)
		run("Exporting all icon groups...", func() error {
			if e := os.Mkdir(folder, 0700); e != nil {
				return e
			}
			for i, g := range copyGroups {
				if e := os.WriteFile(filepath.Join(folder, fmt.Sprintf("icon-%03d.ico", i+1)), g.Data, 0600); e != nil {
					os.RemoveAll(folder)
					return e
				}
			}
			return nil
		}, func() {
			status.SetText("Exported " + folder)
			w.Info("Icons exported", "Saved all icon groups in:\n"+folder)
		})
	}))
	export := w.Button("Export selected ICO", 576, 520, 272, 40, nil)
	export.Lead(true)
	jobs = append(jobs, export)
	export.Do(func() {
		i := groupCombo.Chosen()
		if i < 0 || i >= len(groups) {
			w.Alert("Load a program", "Choose an EXE or DLL with embedded icons first.")
			return
		}
		path := save("Save the selected icon", iconFilter, "ico", "Extracted-icon.ico")
		if path == "" {
			return
		}
		if !strings.EqualFold(filepath.Ext(path), ".ico") {
			w.Alert("Save an ICO file", "The exported filename must end in .ico.")
			return
		}
		data := groups[i].Data
		run("Exporting the selected icon...", func() error { return os.WriteFile(path, data, 0600) }, func() { status.SetText("Exported " + path) })
	})
	w.ShowPage(0)
	w.Run()
}
