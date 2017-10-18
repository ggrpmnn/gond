package main

import (
	// builtins
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	// user packages
	"github.com/fatih/color"
)

const (
	// ErrIn denotes bad input (via the arg flags)
	ErrIn int = 0x1
	// ErrSys denotes a system error
	ErrSys int = 0x2
)

func main() {
	// process and parse command flags
	var (
		name, dir, sep string
		pad            int
		conf, inc      bool
	)
	flag.StringVar(&name, "n", "", "The string to use as the base `filename`.")
	flag.StringVar(&dir, "d", "", "The `directory` in which to rename files.")
	flag.StringVar(&sep, "s", "", "The string to use as a `separator`.")
	flag.IntVar(&pad, "p", 0, "The number of digits used to `pad` the filenumber.")
	flag.BoolVar(&inc, "i", false, "Include directories in the rename operation.")
	flag.BoolVar(&conf, "c", false, "Don't ask for confirmation to rename the files.")
	flag.Parse()

	if name == "" {
		if flag.Arg(0) != "" {
			name = flag.Arg(0)
		} else {
			color.Red("ERROR: a base filename string is required.")
			flag.PrintDefaults()
			os.Exit(ErrIn)
		}
	}

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			color.Red("ERROR: could not get current directory path.")
			os.Exit(ErrSys)
		}
	} else {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			color.Red(fmt.Sprintf("ERROR: the supplied directory ('%s') is not a valid path.", dir))
			os.Exit(ErrIn)
		}
		dir, _ = filepath.Abs(dir)
	}
	if sep == "" {
		sep = "-"
	}
	if pad == 0 {
		pad = 1
	}

	// confirm the file rename - skip if -c flag is provided
	if !conf {
		color.Yellow(fmt.Sprintf("Are you certain you want to rename the files using the provided parameters?\nDirectory: %s\nName: %s%s%s.ext", dir, name, sep, strings.Repeat("#", pad)))
		in := bufio.NewReader(os.Stdin)
		fmt.Print("Confirm: ")
		confirmation, err := in.ReadString('\n')
		if err != nil {
			color.Red("ERROR: failed to get user input.")
			os.Exit(ErrIn)
		}
		if !strings.EqualFold(confirmation, "Y\n") && !strings.EqualFold(confirmation, "YES\n") {
			color.Green("Exiting without changing files.")
			os.Exit(0)
		}
	}

	// change to dir, get the list of files in that dir
	os.Chdir(dir)
	fileList, fileErr := filepath.Glob("*")
	if fileErr != nil {
		color.Red("ERROR: failed to get file listing for the target directory.")
		os.Exit(ErrSys)
	}

	// rename files in sequential order; preserving relative order is important!!
	success := true
	fileNum := 1
	for i := range fileList {
		fileName := fileList[i]
		dirTest := isDir(fmt.Sprintf("%s/%s", dir, fileName))
		if !dirTest || (dirTest && inc) {
			fileExt := path.Ext(fileName)
			var err error
			// handle the case where a file is of the form '.filename'
			if fileExt == "" || fileName == fileExt {
				err = os.Rename(fileName, fmt.Sprintf("%s%s%s", name, sep, leftPad(strconv.Itoa(fileNum), "0", pad)))
				if err != nil {
					color.Red(fmt.Sprintf("ERROR: error while renaming file '%s': %s", fileName, err.Error()))
					success = false
				}
			} else {
				err = os.Rename(fileName, fmt.Sprintf("%s%s%s%s", name, sep, leftPad(strconv.Itoa(fileNum), "0", pad), fileExt))
				if err != nil {
					color.Red(fmt.Sprintf("ERROR: error while renaming file '%s': %s", fileName, err.Error()))
					success = false
				}
			}
			fileNum++
		}
		// continue
	}
	if success {
		color.Green("Finished renaming files successfully.")
	} else {
		color.Red("Finished processing files, with errors.")
	}

	return
}

// Determine whether the file is a directory
func isDir(s string) bool {
	// Ignoring the error because we KNOW the specified file exists
	fi, _ := os.Stat(s)
	return fi.IsDir()
}

// Pad a string to the specified length using the pad string
func leftPad(s string, p string, n int) string {
	if len(s) >= n {
		return s
	}
	padLen := n - len(s)
	return fmt.Sprintf("%s%s", strings.Repeat(p, padLen), s)
}
