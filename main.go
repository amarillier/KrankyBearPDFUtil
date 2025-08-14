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
	appVersion = "0.1.0" // see FyneApp.toml
	appAuthor  = "Allan Marillier"
)

var appCopyright = "Copyright (c) Allan Marillier, 2025-" + strconv.Itoa(time.Now().Year())
var err error
var exists bool

// decrypt a pdf file using provided user password, owner password and AES key length
func decryptFile(inFile string, userPw string, ownerPw string, keyLength int) {
	// Decrypting the file using user specified key length
	// keyLength encryptionBits actually does no matter, pass 0
	conf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
	api.DecryptFile(inFile, "", conf) // write output to inFile, not a new file

}

// encrypt a pdf file using provided user password, owner password and AES key length
func encryptFile(inFile string, userPw string, ownerPw string, keyLength int) {
	// Encrypting the file using user specified key length
	conf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
	api.EncryptFile(inFile, "", conf) // write output to inFile, not a new file
}

// extractPages splits a PDF file into individual pages and saves them in the
// specified directory.
func extractPages(inFile string, outDir string, pageNumbers []string) {
	err = api.ExtractPagesFile(inFile, outDir, pageNumbers, nil)
	if err != nil {
		log.Fatalf("Error extracting pages from PDF file: %v", err)
	}
}

// fileInfo reads and prints basic information about a PDF file.
// It includes details like version, number of pages, title, author, etc.
func fileInfo(inFile string, userPass string, ownerPass string, encryptionBits int) {
	// Read and validate the PDF context info

	// Create a configuration with the user and owner passwords when provided
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass
	encryptionBits = 0 // fake use to avoid errors for now, we don't use it

	/*
		f, err := os.Open(inFile)
		if err != nil {
			log.Fatalf("Failed to read open file: %v", err)
		}
		defer f.Close()

		ctx, err := api.ReadContext(f, conf)
		if err != nil {
			fmt.Println("The file appears to be encrypted, enter a user or owner password")
			return
			// log.Fatalf("Failed to read PDF context: %v", err)
		}
	*/
	ctx, err := api.ReadContextFile(inFile)

	// Print basic PDF info
	fmt.Printf("PDF Version: %s\n", ctx.HeaderVersion)
	fmt.Printf("Number of Pages: %d\n", ctx.PageCount)
	fmt.Printf("Title: %s\n", ctx.Title)
	fmt.Printf("Author: %s\n", ctx.Author)
	fmt.Println("Creator:", ctx.Creator)
	fmt.Println("Producer:", ctx.Producer)
	fmt.Printf("Subject: %s\n", ctx.Subject)
	fmt.Printf("Keywords: %s\n", ctx.Keywords)
	fmt.Println("Modification Date:", ctx.ModDate)
	fmt.Println("Permissions:", ctx.Permissions)
	fmt.Println("Creation Date:", ctx.Configuration.CreationDate)
	fmt.Println("Version:", ctx.Configuration.Version)

	perms, err := api.GetPermissionsFile(inFile, nil)
	if err != nil {
		log.Fatal("Error getting permissions")
	}
	fmt.Println("Permissions:", perms)
}

// insertPDFBetweenPages inserts a PDF file between the pages of another PDF file.
// It splits the original PDF into two parts, inserts the specified PDF file,
// and merges them back together.
func insertPDFBetweenPages(inFile string, insertFile string, outFile string, insertAfterPage int) {
	var pages []string
	// Split the original PDF into two parts

	part1 := makeTempDir()
	part2 := makeTempDir()
	if part1 == "" || part2 == "" { // If directory creation failed, exit
		fmt.Println("Failed to create temporary directories")
		return
	}

	pages = []string{"1-" + fmt.Sprint(insertAfterPage)}
	extractPages(inFile, part1, pages)
	files1, err := readFilesInDir(part1)
	if err != nil {
		log.Fatal(err)
	}

	pages = []string{fmt.Sprint(insertAfterPage+1) + "-l"}
	extractPages(inFile, part2, pages)
	files2, err := readFilesInDir(part2)
	if err != nil {
		log.Fatal(err)
	}

	// Merge the parts with the insert PDF
	pages = files1[:]
	pages = append(pages, insertFile)
	pages = append(pages, files2...)
	err = api.MergeCreateFile(pages, outFile, false, nil)
	if err != nil {
		log.Fatalf("Error merging PDF files: %v", err)
	}

	// cleanup temp directories
	err = os.RemoveAll(part1)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", part1, err)
	}
	err = os.RemoveAll(part2)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", part2, err)
	}
}

// fast test if the file is encrypted before other actions
func isEncrypted(inFile string) bool {
	// Read and validate the PDF context info

	// Create a configuration with the user and owner passwords when provided
	conf := model.NewDefaultConfiguration()
	// conf.UserPW = userPass
	// conf.OwnerPW = ownerPass

	f, err := os.Open(inFile)
	if err != nil {
		log.Fatalf("Failed to read open file: %v", err)
	}
	defer f.Close()

	ctx, err := api.ReadContext(f, conf)
	if err != nil || ctx == nil {
		// fmt.Println(err) // err likely to be "pdfcpu: please provide the correct password"
		return true
	} else {
		return false
	}
}

// mergePDFs merges multiple PDF files into a single PDF file.
// It takes a slice of input file paths and an output file path.
func mergePDFs(inFiles []string, outFile string) {
	err := api.MergeCreateFile(inFiles, outFile, false, nil)
	if err != nil {
		log.Fatalf("Error merging PDF files: %v", err)
	}
}

// rotatePages rotates specified pages of a PDF file by a given angle.
// The rotation can be 90, 180, or 270 degrees.
// It takes the input file, output file, a slice of page numbers, and the rotation
func rotatePages(inFile string, outFile string, pages []string, rotation int) {
	if rotation != 90 && rotation != 180 && rotation != 270 {
		log.Fatalf("Invalid rotation value, must be 90, 180 or 270")
	}
	err := api.RotateFile(inFile, outFile, rotation, pages, nil)
	if err != nil {
		log.Fatalf("Error rotating pages: %v", err)
	}
}

// reversePages reverses the order of all pages in a PDF file.
// It splits the original PDF into individual pages, reads them in reverse order,
// and merges them back together into a new PDF file. setContext specifies if
// the same file context settings (creator, permissions etc) from the source file
// should be applied to the destination file
func reversePages(inFile string, outFile string, setContext bool) {
	var pages []string
	// Split the original PDF, read in reverse and join back together

	outDir := makeTempDir()
	// split all to temp directory
	splitPDF(inFile, outDir)

	// zero pad the names for better sorting
	err = zeroPadNames(outDir, 3)
	if err != nil {
		log.Fatalf("Error zero padding names: %v", err)
	}
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
		log.Fatalf("Error merging PDF files: %v", err)
	}

	// cleanup temp directory
	err = os.RemoveAll(outDir)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", outDir, err)
	}

	if setContext {
		// Read and set the PDF context from original if required
		ctx, err := api.ReadContextFile(inFile)
		if err != nil {
			log.Fatalf("Failed to read PDF context: %v", err)
		}
		api.WriteContextFile(ctx, outFile)
	}
}

// set permissions on a copy of the file
func setPermissions(inFile string, outFile string, userPass string, ownerPass string, encryptBits int) {
	// Create a configuration with owner and user passwords
	// encryptFile(outFile, userPass, ownerPass, encryptBits)
	conf := model.NewAESConfiguration(userPass, ownerPass, encryptBits)
	api.SetPermissionsFile(inFile, outFile, conf)

	// Setting all permissions for the AES-256 encrypted in.pdf.
	//conf := model.NewAESConfiguration("userpw", "ownerpw", 256)
	//conf.Permissions = model.PermissionsAll
	//api.SetPermissionsFile(inFile, "", conf)

	// Restricting permissions for the AES-256 encrypted in.pdf.
	// conf = model.NewAESConfiguration("upw", "opw", 256)
	// conf.Permissions = model.PermissionsNone // PermissionsPrint

	// Disable printing
	// conf.Permissions.Print = false
	// conf.Permissions.PrintHighRes = false
}

// splitPDF splits a PDF file into individual pages and saves them in the
// specified directory.
// It splits the PDF into files named <filename>_pagenumber.pdf.
func splitPDF(inFile string, outDir string) {
	err := api.SplitFile(inFile, outDir, 1, nil)
	if err != nil {
		log.Fatalf("Error splitting PDF file: %v", err)
	}
}

func main() {
	var action string
	var checkupdate bool
	var encryptionBits int
	encryptionBits = 0 // force to 0 initially
	var insertAfterPage int
	var insertFile string
	var inFile string
	var mergeFiles []string
	var mergefilelist string
	var outFile string
	var outDir string
	outDir = "outdir"
	var pages []string
	var pagelist string
	var rotation int
	var userPass string
	var ownerPass string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Println("Kranky Bear pdfutil is a simple PDF management utility")
		fmt.Println(appName, "version", appVersion, "-", appCopyright)
		fmt.Println("Examples:")
		fmt.Println("pdfutil -a | -action [decrypt | encrypt | extract | fileinfo | insert | merge | permissions | rotate | reverse | split]")
		fmt.Println("Different actions require different additional parameters, examples below")
		fmt.Println("pdfutil -a [de | decrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd>")
		fmt.Println("pdfutil -a [en | encrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> [-eb | encryptionbits] <int num> 256, 512 etc")
		fmt.Println("pdfutil -a [ex | extract] [-in | -infile] <file> [-od | -outdir] <output directory> (file can not be encrypted)")
		fmt.Println("pdfutil -a [fi | fileinfo] [-in | -infile] <file> (file not encrypted)")
		fmt.Println("pdfutil -a [fi | fileinfo] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> (file encrypted)")
		fmt.Println("pdfutil -a [in | insert] [-in | -infile] <file> [-if | -ins | -insertfile] <file> [-out | -outfile] <file> [-ia | -insaft | -insertafter] <int page num>")
		fmt.Println("pdfutil -a [me | merge] [-mf | -merge | -mergefiles] <comma separated file list> [-out | -outfile] <file> (files can not be encrypted)")
		fmt.Println("pdfutil -a [p | permissions]")
		fmt.Println("pdfutil -a [ro | rot | rotate] [fi | fileinfo] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -pages] <comma separated page list> [-ro | -rot | -rotation] <int num 90, 180, 270> (file can not be encrypted)")
		fmt.Println("pdfutil -a [re | rev | reverse] [-in | -infile] <file> [-out | -outfile] <file> (file can not be encrypted)")
		fmt.Println("pdfutil -a [sp | split] [-in | -infile] <file> [-od | -outdir] <directory> (file can not be encrypted)")
		fmt.Println("To check for updates, use -checkupdate or -cu")
		fmt.Println("Example: pdfutil [-cu | -checkupdate]")
		fmt.Println("NOTE: This will not auto update, this is a manual check, and a manual\nupdate of a portable app")
		flag.PrintDefaults()
	}

	flag.StringVar(&action, "a", "", "Action to perform (no default)")
	flag.StringVar(&action, "act", "", "Action to perform (no default)")
	flag.StringVar(&action, "action", "", "Action to perform (no default)")
	flag.BoolVar(&checkupdate, "checkupdate", false, "Check for updates (default: 0, no check)")
	flag.BoolVar(&checkupdate, "cu", false, "Check for updates (default: 0, no check)")
	flag.IntVar(&encryptionBits, "eb", 0, "AES encryption bits (default: 0)")
	flag.IntVar(&encryptionBits, "encryptionbits", 0, "AES encryption bits (default: 0)")
	flag.IntVar(&insertAfterPage, "ia", 0, "Insert After Page (default: 0)")
	flag.IntVar(&insertAfterPage, "insaft", 0, "Insert After Page (default: 0)")
	flag.IntVar(&insertAfterPage, "insertafter", 0, "Insert After Page (default: 0)")
	flag.StringVar(&insertFile, "if", "", "Insert File (no default)")
	flag.StringVar(&insertFile, "ins", "", "Insert File (no default)")
	flag.StringVar(&insertFile, "insertfile", "", "Insert File (no default)")
	flag.StringVar(&inFile, "in", "", "Input File (no default)")
	flag.StringVar(&inFile, "infile", "", "Input File (no default)")
	flag.StringVar(&outFile, "out", "", "Output File (no default)")
	flag.StringVar(&outFile, "outfile", "", "Output File (no default)")
	flag.StringVar(&outDir, "od", "", "Output Directory (no default)")
	flag.StringVar(&outDir, "outdir", "", "Output Directory (no default)")
	flag.StringVar(&mergefilelist, "merge", "", "Merge files, comma separated list (no default)")
	flag.StringVar(&mergefilelist, "mf", "", "Merge files, comma separated list (no default)")
	flag.StringVar(&mergefilelist, "mergefiles", "", "Merge files, comma separated list (no default)")
	flag.StringVar(&pagelist, "pg", "", "Pages, comma separated list (no default e.g. '1,3,5-9,100-l' l = last)")
	flag.StringVar(&pagelist, "page", "", "Pages, comma separated list (no default e.g. '1,3,5-9,100-l' l = last)")
	flag.StringVar(&pagelist, "pages", "", "Pages, comma separated list (no default e.g. '50-l' - l = last)")
	flag.IntVar(&rotation, "ro", 0, "Page Rotation (default: 0)")
	flag.IntVar(&rotation, "rot", 0, "Page Rotation (default: 0)")
	flag.IntVar(&rotation, "rotation", 0, "Page Rotation (default: 0)")
	flag.StringVar(&userPass, "up", "", "User Password (no default)")
	flag.StringVar(&userPass, "userpass", "", "User Password (no default)")
	flag.StringVar(&ownerPass, "op", "", "Owner Password (no default)")
	flag.StringVar(&ownerPass, "ownerpass", "", "Owner Password (no default)")

	flag.Parse()

	// Check if help flag is set
	if flag.NFlag() == 0 || flag.Lookup("help") != nil {
		flag.Usage()
		os.Exit(0)
	}

	if checkupdate {
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

	switch action {
	case "decrypt", "de":
		// sample: pdfutil -a de -up allanuser -op allanowner -in test.pdf
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [de | decrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd>")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use [-in | -infile] <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		// first test if the file is not already encrypted, exit with no action
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is already encrypted - continuing")
			if userPass == "" {
				fmt.Println("User password must be provided, use [-up | -userpass] <password>")
				os.Exit(1)
			}
			if ownerPass == "" {
				fmt.Println("Owner password must be provided, use [-op | -ownerpass] <password>")
				os.Exit(1)
			}
		} else {
			fmt.Println(inFile, "is not encrypted - exiting")
			os.Exit(0)
		}
		fmt.Println("Decrypting file", inFile, "user and owner passwords")
		decryptFile(inFile, userPass, ownerPass, 0)
		/*
			// decryptFile("reversed2.pdf", userPass, ownerPass, encryptionBits)
			// encryptionBits actually does no matter, pass 0
			decryptFile("reversed2.pdf", userPass, ownerPass, 0)
			fileInfo("reversed2.pdf", "", "", 0)
		*/
	case "encrypt", "en":
		// sample: pdfutil -a en -up allanuser -op allanowner -in test.pdf -eb 256
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [en | encrypt] [-in | -infile] <file> [-up | -userpass] <passwd> [-op | -ownerpass] <passwd> [-eb | encryptionbits] <int num> 256, 512 etc")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use [-in | -infile] <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is already encrypted - exiting")
			os.Exit(0)
		} else {
			fmt.Println(inFile, "is not yet encrypted - continuing")
			if userPass == "" {
				fmt.Println("User password must be provided, use [-up | -userpass] <password>")
				os.Exit(1)
			}
			if ownerPass == "" {
				fmt.Println("Owner password must be provided, use [-op | -ownerpass] <password>")
				os.Exit(1)
			}
			if encryptionBits == 0 {
				fmt.Println("Encryption bits must be provided, use [-eb | -encryptionbits] <int num> (default 0, set to 128, 256 etc)")
				os.Exit(1)
			}
		}
		fmt.Println("Encrypting file", inFile, "user and owner passwords", encryptionBits, "encryption bits")
		encryptFile(inFile, userPass, ownerPass, encryptionBits)
		/*
			encryptFile(inFile, userPass, ownerPass, encryptionBits)
			// set permissions - set deny printing
			// setPermissions("reversed2.pdf", "temp.pdf", userPass, ownerPass, encryptionBits)
		*/
	case "extract", "ex":
		// sample:
		/*
			// Extract only specific pages - splits to files named <filename>_page_pagenumber.pdf
			// use -l for last page
			pages = []string{"3", "4", "5", "9-13"}
			extractPages(inFile, outDir, pages)
		*/
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [ex | extract] [-in | -infile] <file> [-od | -outdir] <output directory> (file can not be encrypted)")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted - can not extract an encrypted file - exiting")
			os.Exit(0)
		}

		if outDir == "" {
			fmt.Println("Output directory must be provided, use -od | -outdir <directory name>")
			os.Exit(1)
		}

		exists, err = pathExists(outDir)
		if err != nil {
			fmt.Printf("Error checking directory existence: %v\n", err)
			return
		}
		if !exists {
			err = os.MkdirAll(outDir, 0700) // permissions)
			if err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
			}
		}

		if pagelist == "" {
			fmt.Println("Page list must be provided, comma separated, use -pg | -pages e.g. 1,3-5,100-l (no spaces between commas, -l is special syntax for last page)")
			os.Exit(1)
		}

		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		pagelist = re.ReplaceAllString(pagelist, ",")
		pages = strings.Split(pagelist, ",")
		fmt.Println("Extracting specified pages", pages, "from ", inFile, "to", outDir)
		extractPages(inFile, outDir, pages)

	case "fileinfo", "fi":
		// sample use - note fileinfo permissions not working yet
		// fileInfo(inFile, "", "", 0)
		// fileInfo(inFile, userPass, ownerPass, 0)
		// fileInfo(inFile, userPass, ownerPass, encryptionBits) // bits not needed but harmless
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [fi | fileinfo] [-in | -infile] <file> [-up | -userpass] <user passwd> [-op | -ownerpass] <owner passwd>")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted")
			if userPass == "" {
				fmt.Println("User password must be provided, use -up | -userpass <password>")
				os.Exit(1)
			}
			if ownerPass == "" {
				fmt.Println("Owner password must be provided, use -op | -ownerpass <password>")
				os.Exit(1)
			}
			fileInfo(inFile, userPass, ownerPass, 0)
			fmt.Println("File information for", inFile, "(file is encrypted)")
		} else {
			fileInfo(inFile, "", "", 0)
			fmt.Println("File information for", inFile, "(file is not encrypted)")
		}

	case "insert", "in":
		// sample:
		/*
			// insert a PDF between the pages of another pdf - output 2 with 5 pages, insert one more
			inFile1 := outFile2
			insertFile = outDir + "/Custom_Content_Presentation_v10.7_45.pdf"
			outFile3 := "output3.pdf"
			insertAfterPage = 2 // Specify the page number after which to insert the PDF
			fmt.Println("outFile3 - insert", outFile3)
			insertPDFBetweenPages(inFile1, insertFile, outFile3, insertAfterPage)
		*/

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [in | insert] [-in | -infile] <file> [-if | -ins | -insertfile] <file> [-out | -outfile] <file> [-ia | -insaft | -insertafter] <int page num>")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted - can not insert into an encrypted file - exiting")
			os.Exit(0)
		}

		if insertFile == "" {
			fmt.Println("Insertion file must be provided, use -if | -ins | -insertfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(insertFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", insertFile)
			os.Exit(1)
		}
		if isEncrypted(insertFile) {
			fmt.Println(insertFile, "is encrypted - can not insert an encrypted file - exiting")
			os.Exit(0)
		}

		if outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", outFile)
			os.Exit(1)
		}

		if insertAfterPage == 0 {
			fmt.Println("Insert After Page must be provided, use -ia | -insaft | -insafter <int num> (default 0, set to required page number)")
			os.Exit(1)
		}
		fmt.Println("Inserting file", insertFile, "after page", insertAfterPage, "of input", inFile, "output written to", outFile)
		insertPDFBetweenPages(inFile, insertFile, outFile, insertAfterPage)

	case "merge", "me":
		// sample:
		/*
			// merge 3 specific pages into a single pdf
			mergeFiles = []string{outDir + "/Custom_Content_Presentation_v10.7_5.pdf", outDir + "/Custom_Content_Presentation_v10.7_10.pdf", outDir + "/Custom_Content_Presentation_v10.7_13.pdf"} // Add your PDF file paths here
			outFile = "output.pdf"
			fmt.Println("outFile", outFile)
			mergePDFs(mergeFiles, outFile)

			// merge previous 3 pages and two more specific pages into a single pdf
			mergeFiles = []string{outFile, outDir + "/Custom_Content_Presentation_v10.7_108.pdf", outDir + "/Custom_Content_Presentation_v10.7_109.pdf"} // Add your PDF file paths here
			outFile2 := "output2.pdf"
			fmt.Println("outFile2", outFile2)
			mergePDFs(mergeFiles, outFile2)
		*/

		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [me | merge] [-mf | -merge | -mergefiles] <comma separated file list> [-out | -outfile] <file> (files can not be encrypted)")
			os.Exit(1)
		}
		if outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", outFile)
			os.Exit(1)
		}

		if mergefilelist == "" {
			fmt.Println("File list to be merge must be provided, comma separated, use -mf | -merge | -mergefiles e.g. file-1.pdf,file-3.pdf,file-22.pdf (no spaces between commas, use quotes to enclose file names with spaces)")
			os.Exit(1)
		}
		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		mergefilelist = re.ReplaceAllString(mergefilelist, ",")
		mergeFiles = strings.Split(mergefilelist, ",")
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
		fmt.Println("mergefilelist", mergefilelist)
		fmt.Println("mergefiles", mergeFiles)
		fmt.Println("Merging files", mergeFiles, "into", outFile)
		mergePDFs(mergeFiles, outFile)

	case "permissions", "pe":
		fmt.Println("permissions")
		// file must be encrypted to apply permissions
		/*
			// encrypt the file
			userPass = "iamuserpass"
			ownerPass = "iamownerpass"
			encryptionBits = 128
			encryptFile("reversed2.pdf", userPass, ownerPass, encryptionBits)
			// set permissions - set deny printing
			setPermissions("reversed2.pdf", "temp.pdf", userPass, ownerPass, encryptionBits)
		*/

	case "rotate", "ro", "rot":
		// sample:
		/*
			// rotate specified pages 90
			rotation = 90
			rotatePages(inFile, "rotated.pdf", []string{"1-5"}, rotation)
			fmt.Println("Rotated pages 1-5 90 degrees and save to rotated.pdf")
			// rotate all pages 180
			rotation = 180
			rotatePages(inFile, "rotated-all.pdf", []string{"1-l"}, rotation)
			fmt.Println("Rotated all pages 180 degrees and save to rotated-all.pdf")
		*/
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [ro | rot | rotate] [fi | fileinfo] [-in | -infile] <file> [-out | -outfile] <file> [-pg | -pages] <comma separated page list> [-ro | -rot | -rotation] <int num 90, 180, 270> (file can not be encrypted)")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted - can not rotate pages in an encrypted file - exiting")
			os.Exit(0)
		}
		if outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", outFile)
			os.Exit(1)
		}
		if pagelist == "" {
			fmt.Println("Page list to rotate must be provided, comma separated, use -pg | -page | -pages e.g. 1,2,7-9,35-l (no spaces between commas, -l is special syntax for last)")
			os.Exit(1)
		}
		// replace " ," or ", " with just comma
		re := regexp.MustCompile(", | ,")
		pagelist = re.ReplaceAllString(pagelist, ",")
		pages = strings.Split(pagelist, ",")
		if rotation == 0 {
			fmt.Println("Rotation angle must be provided, use -ro | -rot | -rotation e.g. 90, 180 or 270")
			os.Exit(1)
		}
		if rotation != 90 && rotation != 180 && rotation != 270 {
			fmt.Println("Rotation angle must be 90, 180 or 270. No other angles are supported")
			os.Exit(1)
		}

		rotatePages(inFile, outFile, pages, rotation)

	case "reverse", "re", "rev":
		// sample:
		/*
			// reverse the order of all pages
			reversePages(inFile, "reversed.pdf", true)
			fmt.Println("Reversed all pages and saved to reversed.pdf")
			reversePages(inFile, "reversed2.pdf", false)
			fmt.Println("Reversed all pages and saved to reversed.pdf2")
		*/
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [re | rev | reverse] [-in | -infile] <file> [-out | -outfile] <file> (file can not be encrypted)")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted - can not reverse pages in an encrypted file - exiting")
			os.Exit(0)
		}
		if outFile == "" {
			fmt.Println("Output file must be provided, use -out | -outfile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(outFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if exists {
			fmt.Printf("File '%s' already exists, you must remove it manually.\n", outFile)
			os.Exit(1)
		}
		reversePages(inFile, outFile, true)
	case "split", "sp":
		// sample:
		/*
			// Split all pages to separate files - splits to files named <filename>_pagenumber.pdf
			splitPDF(inFile, outDir)
		*/
		if len(os.Args) <= 3 {
			fmt.Println("Not enough parameters provided. Use -h or --help for full usage information.")
			fmt.Println("pdfutil -a [sp | split] [-in | -infile] <file> [-od | -outdir] <directory> (file can not be encrypted)")
			os.Exit(1)
		}
		if inFile == "" {
			fmt.Println("Input file must be provided, use -in | -infile <file name>")
			os.Exit(1)
		}
		exists, err = pathExists(inFile)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}
		if !exists {
			fmt.Printf("File '%s' does not exist.\n", inFile)
			os.Exit(1)
		}
		if isEncrypted(inFile) {
			fmt.Println(inFile, "is encrypted - can not extract an encrypted file - exiting")
			os.Exit(0)
		}

		if outDir == "" {
			fmt.Println("Output directory must be provided, use -od | -outdir <directory name>")
			os.Exit(1)
		}

		exists, err = pathExists(outDir)
		if err != nil {
			fmt.Printf("Error checking directory existence: %v\n", err)
			return
		}
		if !exists {
			err = os.MkdirAll(outDir, 0700) // permissions)
			if err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
			}
		}
		fmt.Println("Splitting file", inFile, "individual files written to", outDir)
		splitPDF(inFile, outDir)

	default:
		fmt.Println("Unknown action:", action)
	}

	// pdfwatermark

}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
