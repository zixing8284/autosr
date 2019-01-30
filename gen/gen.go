// This file is part of autosr.
//
// autosr is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// autosr is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with autosr.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

func main() {
	fmt.Println("gen.go")

	tag, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		panic(err)
	}
	version := strings.Trim(string(tag), " \n")

	f, err := os.Create("version/version.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Println("[gen] version/version.go")
	versionTmpl.Execute(f, struct {
		Version string
	}{
		Version: version,
	})
}

var versionTmpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package version

// String is the version string
var String = "{{ .Version }}"
`))
