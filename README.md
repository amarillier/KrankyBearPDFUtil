# pdfutil - PDF Management Made Simple

![KrankyBearBeret](https://github.com/user-attachments/assets/95aef02f-a72c-4b82-aad2-5d2e4b30315a)

A powerful, easy-to-use command-line PDF utility built with [pdfcpu](https://github.com/pdfcpu/pdfcpu) by Horst H Rutter.

## Highlights

- **🔐 Encryption & Security** - Protect PDFs with passwords and granular permissions
- **📑 Page Manipulation** - Extract, remove, split, rotate, reverse, and merge pages
- **🔒 Permission Control** - Set precise access restrictions (print, modify, extract, forms, etc.)
- **📊 PDF Analysis** - View detailed metadata and permission breakdowns
- **⚡ Simple & Fast** - Straightforward commands with helpful aliases
- **🆓 100% Free** - Open source, no restrictions

### Why pdfutil?

This tool provides **simple, focused PDF operations** with clear syntax and helpful output. While [pdfcpu](https://pdfcpu.io/) offers comprehensive PDF manipulation, pdfutil is designed for:

- Quick, common PDF tasks without complexity
- Easy-to-remember command structure
- Visual permission displays and clear feedback
- Unique features like full page reversal
- Future GUI interface (planned)

---

## Table of Contents
- [Features](#features)
- [Quick Start](#quick-start)
- [Command Examples](#command-examples)
  - [File Information](#-file-information)
  - [Encryption & Decryption](#-encryption--decryption)
  - [Permissions Management](#-permissions-management)
  - [Document Properties Management](#-document-properties-management)
  - [Page Operations](#-page-operations)
  - [Merge & Insert](#-merge--insert)
- [Advanced Usage Tips](#advanced-usage-tips)
- [Quick Reference](#quick-reference)

---

## Features

pdfutil provides the following PDF operations:

* **Check for updates** - Manual update checking (no auto-updates)
* **Change passwords** - Change user or owner passwords on encrypted PDFs
* **Decrypt files** - Remove password protection from PDFs
* **Encrypt files** - Add password protection with AES or RC4 encryption
* **Extract pages** - Extract specific pages or ranges to separate files
* **File information** - View detailed PDF metadata and properties
* **Insert pages** - Insert one PDF into another at any position
* **Merge files** - Combine multiple PDFs into a single file
* **Remove pages** - Delete specific pages or ranges from a PDF
* **Reverse pages** - Reverse the order of all pages (last becomes first)
* **Rotate pages** - Rotate pages by 90, 180, or 270 degrees
* **View permissions** - Display detailed permission breakdown with visual formatting
* **Set permissions** - Apply granular access controls (print, modify, extract, forms, etc.)
* **Document properties** - List, add, and remove PDF metadata (title, author, subject, etc.)
* **Split pages** - Split a PDF into individual page files

### Future Additions
* Watermark - Apply watermarks to pages
* GUI interface - Graphical user interface

---

## Quick Start

Get help on all available commands:
```bash
pdfutil --help
```

Check for updates:
```bash
pdfutil checkupdate
# or
pdfutil cu
```

---

## Command Examples

### 📄 File Information
View PDF metadata, encryption status, and permissions:
```bash
# Basic info (unencrypted file)
pdfutil info -in document.pdf

# Info on encrypted file
pdfutil fileinfo -in encrypted.pdf -up userpass -op ownerpass

# Aliases: info, fi, fileinfo, i, inf
```

### 🔐 Encryption & Decryption

**Encrypt a PDF:**
```bash
# AES-256 encryption (recommended)
pdfutil encrypt -in document.pdf -up userpass -op ownerpass -eb 256

# AES-128 encryption
pdfutil en -in document.pdf -up myuser -op myowner -eb 128

# RC4 encryption
pdfutil encrypt -in document.pdf -op ownerpass -mo rc4 -eb 128

# Aliases: encrypt, en, enc, e
```

**Decrypt a PDF:**
```bash
# Decrypt with owner password
pdfutil decrypt -in encrypted.pdf -op ownerpass

# Decrypt with user password
pdfutil de -in encrypted.pdf -up userpass

# Aliases: decrypt, de, dec, d
```

**Change Passwords:**
```bash
# Change user password (requires owner password)
pdfutil changepassword -in encrypted.pdf -password user -old olduser123 -new newuser456 -op ownerpass

# Change owner password (requires user password)
pdfutil cp -in encrypted.pdf -password owner -old oldowner123 -new newowner456 -up userpass

# Create new file with changed password
pdfutil changepassword -in encrypted.pdf -out newfile.pdf -password user -old olduser123 -new newuser456 -op ownerpass

# Aliases: changepassword, cp, chpwd, changepwd
```

**Password Change Options:**
- `-password user` - Change user password (requires `-op` owner password)
- `-password owner` - Change owner password (requires `-up` user password)
- `-old` / `-oldpass` - Current password
- `-new` / `-newpass` - New password
- `-out` / `-outfile` - Output file (optional, modifies in-place if omitted)

### 🔒 Permissions Management

**View Permissions (Enhanced Display):**
```bash
# View permissions on any PDF
pdfutil permissions -in document.pdf

# View permissions on encrypted PDF
pdfutil perm -in encrypted.pdf -op ownerpass

# Aliases: permissions, perm, perms, p, pe
```

**Example Output:**
```
Raw Permission Value: 0xF0C3 (-3901)

Permission Details
Print Document:         = ALLOWED       
Print High Quality:     = ALLOWED       
Modify Document:        ! DENIED        
Copy/Extract Content:   ! DENIED        
Extract (Rev 3+):       ! DENIED        
Annotate/Comment:       ! DENIED        
Fill Form Fields:       ! DENIED        
Assemble Document:      ! DENIED        
```

**Set Permissions (Must be encrypted first):**
```bash
# In-place modification (recommended - no output file needed)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt print

# Read-only (view and extract only)
pdfutil stp -in encrypted.pdf -op ownerpass -pt readonly

# Allow form filling only
pdfutil setp -in encrypted.pdf -op ownerpass -pt forms

# Allow annotations/comments
pdfutil setperms -in encrypted.pdf -op ownerpass -pt annotate

# Full editing permissions
pdfutil setperms -in encrypted.pdf -op ownerpass -pt modify

# No permissions (maximum restriction)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt none

# Full access (no restrictions)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt all

# Create new file with different permissions (optional)
pdfutil setperms -in encrypted.pdf -out restricted.pdf -op ownerpass -pt readonly

# Aliases: setperms, setperm, stp, setp
```

**Permission Types:**

*Pre-defined profiles:*
- `all` - Full access (no restrictions)
- `none` - All permissions denied
- `print` - Print only (draft + high quality)
- `readonly` - View and extract content only
- `forms` - Fill form fields only
- `annotate` - Add annotations and comments
- `modify` - Full editing permissions

*Individual permissions (can be combined with commas):*
- `print` / `printdraft` - Draft print quality
- `printhq` / `printquality` - High quality print
- `modify` / `edit` - Modify document content
- `extract` / `copy` - Copy and extract content
- `annotate` / `comment` - Add annotations/comments
- `fillforms` / `forms` - Fill form fields
- `assemble` - Assemble document (reorder pages, etc.)

**Custom Permission Combinations:**
```bash
# Your specific example: Print (draft) + Print HQ + Annotate + Copy/Extract
pdfutil setperms -in encrypted.pdf -op ownerpass -pt print,printhq,annotate,extract

# Print + annotate + extract (simplified - 'print' includes both draft and HQ)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt print,annotate,extract

# High-quality print + copy only
pdfutil setperms -in encrypted.pdf -op ownerpass -pt printhq,copy

# Forms + annotations (common for interactive PDFs)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt fillforms,annotate

# Print + extract + assemble
pdfutil setperms -in encrypted.pdf -op ownerpass -pt print,extract,assemble

# Only high-quality print (no draft)
pdfutil setperms -in encrypted.pdf -op ownerpass -pt printhq

# Everything except modify
pdfutil setperms -in encrypted.pdf -op ownerpass -pt print,printhq,extract,annotate,fillforms,assemble
```

### 📄 Document Properties Management

**List Properties:**
```bash
# List all document properties
pdfutil properties list -in document.pdf

# List properties of encrypted PDF
pdfutil props list -in encrypted.pdf -up userpass -op ownerpass

# Aliases: properties, props, prop
```

**Add/Set Properties:**
```bash
# Set title and author
pdfutil properties add -in document.pdf -out result.pdf -title "My Document" -author "John Doe"

# Set multiple properties
pdfutil props add -in document.pdf -out result.pdf -title "Report" -author "Jane Smith" -subject "Q4 Report" -keywords "finance,quarterly"

# Set all properties
pdfutil properties add -in document.pdf -out result.pdf -title "Document" -author "Author" -subject "Subject" -keywords "key,words" -creator "Creator" -producer "Producer"

# Aliases: add, a, set
```

**Remove Properties:**
```bash
# Remove specific properties
pdfutil properties remove -in document.pdf -out result.pdf -prop title,author

# Remove all properties
pdfutil props remove -in document.pdf -out result.pdf -prop title,author,subject,keywords,creator,producer

# Aliases: remove, rm, del
```

**Available Properties:**
- `title` - Document title
- `author` - Document author
- `subject` - Document subject
- `keywords` - Document keywords
- `creator` - Document creator
- `producer` - Document producer

**Example Output (List Properties):**
```
=== Document Properties for: document.pdf ===

📄 Document Information:
├─ Title: My Document
├─ Author: John Doe
├─ Subject: (not set)
├─ Keywords: (not set)
├─ Creator: Microsoft Word
├─ Producer: Microsoft Office
├─ Creation Date: (not set)
└─ Modification Date: (not set)

📋 PDF Version: 1.4

📊 Page Count: 5

🔓 Encryption: Disabled
```

### 📑 Page Operations

**Extract Pages:**
```bash
# Extract specific pages
pdfutil extract -in document.pdf -pg 1,3,5 -od output/

# Extract page range
pdfutil ex -in document.pdf -pg 1-10 -od output/

# Extract from page 5 to last page
pdfutil extract -in document.pdf -pg 5-l -od output/

# Extract with zero-padded filenames
pdfutil ex -in document.pdf -pg 1-20 -od output/ -pad

# Complex selection
pdfutil extract -in document.pdf -pg 1,3-5,10,15-l -od output/

# Aliases: extract, ex, ext, x
```

**Remove Pages:**
```bash
# Remove specific pages
pdfutil remove -in document.pdf -out result.pdf -pg 2,4,6

# Remove page range
pdfutil rem -in document.pdf -out result.pdf -pg 10-20

# Remove from page 50 to end
pdfutil del -in document.pdf -out result.pdf -pg 50-l

# Complex removal
pdfutil remove -in document.pdf -out result.pdf -pg 1,5-10,15,20-l

# Aliases: remove, rem, del, delete, cut
```

**Split Pages:**
```bash
# Split all pages into separate files
pdfutil split -in document.pdf -od output/

# Split with zero-padded filenames (recommended for sorting)
pdfutil sp -in document.pdf -od output/ -pad

# Aliases: split, sp, spl, s
```

**Rotate Pages:**
```bash
# Rotate specific pages 90 degrees
pdfutil rotate -in document.pdf -out rotated.pdf -pg 1,3,5 -ro 90

# Rotate all pages 180 degrees
pdfutil rot -in document.pdf -out rotated.pdf -pg 1-l -ro 180

# Rotate page range 270 degrees
pdfutil ro -in document.pdf -out rotated.pdf -pg 10-20 -ro 270

# Aliases: rotate, ro, rot, r
```

**Reverse Pages:**
```bash
# Reverse all pages (last becomes first)
pdfutil reverse -in document.pdf -out reversed.pdf

# Aliases: reverse, re, rev
```

### 🔗 Merge & Insert

**Merge Multiple PDFs:**
```bash
# Merge multiple files
pdfutil merge -mf file1.pdf,file2.pdf,file3.pdf -out merged.pdf

# Merge with spaces in filenames (use quotes)
pdfutil join -mf "file 1.pdf,file 2.pdf,file 3.pdf" -out result.pdf

# Aliases: merge, me, mer, join, m
```

**Insert PDF Between Pages:**
```bash
# Insert a PDF after page 5
pdfutil insert -in main.pdf -if insert.pdf -out result.pdf -ia 5

# Insert at beginning (after page 0)
pdfutil ins -in main.pdf -if cover.pdf -out result.pdf -ia 0

# Aliases: insert, ins, in
```

---

## Advanced Usage Tips

### Page Selection Syntax
- **Single pages:** `1` or `5` or `10`
- **Ranges:** `1-5` (pages 1 through 5)
- **Last page:** `l` or `10-l` (page 10 to last)
- **Combined:** `1,3-5,10,15-l` (comma-separated, no spaces)

### Password Management
- **User password:** Allows opening and viewing the PDF
- **Owner password:** Allows full access including changing permissions
- Either password can be used for most operations
- **Changing passwords:** Use `changepassword` command with `-password user|owner` switch
- **Password change requirements:** 
  - To change user password: need current owner password
  - To change owner password: need current user password

### Encryption Recommendations
- **AES-256** - Most secure, recommended for sensitive documents
- **AES-128** - Good balance of security and compatibility
- **RC4-128** - Legacy compatibility only (not recommended)
- **RC4-40** - Very weak, avoid unless required for old software

### Permission Workflow
1. **Encrypt** the PDF first (required for permissions)
2. **Set permissions** using the `setperms` command (modifies in-place)
3. **Verify** with the `permissions` command

Example workflow:
```bash
# Step 1: Encrypt
pdfutil encrypt -in document.pdf -up user123 -op owner123 -eb 256

# Step 2: Set permissions (in-place modification)
pdfutil setperms -in document.pdf -op owner123 -pt readonly

# Step 3: Verify
pdfutil permissions -in document.pdf -op owner123

# Alternative: Create a new file with different permissions
pdfutil setperms -in document.pdf -out restricted.pdf -op owner123 -pt print
```

---

## Quick Reference

### Command Cheat Sheet

| Operation | Command | Example |
|-----------|---------|---------|
| **Get help** | `pdfutil --help` | Show all commands |
| **Check updates** | `pdfutil cu` | Check for new version |
| **Info** | `pdfutil info -in file.pdf` | View PDF metadata |
| **Encrypt** | `pdfutil encrypt -in file.pdf -op pass -eb 256` | Encrypt with AES-256 |
| **Decrypt** | `pdfutil decrypt -in file.pdf -op pass` | Remove encryption |
| **Change user pass** | `pdfutil cp -in file.pdf -password user -old old -new new -op owner` | Change user password |
| **Change owner pass** | `pdfutil cp -in file.pdf -password owner -old old -new new -up user` | Change owner password |
| **View perms** | `pdfutil perm -in file.pdf` | Show permissions |
| **Set perms** | `pdfutil setperms -in file.pdf -op pass -pt readonly` | Set read-only (in-place) |
| **List props** | `pdfutil props list -in file.pdf` | Show document properties |
| **Add props** | `pdfutil props add -in file.pdf -out new.pdf -title "Title" -author "Author"` | Set properties |
| **Remove props** | `pdfutil props remove -in file.pdf -out new.pdf -prop title,author` | Remove properties |
| **Extract** | `pdfutil extract -in file.pdf -pg 1-10 -od out/` | Extract pages 1-10 |
| **Remove** | `pdfutil remove -in file.pdf -out new.pdf -pg 5-10` | Remove pages 5-10 |
| **Split** | `pdfutil split -in file.pdf -od out/ -pad` | Split all pages |
| **Merge** | `pdfutil merge -mf a.pdf,b.pdf,c.pdf -out merged.pdf` | Merge 3 files |
| **Rotate** | `pdfutil rotate -in file.pdf -out new.pdf -pg 1-l -ro 90` | Rotate all 90° |
| **Reverse** | `pdfutil reverse -in file.pdf -out new.pdf` | Reverse page order |
| **Insert** | `pdfutil insert -in main.pdf -if insert.pdf -out new.pdf -ia 5` | Insert after page 5 |

### Common Command Aliases

| Full Command | Short Aliases |
|--------------|---------------|
| `checkupdate` | `cu`, `chk`, `c` |
| `changepassword` | `cp`, `chpwd`, `changepwd` |
| `decrypt` | `de`, `dec`, `d` |
| `encrypt` | `en`, `enc`, `e` |
| `extract` | `ex`, `ext`, `x` |
| `fileinfo` | `info`, `fi`, `i`, `inf` |
| `insert` | `ins`, `in` |
| `merge` | `me`, `mer`, `join`, `m` |
| `permissions` | `perm`, `perms`, `p`, `pe` |
| `setperms` | `setperm`, `stp`, `setp` |
| `remove` | `rem`, `del`, `delete`, `cut` |
| `reverse` | `re`, `rev` |
| `rotate` | `ro`, `rot`, `r` |
| `split` | `sp`, `spl`, `s` |

### Common Flag Aliases

| Full Flag | Short Aliases |
|-----------|---------------|
| `-infile` | `-in`, `-i` |
| `-outfile` | `-out`, `-o` |
| `-outdir` | `-od`, `-o` |
| `-userpass` | `-up`, `-u` |
| `-ownerpass` | `-op`, `-o` |
| `-pages` | `-pg`, `-p`, `-page`, `-pagelist` |
| `-encryptionbits` | `-eb`, `-e` |
| `-permtype` | `-pt`, `-perm` |
| `-rotate` | `-ro`, `-rot`, `-r` |
| `-zeropad` | `-pad`, `-z` |

---

## To-Do / Known Issues

- File operations do not work on encrypted PDFs (decrypt first)
- Windows: Icon may not display properly in some contexts

---

## License

See LICENSE file for details.

# License
This is 100% free for anyone to use or misuse any way you like with no warranty as
to suitability or anything else, other than it has no viruses when I compile and
commit to git. But you should always check and scan anything you download from the
internet for viruses anyway. Don't be reckless.

All KrankyBear icons, images, logos used are copyright (c) Allan Marillier, 2024, 2025 ...

Keep copies of your files and test features until you know how they work and trust the application.


In other words, I take no responsibility for how you use this, protect yourself. 
