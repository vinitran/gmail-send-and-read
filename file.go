package main

import (
	"bufio"
	"fmt"
	"os"
)

func CreateFileAndWrite(path string, text string) error {
	path = checkDuplicateFile(path)
	err := createFile(path)
	if err != nil {
		return err
	}
	err = writeFile(path, text)
	if err != nil {
		return err
	}

	return nil
}

func createFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	f.Close()
	return nil
}

func CreateFolder(path string) (string, error) {
	path = checkDuplicateFile(path)
	err := os.Mkdir(path, 0755)
	if err != nil {
		return "", err
	}
	return path, nil
}

func writeFile(path string, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	textWriter := bufio.NewWriter(file)

	_, err = textWriter.WriteString(text)
	if err != nil {
		return err
	}
	textWriter.Flush()
	fmt.Println("Data written to file successfully...")
	return nil
}

func checkDuplicateFile(path string) string {
	duplicate := 0
	for {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// path does not exist
			break
		}
		duplicate++
		if duplicate == 1 {
			path = fmt.Sprintf("%s(%d)", path, duplicate)
			continue
		}
		path = fmt.Sprintf("%s(%d)", path[0:len(path)-3], duplicate)
	}
	return path
}
