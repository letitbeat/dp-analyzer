package dot

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
)

// Generate generates a dot image and returns the base64 representation
// of the generated graph.
func Generate(filename, dot string) (string, error) {

	err := writeStringToFile(filename, dot)
	if err != nil {
		return "", fmt.Errorf("error writing DOT file %s, err %s", filename, err.Error())
	}

	str, err := generate(filename)
	if err != nil {
		return "", fmt.Errorf("error generating DOT file: %s, content: %s, err %s", filename, dot, err.Error())
	}
	return str, nil
}

// Merge merges a given set of dot files and returns the merged representation
// of it using external `gvpack` command.
func Merge(filenames []string) (string, error) {
	str, err := merge(filenames)
	if err != nil {
		return "", fmt.Errorf("error merging DOT files %s, err %s", filenames, err.Error())
	}
	return str, nil
}

// Write stores in filesystem a dot string
func Write(filename string, str string) error {
	return writeStringToFile(filename, str)
}

func writeStringToFile(filename string, str string) error {

	os.Remove(filename)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(str); err != nil {
		return err
	}
	f.Close()

	return nil
}

func generate(filename string) (string, error) {

	cmd := exec.Command("dot", "-Tpng", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func merge(files []string) (string, error) {
	args := []string{"-guv"}
	args = append(args, files...)

	cmd := exec.Command("gvpack", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return string(out.Bytes()), nil
}
