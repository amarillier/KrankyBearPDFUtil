package main

import (
	updatechecker "github.com/amarillier/go-update-checker"
)

// checkForUpdates checks GitHub for new releases
// Returns: (message string, updateAvailable bool)
func checkForUpdates() (string, bool) {
	uc := updatechecker.New(
		"amarillier",
		"KrankyBearPDF",
		"Kranky Bear PDF Utility",
		"https://github.com/amarillier/KrankyBearPDF/releases/latest",
		0,    // verbosity: 0 = quiet
		false, // debug: false
	)
	uc.CheckForUpdate(appVersion)
	return uc.Message, uc.UpdateAvailable
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942

