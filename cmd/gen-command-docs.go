package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"io/ioutil"

	cmd "code.cloudfoundry.org/quarks-operator/cmd/internal"
	"github.com/spf13/cobra/doc"
)

const (
	index = `---
title: "CLI Options"
linkTitle: "CLI Options"
weight: 20
description: >
    CLI Options
---
	`
	docDir = "docs/commands/"
)

func main() {

	cfOperatorCommand := cmd.NewCFOperatorCommand()
	identity := func(s string) string {
		fmt.Println(s)
		return s
	}
	prepend := func(s string) string {
		title := strings.ReplaceAll(s, docDir, "")
		title = strings.ReplaceAll(title, ".md", "")
		title = strings.ReplaceAll(title, "_", " ")

		return `---
title: "` + title + `"
linkTitle: "` + title + `"
weight: 1
---
`
	}
	err := doc.GenMarkdownTreeCustom(cfOperatorCommand, filepath.Join("./", docDir), prepend, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./docs/commands/_index.md", []byte(index), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
