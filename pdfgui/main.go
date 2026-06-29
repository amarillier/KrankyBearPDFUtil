package main

import (
	"strconv"
	"time"

	"pdfutil/pdfgui/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

const (
	appName    = "Kranky Bear PDF Utility"
	appVersion = "0.3.1"
	appAuthor  = "Allan Marillier"
	appID      = "com.krankybear.pdfutil"
)

var appCopyright = "Copyright © Allan Marillier, 2024-" + strconv.Itoa(time.Now().Year())

// ---- Window size persistence (size only; Fyne can't restore position) ----
const (
	prefWindowWidth  = "windowWidth"
	prefWindowHeight = "windowHeight"

	mainWindowDefaultWidth  = float32(1100)
	mainWindowDefaultHeight = float32(850)

	// Sanity bounds: below the floor is unusably small; above the ceiling
	// suggests corrupted prefs or a multi-monitor edge case we won't honor.
	mainWindowMinWidth  = float32(500)
	mainWindowMinHeight = float32(400)
	mainWindowMaxWidth  = float32(8000)
	mainWindowMaxHeight = float32(8000)
)

// mainWindowLaunchSize returns the saved window size if present and sane, else the default.
func mainWindowLaunchSize(a fyne.App) fyne.Size {
	sw := float32(a.Preferences().FloatWithFallback(prefWindowWidth, float64(mainWindowDefaultWidth)))
	sh := float32(a.Preferences().FloatWithFallback(prefWindowHeight, float64(mainWindowDefaultHeight)))
	if sw < mainWindowMinWidth || sh < mainWindowMinHeight ||
		sw > mainWindowMaxWidth || sh > mainWindowMaxHeight {
		return fyne.NewSize(mainWindowDefaultWidth, mainWindowDefaultHeight)
	}
	return fyne.NewSize(sw, sh)
}

// saveMainWindowGeometry persists the current window size. Skips a too-small size
// (window minimized/hidden) so we don't save a useless geometry.
func saveMainWindowGeometry(a fyne.App, w fyne.Window) {
	if w == nil {
		return
	}
	sz := w.Canvas().Size()
	if sz.Width < mainWindowMinWidth || sz.Height < mainWindowMinHeight {
		return
	}
	a.Preferences().SetFloat(prefWindowWidth, float64(sz.Width))
	a.Preferences().SetFloat(prefWindowHeight, float64(sz.Height))
}

// quitApp saves the window geometry before quitting.
func quitApp(a fyne.App, w fyne.Window) {
	saveMainWindowGeometry(a, w)
	a.Quit()
}

func main() {
	myApp := app.NewWithID(appID)
	myWindow := myApp.NewWindow(appName + " v" + appVersion)

	// Load saved theme preference
	loadTheme(myApp)

	// Set window icon
	myWindow.SetIcon(resourceKrankyBearBeanieMultiColorPng)

	// Create shared file state and properties state
	fileState := ui.NewFileState()
	propsState := &ui.PropertiesState{}

	// Create dashboard with sidebar navigation
	content := ui.CreateDashboard(myWindow, fileState, propsState)

	myWindow.SetContent(content)
	myWindow.Resize(mainWindowLaunchSize(myApp)) // restore last saved size (or default)
	myWindow.CenterOnScreen()

	// Set as master window - ensures app quits when window closes
	myWindow.SetMaster()

	// Persist window size when the window is closed via its close button.
	myWindow.SetCloseIntercept(func() {
		quitApp(myApp, myWindow)
	})

	// Create main menu bar
	createMenus(myApp, myWindow)

	// Setup system tray if desktop
	if desk, ok := myApp.(desktop.App); ok {
		setupSystemTray(desk, myApp, myWindow)
	}

	// Optional: Check for updates on startup (disabled by default, enable if desired)
	// go func() {
	// 	time.Sleep(2 * time.Second) // Wait for window to show
	// 	message, available := checkForUpdates()
	// 	if available {
	// 		showUpdateDialog(myApp, message, available)
	// 	}
	// }()

	myWindow.ShowAndRun()
}

func createMenus(a fyne.App, w fyne.Window) {
	// File menu
	quitItem := fyne.NewMenuItem("Quit", func() {
		quitApp(a, w)
	})
	showItem := fyne.NewMenuItem("Show", func() {
		w.Show()
		w.RequestFocus()
	})
	hideItem := fyne.NewMenuItem("Hide", func() {
		w.Hide()
	})
	fileMenu := fyne.NewMenu("Operations", showItem, hideItem, quitItem)

	// Help menu
	aboutItem := fyne.NewMenuItem("About", func() {
		showAbout(a)
	})
	helpItem := fyne.NewMenuItem("Help", func() {
		showHelp(a)
	})
	updateItem := fyne.NewMenuItem("Check for Updates", func() {
		message, available := checkForUpdates()
		showUpdateDialog(a, message, available)
	})
	helpMenu := fyne.NewMenu("Help", aboutItem, updateItem, helpItem)

	// Settings menu
	lightThemeItem := fyne.NewMenuItem("Light Theme", func() {
		setLightTheme(a)
	})
	darkThemeItem := fyne.NewMenuItem("Dark Theme", func() {
		setDarkTheme(a)
	})
	systemThemeItem := fyne.NewMenuItem("System Theme", func() {
		setSystemTheme(a)
	})
	settingsMenu := fyne.NewMenu("Settings", lightThemeItem, darkThemeItem, systemThemeItem)

	// Main menu
	mainMenu := fyne.NewMainMenu(fileMenu, helpMenu, settingsMenu)
	w.SetMainMenu(mainMenu)
}

func setupSystemTray(desk desktop.App, a fyne.App, w fyne.Window) {
	// System tray menu
	showItem := fyne.NewMenuItem("Show", func() {
		w.Show()
		w.RequestFocus()
	})
	hideItem := fyne.NewMenuItem("Hide", func() {
		w.Hide()
	})
	aboutItem := fyne.NewMenuItem("About", func() {
		showAbout(a)
	})
	helpItem := fyne.NewMenuItem("Help", func() {
		showHelp(a)
	})
	updateItem := fyne.NewMenuItem("Check for Updates", func() {
		message, available := checkForUpdates()
		showUpdateDialog(a, message, available)
	})
	// Settings menu
	lightThemeItem := fyne.NewMenuItem("Light Theme", func() {
		setLightTheme(a)
	})
	darkThemeItem := fyne.NewMenuItem("Dark Theme", func() {
		setDarkTheme(a)
	})
	systemThemeItem := fyne.NewMenuItem("System Theme", func() {
		setSystemTheme(a)
	})
	quitItem := fyne.NewMenuItem("Quit", func() {
		quitApp(a, w)
	})

	menu := fyne.NewMenu(appName,
		showItem,
		hideItem,
		fyne.NewMenuItemSeparator(),
		aboutItem,
		helpItem,
		updateItem,
		fyne.NewMenuItemSeparator(),
		lightThemeItem,
		darkThemeItem,
		systemThemeItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)

	desk.SetSystemTrayMenu(menu)
	desk.SetSystemTrayIcon(resourceKrankyBearBeanieMultiColorPng)
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
