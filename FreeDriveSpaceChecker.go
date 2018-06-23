package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"./config"
	"./disk"
	"./mp3"
	"github.com/VividCortex/ewma"

	"log"
)

// TODO
// immer im vordergrund
// Prozentanzeige Toggle
// beim starten die Limits überprüfen und aktivieren dessen Limit noch nicht erreicht ist

type MyMainWindow struct {
	*walk.MainWindow
	prevFilePath string
}

func main() {
	Config := config.Init()

	var zeit *walk.Label
	var a *walk.Label
	var b *walk.Label

	var durationH *walk.Label
	var durationM *walk.Label

	// https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average
	e := ewma.NewMovingAverage() //=> Returns a SimpleEWMA if called without params

	var change uint64
	var changeArr []uint64

	var mi *walk.Action

	drive := Config.Drive

	var lastspace uint64

	t1 := time.Now()
	go func() {
		ticker := time.NewTicker(time.Second)

		out := time.Time{}
		for {
			select {
			case <-ticker.C:
				total, free, err := disk.Space(drive)
				if err != nil {
					panic(err)
				}

				if Config.Alarm1.Notification && Config.Alarm1.Value > float64(free)/MB {
					Config.Alarm1.Notification = false
					if Config.Alarm1.Sound != "" {
						go mp3.Run(Config.Alarm1.Sound)
					}
					go config.SaveConfigFile(Config)
				}
				if Config.Alarm2.Notification && Config.Alarm2.Value > float64(free)/MB {
					Config.Alarm2.Notification = false
					if Config.Alarm2.Sound != "" {
						go mp3.Run(Config.Alarm2.Sound)
					}
					go config.SaveConfigFile(Config)
				}
				if Config.Alarm3.Notification && Config.Alarm3.Value > float64(free)/MB {
					Config.Alarm3.Notification = false
					if Config.Alarm3.Sound != "" {
						go mp3.Run(Config.Alarm3.Sound)
					}
					go config.SaveConfigFile(Config)
				}

				if lastspace < free {
					change = 0
				} else {
					change = lastspace - free
				}

				e.Add(float64(change))

				if len(changeArr) > 4 {
					changeArr = changeArr[len(changeArr)-4 : len(changeArr)]
				}

				durch := e.Value()

				// fmt.Println((total - free) / total * 100) // 88 % sind schon belegt
				// fmt.Println(free / total * 100) // 12 % sind noch Frei

				a.Synchronize(func() {
					b.SetText("[" + drive + "] Frei " + formatSize(float64(free)) + " von " + formatSize(float64(total)))

					if durchschnitt(changeArr) == 0 || durch == 0 { // Keine Aufnahme
						a.SetTextColor(walk.RGB(247, 72, 67)) // rot
						a.SetText("Schreibgeschwindigkeit: 0 B/s")

						durationH.SetText("--")
						durationM.SetText("--")
						zeit.SetText("letzte Aufnahme war: " + out.Format("15:04:05"))

						t1 = time.Now()
					} else { // Bewegung auf der Platte
						diff := time.Now().Sub(t1)
						out = time.Time{}.Add(diff)
						zeit.SetText("Aufnahme läuft seit: " + out.Format("15:04:05"))

						a.SetTextColor(walk.RGB(0, 0, 0))
						a.SetText("Schreibgeschwindigkeit: " + formatSize(durch) + "/s")

						// Ich runde die verbleibende Zeit damit sie nicht noch mehr schwankt als sowie so schon
						switch {
						case durch > MB:
							durch = Round(durch, 5)
						}

						Dura := time.Duration(float64(total)/durch) * time.Second
						DuraH, DuraM := fmtDuration(Dura)

						durationH.SetText(DuraH)
						durationM.SetText(DuraM)
					}

					lastspace = free
				})
			}
		}
	}()

	var MenuItems []MenuItem
	for _, AlleLaufwerke := range disk.Getdrives() {
		Laufwerk := AlleLaufwerke
		MenuItems = append(MenuItems,
			Action{
				AssignTo: &mi,
				Text:     "&" + Laufwerk + ":/",
				OnTriggered: func() {
					drive = Laufwerk + ":/"
					e = ewma.NewMovingAverage() // Reset
					lastspace = 0

					Config.Drive = drive
					go config.SaveConfigFile(Config)
				},
			},
		)
	}

	mw := new(MyMainWindow)

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "FreeDriveSpaceChecker",
		MinSize:  Size{600, 200},
		// FixedSize: true,
		MenuItems: []MenuItem{
			Menu{
				Text: "&Menu",
				Items: []MenuItem{

					Menu{
						Text:  "&Laufwerke",
						Items: MenuItems,
					},

					Action{
						AssignTo: &mi,
						Text:     "&Einstellungen",
						OnTriggered: func() {
							if cmd, err := RunDialog(mw, mw, &Config); err != nil {
								log.Print(err)
							} else if cmd == walk.DlgCmdOK {
								config.SaveConfigFile(Config)
							}
						},
					},

					Action{
						Text:        "About me",
						OnTriggered: mw.aboutActionTriggered,
					},
				},
			},
		},
		Layout: HBox{},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					Label{
						Text:     "Aufnahme läuft seit: 00:00:00",
						AssignTo: &zeit,
						Font:     Font{Family: "Segoe UI", PointSize: 10},
					},
					Label{
						Text:     "[" + Config.Drive + "] Frei 0 B/s von 0 B/s",
						AssignTo: &b,
						Font:     Font{Family: "Segoe UI", PointSize: 10},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text:        "--",
								AssignTo:    &durationH,
								Font:        Font{Family: "Segoe UI", PointSize: 20},
								ToolTipText: "verbleibende Stunden",
							},
							Label{
								Text:        "H",
								Font:        Font{Family: "Segoe UI", PointSize: 16},
								ToolTipText: "verbleibende Stunden",
							},
							Label{
								Text:        "--",
								AssignTo:    &durationM,
								Font:        Font{Family: "Segoe UI", PointSize: 20},
								ToolTipText: "verbleibende Minuten",
							},
							Label{
								Text:        "M",
								Font:        Font{Family: "Segoe UI", PointSize: 16},
								ToolTipText: "verbleibende Minuten",
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text:     "Schreibgeschwindigkeit: 0.00B/s",
								AssignTo: &a,
								Font:     Font{Family: "Segoe UI", PointSize: 10},
							},
						},
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	// TODO wenn mehrere Fehlschlagen in einer MsgBox zusammenfassen
	if Config.Alarm1.Notification && Config.Alarm1.Sound != "" {
		found, e := config.Exists(Config.Alarm1.Sound)
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}
		if !found {
			walk.MsgBox(mw, "Audio Datei", "Audio1 Datei nicht gefunden bitte in den Einstellungen ändern:\n"+Config.Alarm1.Sound, walk.MsgBoxIconError)
		}
	}
	if Config.Alarm2.Notification && Config.Alarm2.Sound != "" {
		found, e := config.Exists(Config.Alarm2.Sound)
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}
		if !found {
			walk.MsgBox(mw, "Audio Datei", "Audio2 Datei nicht gefunden bitte in den Einstellungen ändern:\n"+Config.Alarm1.Sound, walk.MsgBoxIconError)
		}
	}
	if Config.Alarm3.Notification && Config.Alarm3.Sound != "" {
		found, e := config.Exists(Config.Alarm3.Sound)
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}
		if !found {
			walk.MsgBox(mw, "Audio Datei", "Audio3 Datei nicht gefunden bitte in den Einstellungen ändern:\n"+Config.Alarm1.Sound, walk.MsgBoxIconError)
		}
	}

	mw.Run() // https://github.com/lxn/walk/issues/103#issuecomment-278243090
}

func (mw *MyMainWindow) aboutActionTriggered() {
	walk.MsgBox(mw, "About",
		`by spddl
das lernst du auch noch ;)

github.com/spddl
steamcommunity.com/id/spddl
discord spddl#2495`,
		walk.MsgBoxIconInformation)
}

func RunDialog(owner walk.Form, mw *MyMainWindow, config *config.Config) (int, error) {
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var acceptPB, cancelPB *walk.PushButton

	var audio1label, audio2label, audio3label *walk.PushButton

	return Dialog{
		AssignTo:      &dlg,
		Title:         "Einstellungen",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     config,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		MinSize:   Size{300, 100},
		FixedSize: true,
		Layout:    VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 3},
				// StretchFactor: 4,
				Children: []Widget{
					CheckBox{
						Name:    "Alarm1Notification",
						Text:    "Aktiv",
						Checked: Bind("Alarm1.Notification"),
					},
					Label{
						Text: "Audio1:",
					},
					PushButton{
						Enabled:  Bind("Alarm1Notification.Checked"),
						AssignTo: &audio1label,
						Text:     "<none>", // nichts ausgewählt
						OnClicked: func() {
							path, err := mw.openAudio()
							if err != nil {
								panic(err)
							}
							if len(path) != 0 {
								_, file := filepath.Split(path)
								go func() {
									err := mp3.Run(path)
									if err != nil {
										walk.MsgBox(mw, "FEHLER", "Datei konnte nicht geladen werden\nERR: "+err.Error(), walk.MsgBoxIconError)
										audio1label.SetText("<none>")
										config.Alarm1.Sound = ""
									} else {
										audio1label.SetText(file)
										config.Alarm1.Sound = path
									}
								}()
							}
						},
						OnBoundsChanged: func() {
							if len(config.Alarm1.Sound) != 0 {
								_, file := filepath.Split(config.Alarm1.Sound)
								audio1label.SetText(file)
							}
						},
					},

					VSpacer{},
					Label{
						Text:        "Limit:",
						ToolTipText: "Wenn der Freier platz erreicht ist wird Alarm 1 abgespielt",
					},
					NumberEdit{
						Enabled:     Bind("Alarm1Notification.Checked"),
						Value:       Bind("Alarm1.Value"),
						Suffix:      " MB",
						MinValue:    0,
						MaxValue:    math.Inf(+1),
						ToolTipText: "Wenn das Limit erreicht ist wird Alarm 1 abgespielt",
					},

					VSpacer{
						ColumnSpan: 3,
						Size:       8,
					},

					CheckBox{
						Name:    "Alarm2Notification",
						Text:    "Aktiv",
						Checked: Bind("Alarm2.Notification"),
					},
					Label{
						Text: "Audio2:",
					},
					PushButton{
						Enabled:  Bind("Alarm2Notification.Checked"),
						AssignTo: &audio2label,
						Text:     "<none>",
						OnClicked: func() {
							path, err := mw.openAudio()
							if err != nil {
								panic(err)
							}
							if len(path) != 0 {
								_, file := filepath.Split(path)
								go func() {
									err := mp3.Run(path)
									if err != nil {
										walk.MsgBox(mw, "FEHLER", "Datei konnte nicht geladen werden\nERR: "+err.Error(), walk.MsgBoxIconError)
										audio2label.SetText("<none>")
										config.Alarm2.Sound = ""
									} else {
										audio2label.SetText(file)
										config.Alarm2.Sound = path
									}
								}()
							}
						},
						OnBoundsChanged: func() {
							if len(config.Alarm2.Sound) != 0 {
								_, file := filepath.Split(config.Alarm2.Sound)
								audio2label.SetText(file)
							}
						},
					},

					VSpacer{},
					Label{
						Text: "Limit:",
					},
					NumberEdit{
						Enabled:     Bind("Alarm2Notification.Checked"),
						Value:       Bind("Alarm2.Value"),
						Suffix:      " MB",
						MinValue:    0,
						MaxValue:    math.Inf(+1),
						ToolTipText: "Wieviel Freier platz noch zur verfügung steht, bis Alarm 2 abgespielt wird",
					},

					VSpacer{
						ColumnSpan: 3,
						Size:       8,
					},

					CheckBox{
						Name:    "Alarm3Notification",
						Text:    "Aktiv",
						Checked: Bind("Alarm3.Notification"),
					},
					Label{
						Text: "Audio3:",
					},
					PushButton{
						Enabled:  Bind("Alarm3Notification.Checked"),
						AssignTo: &audio3label,
						Text:     "<none>",
						OnClicked: func() {
							path, err := mw.openAudio()
							if err != nil {
								panic(err)
							}
							if len(path) != 0 {
								_, file := filepath.Split(path)
								go func() {
									err := mp3.Run(path)
									if err != nil {
										walk.MsgBox(mw, "FEHLER", "Datei konnte nicht geladen werden\nERR: "+err.Error(), walk.MsgBoxIconError)
										audio3label.SetText("<none>")
										config.Alarm3.Sound = ""
									} else {
										audio3label.SetText(file)
										config.Alarm3.Sound = path
									}
								}()
							}
						},
						OnBoundsChanged: func() {
							if len(config.Alarm3.Sound) != 0 {
								_, file := filepath.Split(config.Alarm3.Sound)
								audio3label.SetText(file)
							}
						},
					},
					VSpacer{},
					Label{
						Text:        "Limit:",
						ToolTipText: "Wenn das Limit erreicht ist wird Alarm 3 abgespielt",
					},
					NumberEdit{
						Enabled:     Bind("Alarm3Notification.Checked"),
						Value:       Bind("Alarm3.Value"),
						Suffix:      " MB",
						MinValue:    0,
						MaxValue:    math.Inf(+1),
						ToolTipText: "Wenn das Limit erreicht ist wird Alarm 3 abgespielt",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								log.Println(err)
								return
							}
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Abbrechen",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(owner)
}

func (mw *MyMainWindow) openAudio() (FilePath string, err error) {
	dlg := new(walk.FileDialog)

	dlg.Title = "Wähle eine Audio Datei"
	dlg.FilePath = mw.prevFilePath

	dlg.Filter = "Audio Datei (*.mp3;*.*)|*.mp3;*.*"

	if ok, err := dlg.ShowOpen(mw); err != nil {
		return "", err
	} else if !ok {
		return "", nil
	}
	mw.prevFilePath = dlg.FilePath

	return dlg.FilePath, nil
}

func Run(mw *MyMainWindow, icon *walk.Icon) {
	// We load our icon from a file.
	icon, err := walk.Resources.Icon("icon.ico")
	if err != nil {
		panic(err)
	}

	// Create the notify icon and make sure we clean it up on exit.
	ni, err := walk.NewNotifyIcon()
	if err != nil {
		panic(err)
	}
	defer ni.Dispose()

	// Set the icon and a tool tip text.
	if err := ni.SetIcon(icon); err != nil {
		panic(err)
	}
	if err := ni.SetToolTip("1 Click for info or use the context menu to exit."); err != nil {
		panic(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		fmt.Println("ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {")
		if button != walk.LeftButton {
			return
		}

		if err := ni.ShowCustom(
			"2 Walk NotifyIcon Example",
			"2 There are multiple ShowX methods sporting different icons."); err != nil {
			panic(err)
		}
	})

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		panic(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		panic(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := ni.SetVisible(true); err != nil {
		panic(err)
	}

	// Now that the icon is visible, we can bring up an info balloon.
	if err := ni.ShowInfo("3 Walk NotifyIcon Example", "3 Click the icon to show again."); err != nil {
		panic(err)
	}
}
