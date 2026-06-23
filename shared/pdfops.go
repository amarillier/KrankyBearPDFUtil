package shared

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDF Operation functions shared between CLI and GUI

// DecryptFile removes encryption from a PDF file
func DecryptFile(inFile string, userPw string, ownerPw string, keyLength int) error {
	conf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
	err := api.DecryptFile(inFile, "", conf)
	return err
}

// EncryptFile encrypts a PDF file using provided passwords and encryption settings
func EncryptFile(inFile string, userPw string, ownerPw string, keyLength int, mode string) error {
	if mode == "AES" {
		aesConf := model.NewAESConfiguration(userPw, ownerPw, keyLength)
		err := api.EncryptFile(inFile, "", aesConf)
		return err
	}
	// for RC4 encryption
	rc4Conf := model.NewRC4Configuration(userPw, ownerPw, keyLength)
	err := api.EncryptFile(inFile, "", rc4Conf)
	return err
}

// ExtractPages extracts specified pages from a PDF file
func ExtractPages(inFile string, outDir string, pageNumbers []string, pad bool, userPass string, ownerPass string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err := api.ExtractPagesFile(inFile, outDir, pageNumbers, conf)
	if err != nil {
		return fmt.Errorf("error extracting pages: %v", err)
	}
	if pad {
		err = ZeroPadNames(outDir, 3)
		if err != nil {
			return fmt.Errorf("error zero padding names: %v", err)
		}
	}
	return nil
}

// IsEncrypted checks if a PDF file is encrypted
func IsEncrypted(inFile string) bool {
	conf := model.NewDefaultConfiguration()
	f, err := os.Open(inFile)
	if err != nil {
		return false
	}
	defer f.Close()

	ctx, err := api.ReadContext(f, conf)
	if err != nil || ctx.Encrypt != nil {
		return true
	}
	return false
}

// MergePDFs merges multiple PDF files into a single PDF
func MergePDFs(inFiles []string, outFile string) error {
	err := api.MergeCreateFile(inFiles, outFile, false, nil)
	return err
}

// RemovePages removes specified pages from a PDF file
func RemovePages(inFile string, outFile string, pages []string) error {
	err := api.RemovePagesFile(inFile, outFile, pages, nil)
	return err
}

// RotatePages rotates specified pages by a given angle (90, 180, or 270)
func RotatePages(inFile string, outFile string, pages []string, rotation int) error {
	if rotation != 90 && rotation != 180 && rotation != 270 {
		return fmt.Errorf("invalid rotation value, must be 90, 180 or 270")
	}
	err := api.RotateFile(inFile, outFile, rotation, pages, nil)
	return err
}

// SplitPDF splits a PDF file into individual pages
func SplitPDF(inFile string, outDir string, pad bool, userPass string, ownerPass string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err := api.SplitFile(inFile, outDir, 1, conf)
	if err != nil {
		return fmt.Errorf("error splitting PDF: %v", err)
	}
	if pad {
		err = ZeroPadNames(outDir, 3)
		if err != nil {
			return fmt.Errorf("error zero padding names: %v", err)
		}
	}
	return nil
}

// ReversePages reverses the order of all pages in a PDF
func ReversePages(inFile string, outFile string, setContext bool, userPass string, ownerPass string) error {
	outDir := MakeTempDir()
	if outDir == "" {
		return fmt.Errorf("failed to create temporary directory")
	}
	defer os.RemoveAll(outDir)

	err := SplitPDF(inFile, outDir, true, userPass, ownerPass)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	var pages []string
	for _, file := range files {
		if !file.IsDir() {
			pages = append(pages, filepath.Join(outDir, file.Name()))
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(pages)))

	err = api.MergeCreateFile(pages, outFile, false, nil)
	if err != nil {
		return fmt.Errorf("error merging PDF files: %v", err)
	}

	if setContext {
		ctx, err := api.ReadContextFile(inFile)
		if err != nil {
			return fmt.Errorf("failed to read PDF context: %v", err)
		}
		err = api.WriteContextFile(ctx, outFile)
		if err != nil {
			return fmt.Errorf("failed to write PDF context: %v", err)
		}
	}
	return nil
}

// InsertPDFBetweenPages inserts a PDF file between pages of another PDF
func InsertPDFBetweenPages(inFile string, insertFile string, outFile string, insertAfterPage int, userPass string, ownerPass string) error {
	part1 := MakeTempDir()
	part2 := MakeTempDir()
	if part1 == "" || part2 == "" {
		return fmt.Errorf("failed to create temporary directories")
	}
	defer os.RemoveAll(part1)
	defer os.RemoveAll(part2)

	pages := []string{fmt.Sprintf("1-%d", insertAfterPage)}
	err := ExtractPages(inFile, part1, pages, false, userPass, ownerPass)
	if err != nil {
		return err
	}
	files1, err := ReadFilesInDir(part1)
	if err != nil {
		return err
	}

	pages = []string{fmt.Sprintf("%d-l", insertAfterPage+1)}
	err = ExtractPages(inFile, part2, pages, false, userPass, ownerPass)
	if err != nil {
		return err
	}
	files2, err := ReadFilesInDir(part2)
	if err != nil {
		return err
	}

	pages = files1[:]
	pages = append(pages, insertFile)
	pages = append(pages, files2...)
	err = api.MergeCreateFile(pages, outFile, false, nil)
	return err
}

// GetFileInfo retrieves basic information about a PDF file
func GetFileInfo(inFile string, userPass string, ownerPass string) (*model.Context, error) {
	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.LISTINFO
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	var ctx *model.Context
	var err error

	if IsEncrypted(inFile) {
		f, err := os.Open(inFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()

		ctx, err = api.ReadAndValidate(f, conf)
		if err != nil {
			return nil, fmt.Errorf("failed to read encrypted PDF: %v", err)
		}
	} else {
		ctx, err = api.ReadContextFile(inFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read PDF: %v", err)
		}
	}
	return ctx, nil
}

// PermissionDetails holds detailed permission information
type PermissionDetails struct {
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

// DecodePermissions decodes permission bits into human-readable details
func DecodePermissions(perms int16) PermissionDetails {
	details := PermissionDetails{
		RawValue: perms,
	}
	details.Print = (perms & (1 << 2)) != 0
	details.Modify = (perms & (1 << 3)) != 0
	details.Extract = (perms & (1 << 4)) != 0
	details.Annotate = (perms & (1 << 5)) != 0
	details.FillForms = (perms & (1 << 8)) != 0
	details.ExtractRev3 = (perms & (1 << 9)) != 0
	details.Assemble = (perms & (1 << 10)) != 0
	details.PrintHighRes = (perms & (1 << 11)) != 0
	return details
}

// GetPermissions retrieves permissions for a PDF file
func GetPermissions(inFile string, userPw string, ownerPw string) (*int16, error) {
	conf := model.NewDefaultConfiguration()
	conf.OwnerPW = ownerPw
	conf.UserPW = userPw

	perms, err := api.GetPermissionsFile(inFile, conf)
	if err != nil {
		return nil, fmt.Errorf("error getting permissions: %v", err)
	}
	return perms, nil
}

// SetPermissions sets specific permissions on a PDF file
func SetPermissions(inFile string, outFile string, userPass string, ownerPass string, encryptBits int, permType string) error {
	conf := model.NewAESConfiguration(userPass, ownerPass, encryptBits)

	// Parse permission type (simplified - would need full logic from main.go)
	switch permType {
	case "all", "full":
		conf.Permissions = model.PermissionsAll
	case "none", "restrict":
		conf.Permissions = model.PermissionsNone
	case "print":
		conf.Permissions = model.PermissionsPrint
	case "readonly", "read":
		conf.Permissions = model.PermissionsNone + model.PermissionExtract + model.PermissionExtractRev3
	case "forms":
		conf.Permissions = model.PermissionsNone + model.PermissionFillRev3 + model.PermissionModAnnFillForm
	case "annotate":
		conf.Permissions = model.PermissionsNone + model.PermissionModAnnFillForm + model.PermissionFillRev3
	case "modify":
		conf.Permissions = model.PermissionsNone + model.PermissionModify + model.PermissionModAnnFillForm +
			model.PermissionFillRev3 + model.PermissionAssembleRev3
	default:
		conf.Permissions = model.PermissionsPrint
	}

	err := api.SetPermissionsFile(inFile, outFile, conf)
	return err
}

// AddProperties sets document properties for a PDF file
func AddProperties(inFile string, outFile string, userPass string, ownerPass string, properties map[string]string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err := api.AddPropertiesFile(inFile, outFile, properties, conf)
	return err
}

// RemoveProperties removes specified document properties from a PDF file
func RemoveProperties(inFile string, outFile string, userPass string, ownerPass string, properties []string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPass
	conf.OwnerPW = ownerPass

	err := api.RemovePropertiesFile(inFile, outFile, properties, conf)
	return err
}

// ChangeUserPassword changes the user password of an encrypted PDF
func ChangeUserPassword(inFile string, outFile string, oldPassword string, newPassword string, ownerPassword string) error {
	conf := model.NewDefaultConfiguration()
	conf.OwnerPW = ownerPassword

	err := api.ChangeUserPasswordFile(inFile, outFile, oldPassword, newPassword, conf)
	return err
}

// ChangeOwnerPassword changes the owner password of an encrypted PDF
func ChangeOwnerPassword(inFile string, outFile string, oldPassword string, newPassword string, userPassword string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPassword

	err := api.ChangeOwnerPasswordFile(inFile, outFile, oldPassword, newPassword, conf)
	return err
}

// Helper function to make temporary directory
func MakeTempDir() string {
	seed := strconv.Itoa(os.Getpid())
	tempDir := seed + "-temp"
	err := os.MkdirAll(tempDir, 0700)
	if err != nil {
		log.Printf("Error creating directory: %v\n", err)
		return ""
	}
	return tempDir
}

// ReadFilesInDir reads all files in a directory and returns their paths
func ReadFilesInDir(dirPath string) ([]string, error) {
	var contents []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(dirPath, entry.Name())
		contents = append(contents, filePath)
	}
	return contents, nil
}

// ZeroPadNames renames files to have zero-padded numbers
func ZeroPadNames(dir string, pad int) error {
	re := regexp.MustCompile(`^.*_(\d+)\.pdf$`)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		name := info.Name()
		matches := re.FindStringSubmatch(name)
		if len(matches) == 2 {
			num, err := strconv.Atoi(matches[1])
			if err != nil {
				return err
			}

			paddedNum := fmt.Sprintf("%0*d", pad, num)
			re := regexp.MustCompile(`_.*\.pdf$`)
			newName := re.ReplaceAllString(name, fmt.Sprintf("_%s.pdf", paddedNum))
			oldPath := filepath.Join(dir, name)
			newPath := filepath.Join(dir, newName)

			if oldPath != newPath {
				return os.Rename(oldPath, newPath)
			}
		}
		return nil
	})
}

// PathExists checks if a file or directory exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

