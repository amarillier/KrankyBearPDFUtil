# pdfutil
![KrankyBearBeret](https://github.com/user-attachments/assets/95aef02f-a72c-4b82-aad2-5d2e4b30315a)

This is a PDF utility, using Horst H Rutter's pdfcpu library
https://github.com/pdfcpu/pdfcpu
https://pdfcpu.io/
If you want simple basics, this version may work for you.
If you want more features, the pdfcpu project also has a CLI (command line interface)
that does all that I do in my simple application and more.

So why would you want mine? Simple use, maybe some features implemented a bit
differently, and a possible GUI option in future. THe choice is yours.

pdfutil provides a number of different actions:
* Check for updates
  -checkupdate or -cu parameter allows checking for updates. This is entirely manual,
  no auto checks, or auto updates will be performed.
* Decrypt files (passwords are required)
* Encrypt files
* Extract selected / all pages from a file
* File information
* Insert one or more pages (pdf/pdfs) into into another pdf
* Merge files
* Rotate pages in a file by 90, 180 or 270 degrees
* Reverse pages in a file (literally saw sequence, first becomes last, last becomes first)
* Set file permissions
* Split all pages from a file into individual pdf pages

* Watermark - possible future addition to apply a watermark to some / all pages
* GUI interface - possible future addition


# To-do / known problems
- Reading permissions may not yet properly return actual print allowed / denied, high def print allowed denied etc
- File info may not work completely with encrypted vs non encrypted files, in progress


# License
This is 100% free for anyone to use or misuse any way you like with no warranty as
to suitability or anything else, other than it has no viruses when I compile and
commit to git. But you should always check and scan anything you download from the
internet for viruses anyway. Don't be reckless.

All KrankyBear icons, images, logos used are copyright (c) Allan Marillier, 2024, 2025 ...

Keep copies of your files and test features until you know how they work and trust the application.


In other words, I take no responsibility for how you use this, protect yourself. 
