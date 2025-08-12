// SPECIAL: update method below to match repo/prefix
package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/TravisStratton/terraform-provider-awx/internal/configprefix"
)

func main() {

	prefix := configprefix.Prefix
	if prefix == "" {
		os.Exit(1)
	}

	// where to start looking for templates

	root := "examples/templates/" // Starting point

	// walk the file system recursively at root
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		afterRoot := strings.Replace(path, root, "", 1)

		dir, name := filepath.Split(filepath.Join(root, "..", afterRoot))

		// is this a directory we need to create?
		if d.IsDir() && (strings.ContainsRune(afterRoot, os.PathSeparator) || afterRoot == "provider") {

			var makeDirName string

			if afterRoot == "provider" {
				makeDirName = filepath.Join(dir, name)
			} else {
				makeDirName = filepath.Join(dir, configprefix.Prefix+"_"+name)
			}

			err := os.MkdirAll(makeDirName, 0755)
			if err != nil {
				fmt.Println("Error creating output directory:", err)
				return err
			}
		}

		if !d.IsDir() {

			parts := strings.Split(dir, string(os.PathSeparator))

			// we need to add the prefix + _ to the containing directory of the file, unless it's the "provider" folder

			index := (len(parts) - 2)

			if parts[index] != "provider" {
				newString := configprefix.Prefix + "_" + parts[index]
				parts[index] = newString
			}

			// where will the resulstant file after parsing the template be
			parsedFileName := filepath.Join(append(parts, strings.TrimSuffix(name, ".tmpl"))...)

			// Parse and execute template
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				fmt.Println("Error parsing template:", err)
				return err
			}

			f, err := os.Create(parsedFileName)
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return err
			}
			defer f.Close()

			err = tmpl.Execute(f, map[string]string{"Prefix": prefix, "ProviderSource": "TravisStratton/awx"}) // SPECIAL: update to match THIS repo's name & provider prefix
			if err != nil {
				fmt.Println("Error executing template:", err)
				return err
			}

		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking the path:", err)
		os.Exit(1)
	}

}
