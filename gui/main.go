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
	appVersion = "0.3.0"
	appAuthor  = "Allan Marillier"
	appID      = "com.krankybear.pdfutil"
)

var appCopyright = "Copyright © Allan Marillier, 2024-" + strconv.Itoa(time.Now().Year())

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
	myWindow.Resize(fyne.NewSize(1100, 700))
	myWindow.CenterOnScreen()

	// Set as master window - ensures app quits when window closes
	myWindow.SetMaster()

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
		a.Quit()
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
		a.Quit()
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
