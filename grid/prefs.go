// Copyright (c) 2021, The Grid Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grid

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/girl"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/svg"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"goki.dev/grid/icons"
)

// Preferences is the overall Grid preferences
type Preferences struct {

	// default physical size, when app is started without opening a file
	Size PhysSize

	// active color preferences
	Colors ColorPrefs

	// named color schemes -- has Light and Dark schemes by default
	ColorSchemes map[string]*ColorPrefs

	// default shape styles
	ShapeStyle girl.Paint

	// default text styles
	TextStyle girl.Paint

	// default line styles
	PathStyle girl.Paint

	// default line styles
	LineStyle girl.Paint

	// turns on the grid display
	GridDisp bool

	// snap positions and sizes to underlying grid
	SnapGrid bool

	// snap positions and sizes to line up with other elements
	SnapGuide bool

	// snap node movements to align with guides
	SnapNodes bool

	// number of screen pixels around target point (in either direction) to snap
	SnapTol int `min:"1"`

	// named-split config in use for configuring the splitters
	SplitName SplitName

	// environment variables to set for this app -- if run from the command line, standard shell environment variables are inherited, but on some OS's (Mac), they are not set when run as a gui app
	EnvVars map[string]string

	// flag that is set by StructView by virtue of changeflag tag, whenever an edit is made.  Used to drive save menus etc.
	Changed bool `view:"-" changeflag:"+" json:"-" xml:"-"`
}

var KiT_Preferences = kit.Types.AddType(&Preferences{}, PreferencesProps)

func (pf *Preferences) Defaults() {
	pf.Size.Defaults()
	pf.Colors.Defaults()
	pf.ColorSchemes = DefaultColorSchemes()
	pf.ShapeStyle.Defaults()
	pf.ShapeStyle.FontStyle.Family = "Arial"
	pf.ShapeStyle.FontStyle.Size.Set(12, units.Px)
	pf.ShapeStyle.FillStyle.Color.SetName("blue")
	pf.ShapeStyle.StrokeStyle.On = true
	pf.ShapeStyle.FillStyle.On = true
	pf.TextStyle.Defaults()
	pf.TextStyle.FontStyle.Family = "Arial"
	pf.TextStyle.FontStyle.Size.Set(12, units.Px)
	pf.TextStyle.StrokeStyle.On = false
	pf.TextStyle.FillStyle.On = true
	pf.PathStyle.Defaults()
	pf.PathStyle.FontStyle.Family = "Arial"
	pf.PathStyle.FontStyle.Size.Set(12, units.Px)
	pf.PathStyle.StrokeStyle.On = true
	pf.PathStyle.FillStyle.On = false
	pf.LineStyle.Defaults()
	pf.LineStyle.FontStyle.Family = "Arial"
	pf.LineStyle.FontStyle.Size.Set(12, units.Px)
	pf.LineStyle.StrokeStyle.On = true
	pf.LineStyle.FillStyle.On = false
	pf.GridDisp = true
	pf.SnapTol = 3
	pf.SnapGrid = true
	pf.SnapGuide = true
	pf.SnapNodes = true
	home := gi.Prefs.User.HomeDir
	pf.EnvVars = map[string]string{
		"PATH": home + "/bin:" + home + "/go/bin:/usr/local/bin:/opt/homebrew/bin:/opt/homebrew/shbin:/Library/TeX/texbin:/usr/bin:/bin:/usr/sbin:/sbin",
	}
}

func (pf *Preferences) Update() {
	pf.Size.Update()
}

// Prefs are the overall Grid preferences
var Prefs = Preferences{}

// InitPrefs must be called at startup in mainrun()
func InitPrefs() {
	Prefs.Defaults()
	Prefs.Open()
	OpenPaths()
	svg.CurIconSet.OpenIconsFromEmbedDir(icons.Icons, ".")
	gi.CustomAppMenuFunc = func(m *gi.Menu, win *gi.Window) {
		m.InsertActionAfter("GoGi Preferences...", gi.ActOpts{Label: "Grid Preferences..."},
			win, func(recv, send ki.Ki, sig int64, data any) {
				PrefsView(&Prefs)
			})
	}
}

// PrefsFileName is the name of the preferences file in GoGi prefs directory
var PrefsFileName = "grid_prefs.json"

// Open preferences from GoGi standard prefs directory, and applies them
func (pf *Preferences) Open() error {
	pdir := oswin.TheApp.AppDataDir()
	pnm := filepath.Join(pdir, PrefsFileName)
	b, err := ioutil.ReadFile(pnm)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, pf)
	AvailSplits.OpenPrefs()
	pf.ApplyEnvVars()
	pf.Changed = false
	return err
}

// Save Preferences to GoGi standard prefs directory
func (pf *Preferences) Save() error {
	pdir := oswin.TheApp.AppDataDir()
	pnm := filepath.Join(pdir, PrefsFileName)
	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		log.Println(err)
		return err
	}
	err = ioutil.WriteFile(pnm, b, 0644)
	if err != nil {
		log.Println(err)
	}
	AvailSplits.SavePrefs()
	pf.Changed = false
	return err
}

// ApplyEnvVars applies environment variables set in EnvVars
func (pf *Preferences) ApplyEnvVars() {
	for k, v := range pf.EnvVars {
		os.Setenv(k, v)
	}
}

// LightMode sets colors to light mode
func (pf *Preferences) LightMode() {
	lc, ok := pf.ColorSchemes["Light"]
	if !ok {
		log.Printf("Light ColorScheme not found\n")
		return
	}
	pf.Colors = *lc
	pf.Save()
	pf.UpdateAll()
}

// DarkMode sets colors to dark mode
func (pf *Preferences) DarkMode() {
	lc, ok := pf.ColorSchemes["Dark"]
	if !ok {
		log.Printf("Dark ColorScheme not found\n")
		return
	}
	pf.Colors = *lc
	pf.Save()
	pf.UpdateAll()
}

// EditSplits opens the SplitsView editor to customize saved splitter settings
func (pf *Preferences) EditSplits() {
	SplitsView(&AvailSplits)
}

// VersionInfo returns Grid version information
func (pf *Preferences) VersionInfo() string {
	vinfo := Version + " date: " + VersionDate + " UTC; git commit-1: " + GitCommit
	return vinfo
}

// UpdateAll updates all open windows with current preferences -- triggers
// rebuild of default styles.
func (pf *Preferences) UpdateAll() {
	gist.RebuildDefaultStyles = true
	gist.ColorSpecCache = nil
	gist.StyleTemplates = nil
	// for _, w := range gi.AllWindows {  // no need and just messes stuff up!
	// 	w.SetSize(w.OSWin.Size())
	// }
	// needs another pass through to get it right..
	for _, w := range gi.AllWindows {
		w.FullReRender()
	}
	gist.RebuildDefaultStyles = false
	// and another without rebuilding?  yep all are required
	for _, w := range gi.AllWindows {
		w.FullReRender()
	}
}

// PreferencesProps define the Toolbar and MenuBar for StructView, e.g., giv.PrefsView
var PreferencesProps = ki.Props{
	"MainMenu": ki.PropSlice{
		{"AppMenu", ki.BlankProp{}},
		{"File", ki.PropSlice{
			{"Open", ki.Props{
				"shortcut": "Command+O",
			}},
			{"Save", ki.Props{
				"shortcut": "Command+S",
				"updtfunc": giv.ActionUpdateFunc(func(pfi any, act *gi.Button) {
					pf := pfi.(*Preferences)
					act.SetActiveState(pf.Changed)
				}),
			}},
			{"sep-color", ki.BlankProp{}},
			{"LightMode", ki.Props{}},
			{"DarkMode", ki.Props{}},
			{"sep-close", ki.BlankProp{}},
			{"Close Window", ki.BlankProp{}},
		}},
		{"Edit", "Copy Cut Paste"},
		{"Window", "Windows"},
	},
	"Toolbar": ki.PropSlice{
		{"Save", ki.Props{
			"desc": "Saves current preferences to standard prefs.json file, which is auto-loaded at startup.",
			"icon": "file-save",
			"updtfunc": giv.ActionUpdateFunc(func(pfi any, act *gi.Button) {
				pf := pfi.(*Preferences)
				act.SetActiveStateUpdt(pf.Changed)
			}),
		}},
		{"sep-color", ki.BlankProp{}},
		{"LightMode", ki.Props{
			"desc": "Set color mode to Light mode as defined in ColorSchemes -- automatically does Save and UpdateAll ",
			"icon": "color",
		}},
		{"DarkMode", ki.Props{
			"desc": "Set color mode to Dark mode as defined in ColorSchemes -- automatically does Save and UpdateAll",
			"icon": "color",
		}},
		{"sep-misc", ki.BlankProp{}},
		{"VersionInfo", ki.Props{
			"desc":        "shows current Grid version information",
			"icon":        "info",
			"show-return": true,
		}},
		{"sep-key", ki.BlankProp{}},
		{"EditSplits", ki.Props{
			"icon": "file-binary",
			"desc": "opens the SplitsView editor of saved named splitter settings.  Current customized settings are saved and loaded with preferences automatically.",
		}},
	},
}

//////////////////////////////////////////////////////////////////////////////////////
//   Saved Projects / Paths

// SavedPaths is a slice of strings that are file paths
var SavedPaths gi.FilePaths

// SavedPathsFileName is the name of the saved file paths file in GoGi prefs directory
var SavedPathsFileName = "grid_saved_paths.json"

// GridViewResetRecents defines a string that is added as an item to the recents menu
var GridViewResetRecents = "<i>Reset Recents</i>"

// GridViewEditRecents defines a string that is added as an item to the recents menu
var GridViewEditRecents = "<i>Edit Recents...</i>"

// SavedPathsExtras are the reset and edit items we add to the recents menu
var SavedPathsExtras = []string{gi.MenuTextSeparator, GridViewResetRecents, GridViewEditRecents}

// SavePaths saves the active SavedPaths to prefs dir
func SavePaths() {
	gi.StringsRemoveExtras((*[]string)(&SavedPaths), SavedPathsExtras)
	pdir := oswin.TheApp.AppDataDir()
	pnm := filepath.Join(pdir, SavedPathsFileName)
	SavedPaths.SaveJSON(pnm)
	// add back after save
	gi.StringsAddExtras((*[]string)(&SavedPaths), SavedPathsExtras)
}

// OpenPaths loads the active SavedPaths from prefs dir
func OpenPaths() {
	// remove to be sure we don't have duplicate extras
	gi.StringsRemoveExtras((*[]string)(&SavedPaths), SavedPathsExtras)
	pdir := oswin.TheApp.AppDataDir()
	pnm := filepath.Join(pdir, SavedPathsFileName)
	SavedPaths.OpenJSON(pnm)
	gi.StringsAddExtras((*[]string)(&SavedPaths), SavedPathsExtras)
}

/////////////////////////////////////////////////////////////////////////////////
//   ColorPrefs

// ColorPrefs for
type ColorPrefs struct {

	// drawing background color
	Background gist.Color

	// border color of the drawing
	Border gist.Color

	// grid line color
	Grid gist.Color
}

var KiT_ColorPrefs = kit.Types.AddType(&ColorPrefs{}, ColorPrefsProps)

func (pf *ColorPrefs) Defaults() {
	pf.Background = gist.White
	pf.Border = gist.Black
	pf.Grid.SetUInt8(220, 220, 220, 255)
}

func (pf *ColorPrefs) DarkDefaults() {
	pf.Background = gist.Black
	pf.Border.SetUInt8(102, 102, 102, 255)
	pf.Grid.SetUInt8(40, 40, 40, 255)
}

func DefaultColorSchemes() map[string]*ColorPrefs {
	cs := map[string]*ColorPrefs{}
	lc := &ColorPrefs{}
	lc.Defaults()
	cs["Light"] = lc
	dc := &ColorPrefs{}
	dc.DarkDefaults()
	cs["Dark"] = dc
	return cs
}

// OpenJSON opens colors from a JSON-formatted file.
func (pf *ColorPrefs) OpenJSON(filename gi.FileName) error {
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, gi.AddOk, gi.NoCancel, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pf)
}

// SaveJSON saves colors to a JSON-formatted file.
func (pf *ColorPrefs) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, gi.AddOk, gi.NoCancel, nil, nil)
		log.Println(err)
	}
	return err
}

// SetToPrefs sets this color scheme as the current active setting in overall
// default prefs.
func (pf *ColorPrefs) SetToPrefs() {
	Prefs.Colors = *pf
	Prefs.UpdateAll()
}

// ColorPrefsProps defines the Toolbar
var ColorPrefsProps = ki.Props{
	"Toolbar": ki.PropSlice{
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"icon":  "file-open",
			"desc":  "open set of colors from a json-formatted file",
			"Args": ki.PropSlice{
				{"Color File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "Saves colors to JSON formatted file.",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"Color File Name", ki.Props{
					"ext": ".json",
				}},
			},
		}},
		{"SetToPrefs", ki.Props{
			"desc": "Sets this color scheme as the current active color scheme in Prefs.",
			"icon": "reset",
		}},
	},
}
