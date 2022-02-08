package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const (
	defaultTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>{{ .Title }}</title>
  </head>
  <body>
{{ .Body }}
  </body>
</html>`
)

type content struct {
	Title string
	Body template.HTML
}

func main() {
	// Parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	tFName := flag.String("t", "", "Alternate template name")
	flag.Parse()

	// If user did not provide input file, show usage
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *tFName, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filename, tFName string, out io.Writer, skipPreview bool) error {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, tFName)
	if err != nil {
		return err
	}

	// Create temporary file and check for errors
	temp, err := ioutil.TempFile("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()

	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}

	defer os.Remove(outName)

	return preview(outName)
}

func parseContent(input []byte, tFName string) ([]byte, error) {
	// Parse the markdown file through blackfriday and bluemonday
	// to generate a valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	// Parse the contents of the defaultTemplate const into a new Template
	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	// If user provided alternate template file, replace template
	if tFName != "" {
		t, err = template.ParseFiles(tFName)
		if err != nil {
			return nil, err
		}
	}

	// Instantiate the content type, adding the title and body
	c := content{
		Title: "Markdown Preview Tool",
		Body: template.HTML(body),
	}

	// Create a buffer of bytes to write to file
	var buffer bytes.Buffer

	// Execute the template with the content type
	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func saveHTML(outFname string, data []byte) error {
	return ioutil.WriteFile(outFname, data, 0644)
}

func preview(fname string) error {
	cName := ""
	cParams := []string{}
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	// Append filename to parameters slice
	cParams = append(cParams, fname)

	// Locate executable in PATH
	cPath, err := exec.LookPath(cName)

	if err != nil {
		return err
	}

	// Open the file using default program
	err = exec.Command(cPath, cParams...).Run()

	time.Sleep(2 * time.Second)
	return err
}