package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// https://github.com/pdfcpu/pdfcpu
// See api docs: https://pkg.go.dev/github.com/pdfcpu/pdfcpu/pkg/api

const (
	appName    = "Kranky Bear pdfutil"
	appVersion = "0.3.0" // see FyneApp.toml
	appAuthor  = "Allan Marillier"
)

var appCopyright = "Copyright (c) Allan Marillier, 2025-" + strconv.Itoa(time.Now().Year())
var err error
var exists bool

// decrypt a pdf file using provided user password, owner password and AES key length
func decryptFile(inFile string, userPw string, ownerPw string, keyLength int) error {
	// Decrypting the file using user specified key length
	// keyLength encryptionBits actually does no matter, pass 0
	conf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
	err = api.DecryptFile(inFile, "", conf) // write output to inFile, not a new file
	return err
}

// encrypt a pdf file using provided user password, owner password and AES key length
func encryptFile(inFile string, userPw string, ownerPw string, keyLength int, mode string) error {
	// Encrypting the file using user specified key length & mode
	if mode == "AES" {
		aesConf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
		err = api.EncryptFile(inFile, "", aesConf)
	} else {
		// for RC4 encryption
		rc4Conf := model.NewRC4Configuration(userPw, ownerPw, keyLength)
		err = api.EncryptFile(inFile, "", rc4Conf)
	}
	// write output to inFile, not a new file
	return err
}

// extractPages splits a PDF file into individual pages and saves them in the
// specified directory.
func extractPages(inFile string, outDir string, pageNumbers []string, pad bool, userPass string, ownerPass string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err = api.ExtractPagesFile(inFile, outDir, pageNumbers, conf)
	if err != nil {
		log.Fatalf("Error extracting pages from PDF file: %v", err)
	}
	if pad {
		// zero pad the names for better sorting
		err = zeroPadNames(outDir, 3) // pad to 3 digits
		if err != nil {
			log.Fatalf("Error zero padding names: %v", err)
		}
	}
	return err
}

// fileInfo reads and prints basic information about a PDF file.
// It includes details like version, number of pages, title, author, etc.
func fileInfo(inFile string, userPass string, ownerPass string, encryptionBits int) {
	// Read and validate the PDF context info

	var ctx *model.Context
	// Create a configuration with the user and owner passwords when provided
	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.LISTINFO
	/*
		return &Command{
			Mode:          model.LISTINFO,
			InFiles:       inFiles,
			PageSelection: pageSelection,
			BoolVal1:      fonts,
			BoolVal2:      json,
			Conf:          conf}
	*/
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass
	encryptionBits = 0 // fake use to avoid errors for now, we don't use it

	if isEncrypted(inFile) {
		f, err := os.Open(inFile)
		if err != nil {
			log.Fatalf("Failed to read open file: %v", err)
		}
		defer f.Close()

		// Use ReadAndValidate instead of ReadContext to ensure metadata is populated
		// Validation parses the Info dictionary and populates Title, Author, etc.
		ctx, err = api.ReadAndValidate(f, conf)
		if err != nil {
			fmt.Println("The file appears to be encrypted, enter a user or owner password")
			return
			// log.Fatalf("Failed to read PDF context: %v", err)
		}
	} else {
		ctx, err = api.ReadContextFile(inFile)
	}

	// Print basic PDF info
	fmt.Println("File:", inFile)
	fmt.Printf("PDF Version: %s\n", ctx.HeaderVersion)
	fmt.Printf("Number of Pages: %d\n", ctx.PageCount)
	fmt.Printf("Title: %s\n", ctx.Title)
	fmt.Printf("Author: %s\n", ctx.Author)
	fmt.Println("Creator:", ctx.Creator)
	fmt.Println("Producer:", ctx.Producer)
	fmt.Printf("Subject: %s\n", ctx.Subject)
	fmt.Printf("Keywords: %s\n", ctx.Keywords)
	fmt.Println("Modification Date:", ctx.ModDate)
	fmt.Println("Creation Date:", ctx.Configuration.CreationDate)
	if ctx.Configuration.Permissions == model.PermissionsNone {
		fmt.Println("Permissions: None, full access")
	} else {
		fmt.Println("Permissions: ", ctx.Permissions)
	}
	fmt.Println("Version:", ctx.Configuration.Version)
	if isEncrypted(inFile) {
		fmt.Println("Encrypted: Yes")
	} else {
		fmt.Println("Encrypted: No")
	}

	//perms, err := api.GetPermissionsFile(inFile, nil)
	//if err != nil {
	//	log.Fatal("Error getting permissions")
	//}
	//fmt.Println("Permissions:", perms)
	permissionsGet(inFile, userPass, ownerPass)
}

// insertPDFBetweenPages inserts a PDF file between the pages of another PDF file.
// It splits the original PDF into two parts, inserts the specified PDF file,
// and merges them back together.
func insertPDFBetweenPages(inFile string, insertFile string, outFile string, insertAfterPage int, userPass string, ownerPass string) error {
	var pages []string
	// Split the original PDF into two parts

	part1 := makeTempDir()
	part2 := makeTempDir()
	if part1 == "" || part2 == "" { // If directory creation failed, exit
		fmt.Println("Failed to create temporary directories")
		return err
	}

	pages = []string{"1-" + fmt.Sprint(insertAfterPage)}
	err = extractPages(inFile, part1, pages, false, userPass, ownerPass)
	files1, err := readFilesInDir(part1)
	if err != nil {
		// log.Fatal(err)
		return err
	}

	pages = []string{fmt.Sprint(insertAfterPage+1) + "-l"}
	err = extractPages(inFile, part2, pages, false, userPass, ownerPass)
	files2, err := readFilesInDir(part2)
	if err != nil {
		// log.Fatal(err)
		return err
	}

	// Merge the parts with the insert PDF
	pages = files1[:]
	pages = append(pages, insertFile)
	pages = append(pages, files2...)
	err = api.MergeCreateFile(pages, outFile, false, nil)
	if err != nil {
		log.Fatalf("Error merging PDF files: %v", err)
		return err
	}

	// cleanup temp directories
	err = os.RemoveAll(part1)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", part1, err)
		return err
	}
	err = os.RemoveAll(part2)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", part2, err)
		return err
	}
	return err
}

// fast test if the file is encrypted before other actions
func isEncrypted(inFile string) bool {
	// Create a configuration with the user and owner passwords when provided
	conf := model.NewDefaultConfiguration()

	f, err := os.Open(inFile)
	if err != nil {
		log.Fatalf("Failed to read open file: %v", err)
	}
	defer f.Close()

	ctx, err := api.ReadContext(f, conf)
	if err != nil || ctx.Encrypt != nil {
		// fmt.Println(err) // err likely to be "pdfcpu: please provide the correct password"
		return true
	} else {
		return false
	}
}

// mergePDFs merges multiple PDF files into a single PDF file.
// It takes a slice of input file paths and an output file path.
func mergePDFs(inFiles []string, outFile string) error {
	err = api.MergeCreateFile(inFiles, outFile, false, nil)
	if err != nil {
		// log.Fatalf("Error merging PDF files: %v", err)
		return err
	}
	return err
}

// removePages removes specified pages from a PDF file and saves the result to a new file.
// The pages parameter should be a slice of strings in pdfcpu format, e.g. ["1", "3-5", "10-l"]
// where "l" represents the last page.
func removePages(inFile string, outFile string, pages []string) error {
	err = api.RemovePagesFile(inFile, outFile, pages, nil)
	if err != nil {
		return err
	}
	return err
}

// permissionDetails holds detailed permission information
type permissionDetails struct {
	RawValue       int16
	Print          bool
	Modify         bool
	Extract        bool
	Annotate       bool
	FillForms      bool
	ExtractRev3    bool
	Assemble       bool
	PrintHighRes   bool
	FullPermission bool
}

// decodePermissions decodes permission bits into human-readable details
func decodePermissions(perms int16) permissionDetails {
	details := permissionDetails{
		RawValue: perms,
	}

	// Permission bit values from pdfcpu model
	// Bit 3:  Print (rev.2) / Draft print (rev.3+)
	details.Print = (perms & (1 << 2)) != 0 // Bit 3 = 1<<2

	// Bit 4:  Modify contents
	details.Modify = (perms & (1 << 3)) != 0 // Bit 4 = 1<<3

	// Bit 5:  Copy/extract text & graphics
	details.Extract = (perms & (1 << 4)) != 0 // Bit 5 = 1<<4

	// Bit 6:  Add or modify annotations, fill form fields
	details.Annotate = (perms & (1 << 5)) != 0 // Bit 6 = 1<<5

	// Bit 9:  Fill form fields (rev.3+)
	details.FillForms = (perms & (1 << 8)) != 0 // Bit 9 = 1<<8

	// Bit 10: Extract text & graphics (rev.3+)
	details.ExtractRev3 = (perms & (1 << 9)) != 0 // Bit 10 = 1<<9

	// Bit 11: Assemble document (rev.3+)
	details.Assemble = (perms & (1 << 10)) != 0 // Bit 11 = 1<<10

	// Bit 12: Print high quality (rev.3+)
	details.PrintHighRes = (perms & (1 << 11)) != 0 // Bit 12 = 1<<11

	return details
}

// displayPermissions displays permissions in a user-friendly format
func displayPermissions(details permissionDetails) {
	fmt.Printf("Raw Permission Value: 0x%04X (%d)\n\n", details.RawValue, details.RawValue)

	fmt.Println("Permission Details")
	fmt.Printf("Print Document:         %s       \n", formatPermission(details.Print))
	fmt.Printf("Print High Quality:     %s       \n", formatPermission(details.PrintHighRes))
	fmt.Printf("Modify Document:        %s       \n", formatPermission(details.Modify))
	fmt.Printf("Copy/Extract Content:   %s       \n", formatPermission(details.Extract))
	fmt.Printf("Extract (Rev 3+):       %s       \n", formatPermission(details.ExtractRev3))
	fmt.Printf("Annotate/Comment:       %s       \n", formatPermission(details.Annotate))
	fmt.Printf("Fill Form Fields:       %s       \n", formatPermission(details.FillForms))
	fmt.Printf("Assemble Document:      %s       \n", formatPermission(details.Assemble))
}

// formatPermission formats a boolean permission as a colored string
func formatPermission(allowed bool) string {
	if allowed {
		return "= ALLOWED "
	}
	return "! DENIED  "
}

// permissionsGet lists permissions for the specified PDF file with detailed breakdown
func permissionsGet(inFile string, userPw string, ownerPw string) int16 {
	conf := model.NewDefaultConfiguration()
	conf.OwnerPW = ownerPw
	conf.UserPW = userPw

	perms, err := api.GetPermissionsFile(inFile, conf)
	if err != nil {
		log.Fatalf("Error getting permissions: %v", err)
	}

	if perms == nil {
		fmt.Println("\n= Document is NOT encrypted")
		fmt.Println("= Full access - all permissions granted (no restrictions)")
		return -1 // All permissions
	}

	details := decodePermissions(*perms)
	displayPermissions(details)

	return *perms
}

// parseIndividualPermission parses individual permission flag names
func parseIndividualPermission(perm string) (model.PermissionFlags, bool) {
	switch strings.TrimSpace(strings.ToLower(perm)) {
	case "printdraft", "print-draft", "print":
		return model.PermissionPrintRev2, true
	case "printhq", "print-hq", "printhighquality", "print-high-quality", "printquality":
		return model.PermissionPrintRev3, true
	case "modify", "edit":
		return model.PermissionModify, true
	case "extract", "copy", "copyextract", "copy-extract":
		return model.PermissionExtract, true
	case "annotate", "annotations", "comment", "comments":
		return model.PermissionModAnnFillForm, true
	case "fillforms", "fill-forms", "forms":
		return model.PermissionFillRev3, true
	case "assemble", "assembledoc", "assemble-doc":
		return model.PermissionAssembleRev3, true
	case "extractrev3", "extract-rev3", "extract3":
		return model.PermissionExtractRev3, true
	default:
		return 0, false
	}
}

// permissionsSet sets specific permissions on a PDF file
// The file must be encrypted for permissions to apply
// Supports both pre-defined profiles and custom combinations of permissions
func permissionsSet(inFile string, outFile string, userPass string, ownerPass string, encryptBits int, permType string) error {
	// Create a configuration with owner and user passwords
	conf := model.NewAESConfiguration(userPass, ownerPass, encryptBits)

	// Check if permType contains comma (custom combination)
	if strings.Contains(permType, ",") {
		// Parse comma-separated individual permissions
		permissions := strings.Split(permType, ",")
		conf.Permissions = model.PermissionsNone
		var permNames []string

		for _, perm := range permissions {
			flag, ok := parseIndividualPermission(perm)
			if ok {
				conf.Permissions = conf.Permissions + flag
				permNames = append(permNames, strings.TrimSpace(perm))
			} else {
				fmt.Printf("Warning: Unknown permission '%s' ignored\n", strings.TrimSpace(perm))
			}
		}

		if len(permNames) > 0 {
			fmt.Printf("Setting CUSTOM permissions: %s\n", strings.Join(permNames, ", "))
		} else {
			fmt.Println("Warning: No valid permissions specified, defaulting to NONE")
			conf.Permissions = model.PermissionsNone
		}
	} else {
		// Use pre-defined permission profiles or single permission
		switch permType {
		case "all", "full":
			conf.Permissions = model.PermissionsAll
			fmt.Println("Setting: ALL permissions (full access)")
		case "none", "restrict":
			conf.Permissions = model.PermissionsNone
			fmt.Println("Setting: NO permissions (all restricted)")
		case "print":
			conf.Permissions = model.PermissionsPrint
			fmt.Println("Setting: PRINT permissions (draft + high quality)")
		case "readonly", "read":
			// No modifications, no printing, only viewing/extracting
			conf.Permissions = model.PermissionsNone + model.PermissionExtract + model.PermissionExtractRev3
			fmt.Println("Setting: READ-ONLY permissions (view and extract only)")
		case "forms":
			// Can fill forms but not modify document
			conf.Permissions = model.PermissionsNone + model.PermissionFillRev3 + model.PermissionModAnnFillForm
			fmt.Println("Setting: FORMS permissions (fill forms only)")
		case "annotate":
			// Can add annotations but not modify document
			conf.Permissions = model.PermissionsNone + model.PermissionModAnnFillForm + model.PermissionFillRev3
			fmt.Println("Setting: ANNOTATE permissions (add comments/annotations)")
		case "modify":
			// Full modify permissions
			conf.Permissions = model.PermissionsNone + model.PermissionModify + model.PermissionModAnnFillForm +
				model.PermissionFillRev3 + model.PermissionAssembleRev3
			fmt.Println("Setting: MODIFY permissions (edit document)")
		default:
			// Try to parse as a single individual permission
			flag, ok := parseIndividualPermission(permType)
			if ok {
				conf.Permissions = model.PermissionsNone + flag
				fmt.Printf("Setting: %s permission\n", strings.ToUpper(permType))
			} else {
				// Unknown, default to print
				conf.Permissions = model.PermissionsPrint
				fmt.Printf("Unknown permission type '%s', defaulting to PRINT permissions\n", permType)
			}
		}
	}

	err := api.SetPermissionsFile(inFile, outFile, conf)
	if err != nil {
		return fmt.Errorf("error setting permissions: %v", err)
	}

	if outFile == "" {
		fmt.Printf("\nPermissions successfully set on: %s\n", inFile)
	} else {
		fmt.Printf("\nPermissions successfully set on: %s\n", outFile)
	}
	return nil
}

// rotatePages rotates specified pages of a PDF file by a given angle.
// The rotation can be 90, 180, or 270 degrees.
// It takes the input file, output file, a slice of page numbers, and the rotation
func rotatePages(inFile string, outFile string, pages []string, rotation int) error {
	if rotation != 90 && rotation != 180 && rotation != 270 {
		log.Fatalf("Invalid rotation value, must be 90, 180 or 270")
	}
	err = api.RotateFile(inFile, outFile, rotation, pages, nil)
	if err != nil {
		// log.Fatalf("Error rotating pages: %v", err)
		return err
	}
	return err
}

// reversePages reverses the order of all pages in a PDF file.
// It splits the original PDF into individual pages, reads them in reverse order,
// and merges them back together into a new PDF file. setContext specifies if
// the same file context settings (creator, permissions etc) from the source file
// should be applied to the destination file
func reversePages(inFile string, outFile string, setContext bool, userPass string, ownerPass string) error {
	var pages []string
	// Split the original PDF, read in reverse and join back together

	outDir := makeTempDir()
	// split all to temp directory
	err = splitPDF(inFile, outDir, true, userPass, ownerPass)
	if err != nil {
		return err
	}

	/* redundant since we have pad in the call to splitPDF function now
	// zero pad the names for better sorting
	err = zeroPadNames(outDir, 3) // pad to 3 digits
	if err != nil {
		log.Fatalf("Error zero padding names: %v", err)
	}
	*/

	// read all files in the temp directory
	// Note: os.ReadDir returns []os.DirEntry, which is more efficient than reading
	// file contents but we need to convert it to []string for api.MergeCreateFile
	// pages, err := readFilesInDir(outDir)
	files, err := os.ReadDir(outDir)
	if err != nil {
		log.Fatalf("Error reading directory %s: %v", outDir, err)
	}
	// Convert os.DirEntry to []string
	for _, file := range files {
		if !file.IsDir() {
			pages = append(pages, filepath.Join(outDir, file.Name()))
		}
	}

	// sort the pages slice in reverse order
	sort.Sort(sort.Reverse(sort.StringSlice(pages)))

	// merge the files in reverse order
	err = api.MergeCreateFile(pages, outFile, false, nil)
	if err != nil {
		// log.Fatalf("Error merging PDF files: %v", err)
		return err
	}

	// cleanup temp directory
	cleanupErr := os.RemoveAll(outDir)
	if cleanupErr != nil {
		fmt.Printf("Warning: Error removing temporary directory %s: %v\n", outDir, cleanupErr)
	}

	if setContext {
		// Read and set the PDF context from original if required
		ctx, err := api.ReadContextFile(inFile)
		if err != nil {
			return fmt.Errorf("failed to read PDF context: %v", err)
		}
		err = api.WriteContextFile(ctx, outFile)
		if err != nil {
			return fmt.Errorf("failed to write PDF context: %v", err)
		}
		fmt.Println("PDF context (metadata, permissions) preserved from original file")
	}
	return nil
}

// splitPDF splits a PDF file into individual pages and saves them in the
// specified directory.
// It splits the PDF into files named <filename>_pagenumber.pdf
func splitPDF(inFile string, outDir string, pad bool, userPass string, ownerPass string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err := api.SplitFile(inFile, outDir, 1, conf)
	if err != nil {
		// log.Fatalf("Error splitting PDF file: %v", err)
		return err
	}
	if pad {
		// zero pad the names for better sorting
		err = zeroPadNames(outDir, 3) // pad to 3 digits
		if err != nil {
			log.Fatalf("Error zero padding names: %v", err)
		}
	}
	return err
}

// listProperties displays all document properties of a PDF file
func listProperties(inFile string, userPass string, ownerPass string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	var ctx *model.Context
	var err error

	if isEncrypted(inFile) {
		f, err := os.Open(inFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()

		// Use ReadAndValidate instead of ReadContext to ensure metadata is populated
		// Validation parses the Info dictionary and populates Title, Author, etc.
		ctx, err = api.ReadAndValidate(f, conf)
		if err != nil {
			return fmt.Errorf("error reading encrypted PDF context: %v", err)
		}
	} else {
		ctx, err = api.ReadContextFile(inFile)
		if err != nil {
			return fmt.Errorf("error reading PDF context: %v", err)
		}
	}

	fmt.Printf("\n=== Document Properties for: %s ===\n", inFile)
	fmt.Println()

	// Basic document info
	fmt.Println("📄 Document Information:")
	fmt.Println("├─ Title:", getStringValue(ctx.Title))
	fmt.Println("├─ Author:", getStringValue(ctx.Author))
	fmt.Println("├─ Subject:", getStringValue(ctx.Subject))
	fmt.Println("├─ Keywords:", getStringValue(ctx.Keywords))
	fmt.Println("├─ Creator:", getStringValue(ctx.Creator))
	fmt.Println("├─ Producer:", getStringValue(ctx.Producer))
	fmt.Println("├─ Creation Date:", getStringValue(ctx.Configuration.CreationDate))
	fmt.Println("└─ Modification Date:", getStringValue(ctx.ModDate))
	fmt.Println()

	// PDF version
	fmt.Printf("📋 PDF Version: %s\n", ctx.HeaderVersion)
	fmt.Println()

	// Page count
	fmt.Printf("📊 Page Count: %d\n", ctx.PageCount)

	// Encryption status
	if ctx.Encrypt != nil {
		fmt.Println("🔒 Encryption: Enabled")
		fmt.Printf("├─ User Password Required: %t\n", ctx.Encrypt != nil)
		fmt.Printf("└─ Owner Password Required: %t\n", ctx.Encrypt != nil)
	} else {
		fmt.Println("🔓 Encryption: Disabled")
	}

	fmt.Println()
	return nil
}

// addProperties sets document properties for a PDF file
func addProperties(inFile string, outFile string, userPass string, ownerPass string, properties map[string]string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	// Convert property names to the correct case for pdfcpu
	pdfProperties := make(map[string]string)
	for key, value := range properties {
		switch strings.ToLower(key) {
		case "title":
			pdfProperties["Title"] = value
		case "author":
			pdfProperties["Author"] = value
		case "subject":
			pdfProperties["Subject"] = value
		case "keywords":
			pdfProperties["Keywords"] = value
		case "creator":
			pdfProperties["Creator"] = value
		case "producer":
			pdfProperties["Producer"] = value
		default:
			fmt.Printf("⚠️  Warning: Unknown property '%s' - skipping\n", key)
		}
	}

	// Use the correct pdfcpu API for adding properties
	err := api.AddPropertiesFile(inFile, outFile, pdfProperties, conf)
	if err != nil {
		return fmt.Errorf("error adding properties: %v", err)
	}

	fmt.Printf("✅ Properties successfully added to: %s\n", outFile)
	for key, value := range properties {
		fmt.Printf("   %s: %s\n", strings.Title(key), value)
	}

	return nil
}

// removeProperties removes specified document properties from a PDF file
func removeProperties(inFile string, outFile string, userPass string, ownerPass string, properties []string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	// Convert property names to the correct case for pdfcpu
	var pdfProperties []string
	for _, prop := range properties {
		switch strings.ToLower(prop) {
		case "title":
			pdfProperties = append(pdfProperties, "Title")
		case "author":
			pdfProperties = append(pdfProperties, "Author")
		case "subject":
			pdfProperties = append(pdfProperties, "Subject")
		case "keywords":
			pdfProperties = append(pdfProperties, "Keywords")
		case "creator":
			pdfProperties = append(pdfProperties, "Creator")
		case "producer":
			pdfProperties = append(pdfProperties, "Producer")
		default:
			fmt.Printf("⚠️  Warning: Unknown property '%s' - skipping\n", prop)
		}
	}

	// Use the correct pdfcpu API for removing properties
	err := api.RemovePropertiesFile(inFile, outFile, pdfProperties, conf)
	if err != nil {
		return fmt.Errorf("error removing properties: %v", err)
	}

	fmt.Printf("✅ Successfully removed properties from: %s\n", outFile)
	for _, prop := range properties {
		fmt.Printf("   Removed: %s\n", strings.Title(prop))
	}

	return nil
}

// getStringValue safely extracts string value from PDF property
func getStringValue(value string) string {
	if value == "" {
		return "(not set)"
	}
	return value
}

// changeUserPassword changes the user password of an encrypted PDF file
func changeUserPassword(inFile string, outFile string, oldPassword string, newPassword string, ownerPassword string) error {
	conf := model.NewDefaultConfiguration()
	conf.OwnerPW = ownerPassword

	err := api.ChangeUserPasswordFile(inFile, outFile, oldPassword, newPassword, conf)
	if err != nil {
		return fmt.Errorf("error changing user password: %v", err)
	}

	// Show appropriate file name for success message
	successFile := outFile
	if successFile == "" {
		successFile = inFile
	}
	fmt.Printf("✅ User password successfully changed for: %s\n", successFile)
	return nil
}

// changeOwnerPassword changes the owner password of an encrypted PDF file
func changeOwnerPassword(inFile string, outFile string, oldPassword string, newPassword string, userPassword string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPassword

	err := api.ChangeOwnerPasswordFile(inFile, outFile, oldPassword, newPassword, conf)
	if err != nil {
		return fmt.Errorf("error changing owner password: %v", err)
	}

	// Show appropriate file name for success message
	successFile := outFile
	if successFile == "" {
		successFile = inFile
	}
	fmt.Printf("✅ Owner password successfully changed for: %s\n", successFile)
	return nil
}

func main() {

	if len(os.Args) <= 1 {
		fmt.Println(appName)
		fmt.Printf("Version: %s\n", appVersion)
		fmt.Println(appCopyright)
		fmt.Println()
		fmt.Println("A simple PDF management utility for encryption, decryption, page manipulation,")
		fmt.Println("permissions management, document properties, and more.")
		fmt.Println()
		fmt.Println("For detailed usage information, run: pdfutil --help")
		fmt.Println()
		fmt.Println("Expected [checkupdate | changepassword | decrypt | encrypt | extract | fileinfo | insert | join | merge | permissions | properties | setperms | remove | reverse | rotate | split] action as first parameter")
		os.Exit(1)
	}

	action := os.Args[1]

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Println("Kranky Bear pdfutil is a simple PDF management utility")
		fmt.Println(appName, "version", appVersion, "-", appCopyright)
		fmt.Println("Examples:")
		fmt.Println("pdfutil [changepassword | decrypt | encrypt | extract | fileinfo | insert | join | merge | permissions | setperms | remove | reverse | rotate | split]")
		fmt.Println("Different actions require different additional parameters, examples below")
		fmt.Println("pdfutil [cp | changepassword] [-in | -infile] <file> [-password user|owner] [-old | -oldpass] <oldpass> [-new | -newpass] <newpass> [-op | -ownerpass] <passwd> [-up | -userpass] <passwd>")
		fmt.Println("pdfutil [de | dec | decrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd>")
		fmt.Println("pdfutil [en | enc | encrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> [-eb | -encryptionbits] <int num> 40, 128, or 256 (default 256)")
		fmt.Println("pdfutil [ex | ext | extract] [-in | -infile] <file> [ -pg | -pages | -pagelist] <page list> [-od | -outdir] <output directory> (file can not be encrypted) [-pad | -zeropad] (optional zero pad filenames)")
		fmt.Println("\tPage list must be provided, e.g. 1,3-5,100-l (no spaces between commas, -l is special syntax for last page)")
		fmt.Println("pdfutil [fi | info | fileinfo] [-in | -infile] <file> (file not encrypted)")
		fmt.Println("pdfutil [fi | info | fileinfo] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> (file encrypted)")
		fmt.Println("pdfutil [in | ins | insert] [-in | -infile] <file> [-if | -ins | -insertfile] <file> [-out | -outfile] <file> [-ia | -insaft | -insertafter] <int page num>")
		fmt.Println("pdfutil [me | mer | merge | join] [-mf | -merge | -mergefiles] <comma separated file list> [-out | -outfile] <file> (files can not be encrypted)")
		fmt.Println("pdfutil [p | perm | permissions | getperms | getpermissions] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> (shows detailed permission breakdown)")
		fmt.Println("pdfutil [stp | setp | setperm | setperms | setpermissions] [-in | -infile] <file> [-op | -ownerpass] <passwd> [-pt | -permtype] <type> (modifies in-place)")
		fmt.Println("\tOptional: [-out | -outfile] <file> (creates new file instead of in-place)")
		fmt.Println("\tPermission types: all, none, print, readonly, forms, annotate, modify")
		fmt.Println("\tCustom combinations: Use comma-separated individual permissions (e.g., print,annotate,extract)")
		fmt.Println("pdfutil [prop | props | properties] <list|add|remove> [-in | -infile] <file> [-out | -outfile] <file> (document properties management)")
		fmt.Println("\tList: pdfutil properties list -in <file> [-up <userpass>] [-op <ownerpass>]")
		fmt.Println("\tAdd: pdfutil properties add -in <file> -out <file> -title <title> -author <author> ...")
		fmt.Println("\tRemove: pdfutil properties remove -in <file> -out <file> -prop <property1,property2,...>")
		fmt.Println("pdfutil [rm | rem | remove] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -page | -pages] <comma separated page list> (file can not be encrypted)")
		fmt.Println("pdfutil [re | rev | reverse] [-in | -infile] <file> [-out | -outfile] <file> (file can not be encrypted)")
		fmt.Println("pdfutil [ro | rot | rotate] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -page | -pages] <comma separated page list> [-ro | -rot | -rotate] <int num 90, 180, 270> (file can not be encrypted)")
		fmt.Println("pdfutil [sp | spl | split] [-in | -infile] <file> [-od | -outdir] <directory> (file can not be encrypted) [-pad | -zeropad] (optional zero pad filenames)")
		fmt.Println("To check for updates, use cu or checkupdate")
		fmt.Println("Example: pdfutil [cu | chk | checkupdate]")
		fmt.Println("NOTE: This will not auto update, this is a manual check, and a manual\nupdate of a portable app")
		flag.PrintDefaults()
	}

	// Create separate FlagSets for each action
	// This happens within case statements below for efficiency

	switch action {
	case "help", "h", "-h", "--h", "-help", "--help":
		{
			flag.Usage()
			os.Exit(0)
		}
	case "version", "ver", "-v", "--v", "-version", "--version":
		{
			fmt.Println(appName, "version", appVersion, "-", appCopyright)
			os.Exit(0)
		}
	case "checkupdate", "cu", "chk", "c":
		{
			// Check for updates and exit
			fmt.Println("Checking for updates...")
			updtmsg, updateAvail := updateChecker("amarillier", "pdfutil", "pdfutil", "https://github.com/amarillier/pdfutil/releases/latest")
			fmt.Println(updtmsg)
			if updateAvail {
				fmt.Println("Update available:", updtmsg)
				fmt.Println("Please visit the GitHub releases page to download the latest version.")
			}
			os.Exit(0)
		}

	case "changepassword", "cp", "chpwd", "changepwd":
		{
			// Change user or owner password of an encrypted PDF
			changePwdCmd := flag.NewFlagSet("changepassword", flag.ExitOnError)
			inFile := changePwdCmd.String("in", "", "Input file (no default value)")
			outFile := changePwdCmd.String("out", "", "Output file (no default value)")
			passwordType := changePwdCmd.String("password", "", "Password type: user or owner (no default value)")
			oldPassword := changePwdCmd.String("old", "", "Old password (no default value)")
			newPassword := changePwdCmd.String("new", "", "New password (no default value)")
			userPassword := changePwdCmd.String("up", "", "User password (if changing owner password)")
			ownerPassword := changePwdCmd.String("op", "", "Owner password (if changing user password)")
			// Aliases
			changePwdCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
			changePwdCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
			changePwdCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
			changePwdCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
			changePwdCmd.StringVar(passwordType, "pwd", "", "Password type (alias for -password)")
			changePwdCmd.StringVar(passwordType, "p", "", "Password type (alias for -password)")
			changePwdCmd.StringVar(oldPassword, "oldpass", "", "Old password (alias for -old)")
			changePwdCmd.StringVar(oldPassword, "oldpwd", "", "Old password (alias for -old)")
			changePwdCmd.StringVar(newPassword, "newpass", "", "New password (alias for -new)")
			changePwdCmd.StringVar(newPassword, "newpwd", "", "New password (alias for -new)")
			changePwdCmd.StringVar(userPassword, "u", "", "User password (alias for -up)")
			changePwdCmd.StringVar(userPassword, "userpass", "", "User password (alias for -up)")
			changePwdCmd.StringVar(ownerPassword, "ownerpass", "", "Owner password (alias for -op)")
			changePwdCmd.Parse(os.Args[2:])

			if len(os.Args) <= 3 {
				fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
				fmt.Println("pdfutil [cp | changepassword] [-in | -infile] <file> [-password user|owner] [-old | -oldpass] <oldpass> [-new | -newpass] <newpass> [-op | -ownerpass] <passwd> [-up | -userpass] <passwd>")
				fmt.Println("Optional: [-out | -outfile] <file> (if omitted, modifies file in-place)")
				os.Exit(1)
			}

			if *inFile == "" {
				fmt.Println("Input file must be provided, use -in | -infile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*inFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if !exists {
				fmt.Printf("File '%s' does not exist.\n", *inFile)
				os.Exit(1)
			}

			if !isEncrypted(*inFile) {
				fmt.Println(*inFile, "is NOT encrypted - file must be encrypted to change passwords")
				os.Exit(1)
			}

			if *passwordType == "" {
				fmt.Println("Password type must be provided, use -password user or -password owner")
				os.Exit(1)
			}

			if *oldPassword == "" {
				fmt.Println("Old password must be provided, use -old | -oldpass <password>")
				os.Exit(1)
			}

			if *newPassword == "" {
				fmt.Println("New password must be provided, use -new | -newpass <password>")
				os.Exit(1)
			}

			*passwordType = strings.TrimSpace(strings.ToLower(*passwordType))
			if *passwordType != "user" && *passwordType != "owner" {
				fmt.Println("Password type must be 'user' or 'owner', use -password user or -password owner")
				os.Exit(1)
			}

			// If no output file specified, modify in-place
			if *outFile == "" {
				fmt.Printf("No output file specified, modifying '%s' in-place\n", *inFile)
			} else {
				// Check if output file already exists
				exists, err = pathExists(*outFile)
				if err != nil {
					fmt.Printf("Error checking file existence: %v\n", err)
					return
				}
				if exists {
					fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
					os.Exit(1)
				}
			}

			if *passwordType == "user" {
				if *ownerPassword == "" {
					fmt.Println("Owner password must be provided when changing user password, use -op | -ownerpass <password>")
					os.Exit(1)
				}
				if *outFile == "" {
					fmt.Printf("Changing user password for %s (in-place)\n", *inFile)
				} else {
					fmt.Printf("Changing user password for %s, output to %s\n", *inFile, *outFile)
				}
				err = changeUserPassword(*inFile, *outFile, *oldPassword, *newPassword, *ownerPassword)
				if err != nil {
					fmt.Printf("Error changing user password: %v\n", err)
					os.Exit(1)
				}
			} else { // owner password
				if *userPassword == "" {
					fmt.Println("User password must be provided when changing owner password, use -up | -userpass <password>")
					os.Exit(1)
				}
				if *outFile == "" {
					fmt.Printf("Changing owner password for %s (in-place)\n", *inFile)
				} else {
					fmt.Printf("Changing owner password for %s, output to %s\n", *inFile, *outFile)
				}
				err = changeOwnerPassword(*inFile, *outFile, *oldPassword, *newPassword, *userPassword)
				if err != nil {
					fmt.Printf("Error changing owner password: %v\n", err)
					os.Exit(1)
				}
			}
		}

	case "decrypt", "de", "dec", "d":
		// sample: pdfutil de -up allanuser -op allanowner -in test.pdf
		/*
			// decryptFile("reversed2.pdf", userPass, ownerPass, encryptionBits)
			// encryptionBits actually does no matter, pass 0
			decryptFile("reversed2.pdf", userPass, ownerPass, 0)
			fileInfo("reversed2.pdf", "", "", 0)
		*/

		decryptCmd := flag.NewFlagSet("decrypt", flag.ExitOnError)
		inFile := decryptCmd.String("in", "", "Input file to process (no default value)")
		userPass := decryptCmd.String("up", "", "User password for decryption (no default value)")
		ownerPass := decryptCmd.String("op", "", "Owner password for decryption (no default value)")
		// Aliases for flags
		decryptCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		decryptCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		decryptCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		decryptCmd.StringVar(userPass, "u", "", "User password (alias for -up)")

		decryptCmd.StringVar(ownerPass, "o", "", "Owner password (alias for -op)")
		decryptCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [de | dec | decrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd>")
			os.Exit(1)
		}

		if *inFile == "" {
			fmt.Println("Input file must be provided, use [-in | -infile] <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if *userPass == "" && *ownerPass == "" {
			fmt.Println("Owner password or user password must be provided, either password will work. Use [-up | -userpass] <password> or [-op | -ownerpass] <password>")
			os.Exit(1)
		}
		// first test if the file is not already encrypted, exit with no action
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is already encrypted - continuing")
		} else {
			fmt.Println(*inFile, "is not encrypted - exiting")
			os.Exit(0)
		}
		fmt.Println("Decrypting file", *inFile, "with user and owner passwords")
		err = decryptFile(*inFile, *userPass, *ownerPass, 0)
		if err != nil {
			fmt.Printf("Error decrypting file: %v\n", err)
		}

	case "encrypt", "en", "enc", "e":
		// sample: pdfutil en -up allanuser -op allanowner -in test.pdf -eb 256
		// encryptFile(inFile, userPass, ownerPass, encryptionBits)
		// set permissions - set deny printing
		// permissionsSet("reversed2.pdf", "temp.pdf", userPass, ownerPass, encryptionBits)

		encryptCmd := flag.NewFlagSet("encrypt", flag.ExitOnError)
		inFile := encryptCmd.String("in", "", "Input file to process (no default value)")
		userPass := encryptCmd.String("up", "", "User password for decryption (no default value)")
		ownerPass := encryptCmd.String("op", "", "Owner password for decryption (no default value)")
		encryptionBits := encryptCmd.Int("eb", 256, "Encryption bits (40, 128 or 256 - default 256)")
		encryptionMode := encryptCmd.String("mo", "aes", "Encryption mode (aes or rc4 - default aes)")
		// Aliases for flags

		encryptCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		encryptCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		encryptCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		encryptCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		encryptCmd.StringVar(ownerPass, "o", "", "Owner password (alias for -op)")
		encryptCmd.IntVar(encryptionBits, "encryptionbits", 256, "Encryption bits (alias for -eb)")
		encryptCmd.IntVar(encryptionBits, "e", 256, "Encryption bits (alias for -eb)")
		encryptCmd.StringVar(encryptionMode, "mode", "aes", "Encryption mode (alias for -mo)")
		encryptCmd.StringVar(encryptionMode, "m", "aes", "Encryption mode (alias for -mo)")
		encryptCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [en | enc | encrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> [-eb | encryptionbits] <int num> 40, 128 or 256")
			fmt.Println("Encryption must be AES 256 for PDF version 2 and above")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use [-in | -infile] <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is already encrypted - exiting")
			os.Exit(0)
		} else {
			fmt.Println(*inFile, "is not yet encrypted - continuing")
			if *ownerPass == "" {
				fmt.Println("Owner password must be provided, user password is optional. Use [-op | -ownerpass] <password> [-up | -userpass] <password>")
				os.Exit(1)
			}
		}
		*encryptionMode = strings.TrimSpace(strings.ToUpper(*encryptionMode))
		fmt.Println("Encrypting file", *inFile, "user and owner passwords", *encryptionBits, "", *encryptionMode, "encryption bits")
		err = encryptFile(*inFile, *userPass, *ownerPass, *encryptionBits, *encryptionMode)
		if err != nil {
			fmt.Printf("Error encrypting file: %v\n", err)
		}

	case "extract", "ex", "x", "ext":
		// sample:
		// Extract only specific pages - splits to files named <filename>_page_pagenumber.pdf
		// use -l for last page
		// pages = []string{"3", "4", "5", "9-13"}
		// extractPages(inFile, outDir, pages)

		extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
		inFile := extractCmd.String("in", "", "Input file to process (no default value)")
		outDir := extractCmd.String("od", "", "Output directory (no default value)")
		pageList := extractCmd.String("pg", "", "Page list (no default value)")
		userPass := extractCmd.String("up", "", "User password (if encrypted)")
		ownerPass := extractCmd.String("op", "", "Owner password (if encrypted)")
		pad := extractCmd.Bool("pad", false, "Zero pad output file names (default false)")
		// Aliases for flags
		extractCmd.StringVar(inFile, "infile", "", "Input file password (alias for -in)")
		extractCmd.StringVar(inFile, "i", "", "Input file password (alias for -in)")
		extractCmd.StringVar(outDir, "out", "", "Output directory (alias for -od)")
		extractCmd.StringVar(outDir, "outdir", "", "Output directory (alias for -od)")
		extractCmd.StringVar(outDir, "o", "", "Output directory (alias for -od)")
		extractCmd.StringVar(pageList, "pages", "", "Page list (alias for -pg)")
		extractCmd.StringVar(pageList, "pagelist", "", "Page list (alias for -pg)")
		extractCmd.StringVar(pageList, "p", "", "Page list (alias for -pg)")
		extractCmd.BoolVar(pad, "zeropad", false, "Zero pad output file names (alias for -pad)")
		extractCmd.BoolVar(pad, "z", false, "Zero pad output file names (alias for -pad)")
		extractCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		extractCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		extractCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		// Note: "o" is used for output directory, not owner password, to avoid conflict
		extractCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [ex | ext | extract] [-in | -infile] <file> [ -pg | -pages | -pagelist] <page list> [-out | -od | -outdir] <output directory> (file can not be encrypted) [-pad | -zeropad] (optional zero pad filenames)")
			fmt.Println("Page list must be comma separated, e.g. 1,3-5,100-l (no spaces between commas, -l is special syntax for last page)")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted - can not extract an encrypted file - exiting")
			os.Exit(0)
		}

		if *outDir == "" {
			fmt.Println("Output directory must be provided, use -od | -outdir <directory name>")
			os.Exit(1)
		}

		exists, err = pathExists(*outDir)
		if err != nil {
			fmt.Printf("Error checking directory existence: %v\n", err)
			return
		}
		if !exists {
			err = os.MkdirAll(*outDir, 0700) // permissions)
			if err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
			}
		}

		if *pageList == "" {
			fmt.Println("Page list must be provided, comma separated, use -pg | -pages e.g. 1,3-5,100-l (no spaces between commas, -l is special syntax for last page)")
			os.Exit(1)
		}

		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		*pageList = re.ReplaceAllString(*pageList, ",")
		pages := strings.Split(*pageList, ",")
		fmt.Println("Extracting specified pages", pages, "from ", *inFile, "to", *outDir)
		err = extractPages(*inFile, *outDir, pages, *pad, *userPass, *ownerPass)
		if err != nil {
			fmt.Printf("Error extracting pages from file: %v\n", err)
		}

	case "info", "in", "fi", "fileinfo", "i", "inf":
		// sample use - note fileinfo permissions not working yet
		// fileInfo(inFile, "", "", 0)
		// fileInfo(inFile, userPass, ownerPass, 0)
		// fileInfo(inFile, userPass, ownerPass, encryptionBits) // bits not needed but harmless

		infoCmd := flag.NewFlagSet("info", flag.ExitOnError)
		inFile := infoCmd.String("in", "", "Input file to process (no default value)")
		userPass := infoCmd.String("up", "", "User password for decryption (no default value)")
		ownerPass := infoCmd.String("op", "", "Owner password for decryption (no default value)")
		// Aliases for flags
		infoCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		infoCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		infoCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		infoCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		infoCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		infoCmd.StringVar(ownerPass, "o", "", "Owner password (alias for -op)")
		infoCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("we got ", len(os.Args), "parameters")
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [fi | info | fileinfo] [-in | -infile] <file> [-up | -userpass] <user passwd> [-op | -ownerpass] <owner passwd>")
			fmt.Println("If the file is encrypted, a password is required, user, owner, or both")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted")
			if *userPass == "" && *ownerPass == "" {
				fmt.Println("Owner or User password (or both) must be provided, use [-op | -ownerpass] <password> and / or [-up | -userpass] <password>")
				os.Exit(1)
			}
			fileInfo(*inFile, *userPass, *ownerPass, 0)
			// fmt.Println("File information for", *inFile, "(file is encrypted)")
		} else {
			fileInfo(*inFile, "", "", 0)
			// fmt.Println("File information for", *inFile, "(file is not encrypted)")
		}

	case "insert", "ins":
		// sample:
		// insert a PDF between the pages of another pdf - output 2 with 5 pages, insert one more
		//inFile1 := outFile2
		//insertFile = outDir + "/CustomTest.pdf"
		//outFile3 := "output3.pdf"
		//insertAfterPage = 2 // Specify the page number after which to insert the PDF
		//fmt.Println("outFile3 - insert", outFile3)
		//insertPDFBetweenPages(inFile1, insertFile, outFile3, insertAfterPage)

		insertCmd := flag.NewFlagSet("insert", flag.ExitOnError)
		inFile := insertCmd.String("in", "", "Input file to process (no default value)")
		outFile := insertCmd.String("out", "", "Output file (no default value)")
		insertFile := insertCmd.String("if", "", "File to insert (no default value)")
		insertAfter := insertCmd.Int("ia", 0, "Insert after page (no default value)")
		userPass := insertCmd.String("up", "", "User password (if encrypted)")
		ownerPass := insertCmd.String("op", "", "Owner password (if encrypted)")
		// Aliases for flags
		insertCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		insertCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		insertCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		insertCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		insertCmd.StringVar(insertFile, "ins", "", "File to insert (alias for -if)")
		insertCmd.StringVar(insertFile, "insertfile", "", "File to insert (alias for -if)")
		insertCmd.IntVar(insertAfter, "insaft", 0, "Insert after page (alias for -ia)")
		insertCmd.IntVar(insertAfter, "insafter", 0, "Insert after page (alias for -ia)")
		insertCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		insertCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		insertCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		// Note: "o" is used for output file, not owner password, to avoid conflict
		insertCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [ins | insert] [-in | -infile] <file> [-if | -ins | -insertfile] <file> [-out | -outfile] <file> [-ia | -insaft | -insertafter] <int page num>")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		// Note: Password-protected files are now supported with -up/-op flags

		if *insertFile == "" {
			fmt.Println("Insertion file must be provided, use -if | -ins | -insertfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*insertFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *insertFile)
			os.Exit(1)
		}
		// Note: Password-protected files are now supported with -up/-op flags

		if *outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
			os.Exit(1)
		}

		if *insertAfter == 0 {
			fmt.Println("Insert After Page must be provided, use -ia | -insaft | -insafter <int num> (default 0, set to required page number)")
			os.Exit(1)
		}
		fmt.Println("Inserting file", *insertFile, "after page", *insertAfter, "of input", *inFile, "output written to", *outFile)
		err = insertPDFBetweenPages(*inFile, *insertFile, *outFile, *insertAfter, *userPass, *ownerPass)
		if err != nil {
			fmt.Printf("Error inserting pages in file: %v\n", err)
		}

	case "merge", "me", "mer", "join", "m":
		// sample:
		// merge 3 specific pages into a single pdf
		//mergeFiles = []string{outDir + "/CustomTest.pdf", outDir + "/CustomTest2.pdf", outDir + "/CustomTest3.pdf"} // Add your PDF file paths here
		//outFile = "output.pdf"
		//fmt.Println("outFile", outFile)
		//mergePDFs(mergeFiles, outFile)

		// merge previous 3 pages and two more specific pages into a single pdf
		//mergeFiles = []string{outFile, outDir + "/CustomTest.pdf", outDir + "/CustomTest2.pdf"} // Add your PDF file paths here
		//outFile2 := "output2.pdf"
		//fmt.Println("outFile2", outFile2)
		//mergePDFs(mergeFiles, outFile2)

		mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
		mergeFileList := mergeCmd.String("mf", "", "Merge file list, comma separated, no spaces (no default value)")
		outFile := mergeCmd.String("out", "", "Output file (no default value)")
		// Aliases for flags
		mergeCmd.StringVar(mergeFileList, "merge", "", "Merge file list, comma separated, no spaces (alias for -mf)")
		mergeCmd.StringVar(mergeFileList, "mergefiles", "", "Merge file list, comma separated, no spaces (alias for -mf)")
		mergeCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		mergeCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		mergeCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [m | me | mer | merge | join] [-mf | -merge | -mergefiles] <comma separated file list> [-out | -outfile] <file> (files can not be encrypted)")
			os.Exit(1)
		}
		if *outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
			os.Exit(1)
		}

		if *mergeFileList == "" {
			fmt.Println("File list to be merge must be provided, comma separated, use -mf | -merge | -mergefiles e.g. file-1.pdf,file-3.pdf,file-22.pdf (no spaces between commas, use quotes to enclose file names with spaces)")
			os.Exit(1)
		}
		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		*mergeFileList = re.ReplaceAllString(*mergeFileList, ",")
		mergeFiles := strings.Split(*mergeFileList, ",")
		encryptedFiles := 0
		for _, mergefile := range mergeFiles {
			if isEncrypted(mergefile) {
				fmt.Println(mergefile, "is encrypted - can not merge encrypted files - exiting")
				encryptedFiles++
			}
		}
		if encryptedFiles > 0 {
			fmt.Println("Exiting with no action")
			os.Exit(0)
		}
		fmt.Println("mergefilelist", *mergeFileList)
		fmt.Println("mergefiles", mergeFiles)
		fmt.Println("Merging files", mergeFiles, "into", *outFile)
		err = mergePDFs(mergeFiles, *outFile)
		if err != nil {
			fmt.Printf("Error merging files: %v\n", err)
		}

	case "pe", "perm", "perms", "permissions", "p", "getperms", "getpermissions":
		fmt.Println("permissions")
		// sample:
		// file must be encrypted to apply permissions
		// encrypt the file
		//userPass = "userpass"
		//ownerPass = "ownerpass"
		//encryptionBits = 128
		//encryptFile("reversed2.pdf", userPass, ownerPass, encryptionBits)
		// set permissions - set deny printing
		//permissionsSet("reversed2.pdf", "temp.pdf", userPass, ownerPass, encryptionBits)
		permsCmd := flag.NewFlagSet("perms", flag.ExitOnError)
		inFile := permsCmd.String("in", "", "Input file to process (no default value)")
		userPass := permsCmd.String("up", "", "User password for decryption (no default value)")
		ownerPass := permsCmd.String("op", "", "Owner password for decryption (no default value)")
		// Aliases for flags
		permsCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		permsCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		permsCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		permsCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		permsCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		permsCmd.StringVar(ownerPass, "o", "", "Owner password (alias for -op)")
		permsCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("we got ", len(os.Args), "parameters")
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [p | perm | perms | permissions] [-in | -infile] <file> [-up | -userpass] <user passwd> [-op | -ownerpass] <owner passwd>")
			fmt.Println("If the file is encrypted, a password is required, user, owner, or both")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted")
			if *userPass == "" && *ownerPass == "" {
				fmt.Println("Owner or User password (or both) must be provided, use [-op | -ownerpass] <password> and / or [-up | -userpass] <password>")
				os.Exit(1)
			}
			permissionsGet(*inFile, *userPass, *ownerPass)
		} else {
			permissionsGet(*inFile, "", "")
		}

	case "setperms", "setperm", "stp", "setp", "setpermissions":
		fmt.Println("Set permissions on encrypted PDF")
		// Add a command to set permissions on a PDF file
		// The file must already be encrypted
		setpermsCmd := flag.NewFlagSet("setperms", flag.ExitOnError)
		inFile := setpermsCmd.String("in", "", "Input file (no default value)")
		outFile := setpermsCmd.String("out", "", "Output file (no default value)")
		userPass := setpermsCmd.String("up", "", "User password (no default value)")
		ownerPass := setpermsCmd.String("op", "", "Owner password (no default value)")
		encryptBits := setpermsCmd.Int("eb", 256, "Encryption bits (40, 128 or 256 - default 256)")
		permType := setpermsCmd.String("pt", "print", "Permission type: all, none, print, readonly, forms, annotate, modify (default print)")
		// Aliases for flags
		setpermsCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		setpermsCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		setpermsCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		setpermsCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		setpermsCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		setpermsCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		setpermsCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		setpermsCmd.IntVar(encryptBits, "e", 256, "Encryption bits (alias for -eb)")
		setpermsCmd.IntVar(encryptBits, "encryptionbits", 256, "Encryption bits (alias for -eb)")
		setpermsCmd.StringVar(permType, "permtype", "print", "Permission type (alias for -pt)")
		setpermsCmd.StringVar(permType, "perm", "print", "Permission type (alias for -pt)")
		setpermsCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [stp | setp | setperm | setperms] [-in | -infile] <file> [-op | -ownerpass] <passwd> [-pt | -permtype] <type>")
			fmt.Println("Optional: [-out | -outfile] <file> (if omitted, modifies file in-place)")
			fmt.Println("Optional: [-up | -userpass] <passwd> [-eb | -encryptionbits] <int>")
			fmt.Println("\nPre-defined permission profiles:")
			fmt.Println("  all      - Full access (no restrictions)")
			fmt.Println("  none     - All permissions denied (all restricted)")
			fmt.Println("  print    - Print only (draft + high quality)")
			fmt.Println("  readonly - View and extract only")
			fmt.Println("  forms    - Fill forms only")
			fmt.Println("  annotate - Add annotations/comments")
			fmt.Println("  modify   - Full editing permissions")
			fmt.Println("\nIndividual permissions (can be combined with commas):")
			fmt.Println("  print, printhq, modify, extract, annotate, fillforms, assemble")
			fmt.Println("Examples:")
			fmt.Println("  -pt print,annotate,extract")
			fmt.Println("  -pt printhq,copy,forms")
			os.Exit(1)
		}

		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}

		if !isEncrypted(*inFile) {
			fmt.Println(*inFile, "is NOT encrypted - file must be encrypted to set permissions")
			fmt.Println("Use the 'encrypt' command first to encrypt the file")
			os.Exit(1)
		}

		// If no output file specified, modify in-place (pass empty string to API)
		if *outFile == "" {
			fmt.Printf("No output file specified, modifying '%s' in-place\n", *inFile)
		} else {
			// Check if output file already exists
			exists, err = pathExists(*outFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if exists {
				fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
				os.Exit(1)
			}
		}

		if *ownerPass == "" {
			fmt.Println("Owner password must be provided, use -op | -ownerpass <password>")
			os.Exit(1)
		}

		*permType = strings.TrimSpace(strings.ToLower(*permType))
		if *outFile == "" {
			fmt.Printf("Setting '%s' permissions on %s (in-place)\n", *permType, *inFile)
		} else {
			fmt.Printf("Setting '%s' permissions on %s, output to %s\n", *permType, *inFile, *outFile)
		}
		err = permissionsSet(*inFile, *outFile, *userPass, *ownerPass, *encryptBits, *permType)
		if err != nil {
			fmt.Printf("Error setting permissions: %v\n", err)
			os.Exit(1)
		}

	case "properties", "props", "prop":
		// Handle properties subcommands: list, add, remove
		if len(os.Args) < 3 {
			fmt.Println("Properties command requires a subcommand: list, add, or remove")
			fmt.Println("Usage:")
			fmt.Println("  pdfutil properties list -in <file> [-up <userpass>] [-op <ownerpass>]")
			fmt.Println("  pdfutil properties add -in <file> -out <file> -title <title> -author <author> ...")
			fmt.Println("  pdfutil properties remove -in <file> -out <file> -prop <property1,property2,...>")
			os.Exit(1)
		}

		subCommand := os.Args[2]
		os.Args = append(os.Args[:2], os.Args[3:]...) // Remove subcommand from args

		switch subCommand {
		case "list", "l", "show":
			// List properties
			listCmd := flag.NewFlagSet("properties list", flag.ExitOnError)
			inFile := listCmd.String("in", "", "Input file (no default value)")
			userPass := listCmd.String("up", "", "User password (if encrypted)")
			ownerPass := listCmd.String("op", "", "Owner password (if encrypted)")
			// Aliases
			listCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
			listCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
			listCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
			listCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
			listCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
			listCmd.StringVar(ownerPass, "o", "", "Owner password (alias for -op)")
			listCmd.Parse(os.Args[2:])

			if *inFile == "" {
				fmt.Println("Input file must be provided, use -in | -infile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*inFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if !exists {
				fmt.Printf("File '%s' does not exist.\n", *inFile)
				os.Exit(1)
			}

			err = listProperties(*inFile, *userPass, *ownerPass)
			if err != nil {
				fmt.Printf("Error listing properties: %v\n", err)
				os.Exit(1)
			}

		case "add", "a", "set":
			// Add properties
			addCmd := flag.NewFlagSet("properties add", flag.ExitOnError)
			inFile := addCmd.String("in", "", "Input file (no default value)")
			outFile := addCmd.String("out", "", "Output file (no default value)")
			userPass := addCmd.String("up", "", "User password (if encrypted)")
			ownerPass := addCmd.String("op", "", "Owner password (if encrypted)")
			title := addCmd.String("title", "", "Document title")
			author := addCmd.String("author", "", "Document author")
			subject := addCmd.String("subject", "", "Document subject")
			keywords := addCmd.String("keywords", "", "Document keywords")
			creator := addCmd.String("creator", "", "Document creator")
			producer := addCmd.String("producer", "", "Document producer")
			// Aliases
			addCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
			addCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
			addCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
			addCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
			addCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
			addCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
			addCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
			addCmd.StringVar(ownerPass, "owner", "", "Owner password (alias for -op)")
			addCmd.Parse(os.Args[2:])

			if *inFile == "" {
				fmt.Println("Input file must be provided, use -in | -infile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*inFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if !exists {
				fmt.Printf("File '%s' does not exist.\n", *inFile)
				os.Exit(1)
			}

			if *outFile == "" {
				fmt.Println("Output file must be provided, use -out | -outfile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*outFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if exists {
				fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
				os.Exit(1)
			}

			// Build properties map
			properties := make(map[string]string)
			if *title != "" {
				properties["title"] = *title
			}
			if *author != "" {
				properties["author"] = *author
			}
			if *subject != "" {
				properties["subject"] = *subject
			}
			if *keywords != "" {
				properties["keywords"] = *keywords
			}
			if *creator != "" {
				properties["creator"] = *creator
			}
			if *producer != "" {
				properties["producer"] = *producer
			}

			if len(properties) == 0 {
				fmt.Println("No properties specified. Use -title, -author, -subject, -keywords, -creator, or -producer")
				os.Exit(1)
			}

			err = addProperties(*inFile, *outFile, *userPass, *ownerPass, properties)
			if err != nil {
				fmt.Printf("Error adding properties: %v\n", err)
				os.Exit(1)
			}

		case "remove", "rm", "del":
			// Remove properties
			removeCmd := flag.NewFlagSet("properties remove", flag.ExitOnError)
			inFile := removeCmd.String("in", "", "Input file (no default value)")
			outFile := removeCmd.String("out", "", "Output file (no default value)")
			userPass := removeCmd.String("up", "", "User password (if encrypted)")
			ownerPass := removeCmd.String("op", "", "Owner password (if encrypted)")
			propList := removeCmd.String("prop", "", "Comma-separated list of properties to remove")
			// Aliases
			removeCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
			removeCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
			removeCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
			removeCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
			removeCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
			removeCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
			removeCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
			removeCmd.StringVar(ownerPass, "owner", "", "Owner password (alias for -op)")
			removeCmd.StringVar(propList, "properties", "", "Properties to remove (alias for -prop)")
			removeCmd.StringVar(propList, "p", "", "Properties to remove (alias for -prop)")
			removeCmd.Parse(os.Args[2:])

			if *inFile == "" {
				fmt.Println("Input file must be provided, use -in | -infile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*inFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if !exists {
				fmt.Printf("File '%s' does not exist.\n", *inFile)
				os.Exit(1)
			}

			if *outFile == "" {
				fmt.Println("Output file must be provided, use -out | -outfile <file name>")
				os.Exit(1)
			}
			exists, err = pathExists(*outFile)
			if err != nil {
				fmt.Printf("Error checking file existence: %v\n", err)
				return
			}
			if exists {
				fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
				os.Exit(1)
			}

			if *propList == "" {
				fmt.Println("Properties to remove must be provided, use -prop <property1,property2,...>")
				fmt.Println("Available properties: title, author, subject, keywords, creator, producer")
				os.Exit(1)
			}

			properties := strings.Split(*propList, ",")
			for i, prop := range properties {
				properties[i] = strings.TrimSpace(prop)
			}

			err = removeProperties(*inFile, *outFile, *userPass, *ownerPass, properties)
			if err != nil {
				fmt.Printf("Error removing properties: %v\n", err)
				os.Exit(1)
			}

		default:
			fmt.Printf("Unknown properties subcommand: %s\n", subCommand)
			fmt.Println("Available subcommands: list, add, remove")
			os.Exit(1)
		}

	case "rem", "remove", "del", "delete", "cut":
		fmt.Println("remove pages")
		// add a function to remove pages from a pdf
		// removePages(inFile, outFile, pageList)
		// fmt.Println("Removed pages from", inFile, "and saved to", outFile)
		removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
		inFile := removeCmd.String("in", "", "Input file (no default value)")
		outFile := removeCmd.String("out", "", "Output file (no default value)")
		pageList := removeCmd.String("pg", "", "Page list, comma separated, no spaces (no default value)")
		// Aliases for flags
		removeCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		removeCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		removeCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		removeCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		removeCmd.StringVar(pageList, "p", "", "Page list, comma separated, no spaces (alias for -pg)")
		removeCmd.StringVar(pageList, "page", "", "Page list, comma separated, no spaces (alias for -pg)")
		removeCmd.StringVar(pageList, "pages", "", "Page list, comma separated, no spaces (alias for -pg)")
		removeCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [rem | remove | del | delete | cut] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -page | -pages] <comma separated page list> (file can not be encrypted)")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted - can not remove pages from an encrypted file - exiting")
			os.Exit(0)
		}
		if *outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
			os.Exit(1)
		}
		if *pageList == "" {
			fmt.Println("Page list to remove must be provided, comma separated, use -pg | -page | -pages e.g. 1,2,7-9,35-l (no spaces between commas, -l is special syntax for last)")
			os.Exit(1)
		}
		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		*pageList = re.ReplaceAllString(*pageList, ",")
		pages := strings.Split(*pageList, ",")
		err = removePages(*inFile, *outFile, pages)
		if err != nil {
			fmt.Printf("Error removing pages from file: %v\n", err)
		}
		fmt.Println("Removed pages", pages, "from", *inFile, "and saved to", *outFile)
	case "rotate", "ro", "rot", "r":
		// sample:
		// rotate specified pages 90
		//rotation = 90
		//rotatePages(inFile, "rotated.pdf", []string{"1-5"}, rotation)
		//fmt.Println("Rotated pages 1-5 90 degrees and save to rotated.pdf")
		// rotate all pages 180
		//rotation = 180
		//rotatePages(inFile, "rotated-all.pdf", []string{"1-l"}, rotation)
		//fmt.Println("Rotated all pages 180 degrees and save to rotated-all.pdf")

		rotateCmd := flag.NewFlagSet("rotate", flag.ExitOnError)
		inFile := rotateCmd.String("in", "", "Input file (no default value)")
		outFile := rotateCmd.String("out", "", "Output file (no default value)")
		pageList := rotateCmd.String("pg", "", "Page list, comma separated, no spaces (no default value)")
		rotation := rotateCmd.Int("ro", 90, "Page rotation (90 degree default value)")
		// Aliases for flags
		rotateCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		rotateCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		rotateCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		rotateCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		rotateCmd.StringVar(pageList, "p", "", "Page list, comma separated, no spaces (alias for -pg)")
		rotateCmd.StringVar(pageList, "page", "", "Page list, comma separated, no spaces (alias for -pg)")
		rotateCmd.StringVar(pageList, "pages", "", "Page list, comma separated, no spaces (alias for -pg)")
		rotateCmd.IntVar(rotation, "r", 90, "Page rotation (90 degree default value) (alias for -ro)")
		rotateCmd.IntVar(rotation, "rot", 90, "Page rotation (90 degree default value) (alias for -ro)")
		rotateCmd.IntVar(rotation, "rotate", 90, "Page rotation (90 degree default value) (alias for -ro)")

		rotateCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [ro | rot | rotate] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -page | -pages] <comma separated page list> [-ro | -rot | -rotate] <int num 90, 180, 270> (file can not be encrypted)")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted - can not rotate pages in an encrypted file - exiting")
			os.Exit(0)
		}
		if *outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
			os.Exit(1)
		}
		if *pageList == "" {
			fmt.Println("Page list to rotate must be provided, comma separated, use -pg | -page | -pages e.g. 1,2,7-9,35-l (no spaces between commas, -l is special syntax for last)")
			os.Exit(1)
		}
		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		*pageList = re.ReplaceAllString(*pageList, ",")
		pages := strings.Split(*pageList, ",")
		if *rotation == 0 {
			fmt.Println("Rotation angle must be provided, use -ro | -rot | -rotate e.g. 90, 180 or 270")
			os.Exit(1)
		}
		if *rotation != 90 && *rotation != 180 && *rotation != 270 {
			fmt.Println("Rotation angle must be 90, 180 or 270. No other angles are supported")
			os.Exit(1)
		}
		err = rotatePages(*inFile, *outFile, pages, *rotation)
		if err != nil {
			fmt.Printf("Error rotating pages in file: %v\n", err)
		}

	case "reverse", "re", "rev":
		// sample:
		// reverse the order of all pages
		//reversePages(inFile, "reversed.pdf", true)
		//fmt.Println("Reversed all pages and saved to reversed.pdf")
		//reversePages(inFile, "reversed2.pdf", false)
		//fmt.Println("Reversed all pages and saved to reversed.pdf2")

		reverseCmd := flag.NewFlagSet("reverse", flag.ExitOnError)
		inFile := reverseCmd.String("in", "", "Input file (no default value)")
		outFile := reverseCmd.String("out", "", "Output file (no default value)")
		userPass := reverseCmd.String("up", "", "User password (if encrypted)")
		ownerPass := reverseCmd.String("op", "", "Owner password (if encrypted)")
		// Aliases for flags
		reverseCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		reverseCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		reverseCmd.StringVar(outFile, "o", "", "Output file (alias for -out)")
		reverseCmd.StringVar(outFile, "outfile", "", "Output file (alias for -out)")
		reverseCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		reverseCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		reverseCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		// Note: "o" is used for output file, not owner password, to avoid conflict

		reverseCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [re | rev | reverse] [-in | -infile] <file> [-out | -outfile] <file> (file can not be encrypted)")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		// Note: Password-protected files are now supported with -up/-op flags
		if *outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", *outFile)
			os.Exit(1)
		}
		fmt.Println("Reversing pages in file", *inFile, "individual files written to", *outFile)
		err = reversePages(*inFile, *outFile, false, *userPass, *ownerPass) // reversePDF auto zero pad, set false here
		if err != nil {
			fmt.Printf("Error reversing pages in file: %v\n", err)
		}

	case "split", "sp", "s":
		// sample:
		// Split all pages to separate files - splits to files named <filename>_pagenumber.pdf
		// splitPDF(inFile, outDir)

		splitCmd := flag.NewFlagSet("split", flag.ExitOnError)
		inFile := splitCmd.String("in", "", "Input file (no default value)")
		outDir := splitCmd.String("od", "", "Output directory (no default value)")
		userPass := splitCmd.String("up", "", "User password (if encrypted)")
		ownerPass := splitCmd.String("op", "", "Owner password (if encrypted)")
		pad := splitCmd.Bool("pad", false, "Zero pad output file names (default false)")
		// Aliases for flags
		splitCmd.StringVar(inFile, "i", "", "Input file (alias for -in)")
		splitCmd.StringVar(inFile, "infile", "", "Input file (alias for -in)")
		splitCmd.StringVar(outDir, "o", "", "Output directory (alias for -od)")
		splitCmd.StringVar(outDir, "out", "", "Output directory (alias for -od)")
		splitCmd.StringVar(outDir, "outdir", "", "Output directory (alias for -od)")
		splitCmd.StringVar(userPass, "userpass", "", "User password (alias for -up)")
		splitCmd.StringVar(userPass, "u", "", "User password (alias for -up)")
		splitCmd.StringVar(ownerPass, "ownerpass", "", "Owner password (alias for -op)")
		// Note: "o" is used for output directory, not owner password, to avoid conflict
		splitCmd.BoolVar(pad, "z", false, "Zero pad output file names (alias for -pad)")
		splitCmd.BoolVar(pad, "zeropad", false, "Zero pad output file names (alias for -pad)")

		splitCmd.Parse(os.Args[2:])

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil [sp | split] [-in | -infile] <file> [-out | -od | -outdir] <directory> (file can not be encrypted) [-pad | -zeropad] (optional zero pad filenames)")
			os.Exit(1)
		}
		if *inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(*inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", *inFile)
			os.Exit(1)
		}
		if isEncrypted(*inFile) {
			fmt.Println(*inFile, "is encrypted - can not extract an encrypted file - exiting")
			os.Exit(0)
		}

		if *outDir == "" {
			fmt.Println("Output directory must be provided, use -out | -od | -outdir <directory name>")
			os.Exit(1)
		}

		exists, err = pathExists(*outDir)
		if err != nil {
			fmt.Printf("Error checking directory existence: %v\n", err)
			return
		}
		if !exists {
			err = os.MkdirAll(*outDir, 0700) // permissions)
			if err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
			}
		}
		fmt.Println("Splitting file", *inFile, "individual files written to", *outDir)
		err = splitPDF(*inFile, *outDir, *pad, *userPass, *ownerPass)
		if err != nil {
			fmt.Printf("Error splitting pages from file: %v\n", err)
		}

	default:
		fmt.Println("Unknown action:", action)
	}

	// pdfwatermark

}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
