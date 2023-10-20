package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Path string
	// Add more fields as needed
}

func printStruct(data []struct{ word, value string }) {
	if len(data) == 0 {
		return
	}

	// Find the length of the longest word
	maxWordLen := len(data[0].word)
	for _, item := range data {
		if len(item.word) > maxWordLen {
			maxWordLen = len(item.word)
		}
	}

	// Print the data with aligned spacing
	for _, item := range data {
		fmt.Printf("%-*s %s\n", maxWordLen, item.word, item.value)
	}
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err
}

func moveFiles(sourceDir, destinationDir string) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return err
	}

	// Read the contents of the source directory
	fileInfos, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		sourcePath := filepath.Join(sourceDir, fileInfo.Name())
		destinationPath := filepath.Join(destinationDir, fileInfo.Name())

		if fileInfo.IsDir() {
			// Recursively move directories
			if err := moveFiles(sourcePath, destinationPath); err != nil {
				return err
			}
		} else {
			// Move files
			if err := moveFile(sourcePath, destinationPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func moveFile(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// if err := os.Remove(sourcePath); err != nil {
	//     return err
	// }

	//fmt.Printf("Moved: %s to %s\n", sourcePath, destinationPath)
	return nil
}

func isDirectoryEmpty(directoryPath string) (bool, error) {
	// Read the contents of the directory
	entries, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return false, err
	}

	// Check if there are no entries (directory is empty)
	return len(entries) == 0, nil
}

func main() {
	args := os.Args

	executableName := filepath.Base(args[0])

	if len(args) < 2 {
		fmt.Printf("Usage: %s <command>\n", executableName)
		cmds := []struct {
			word, value string
		}{
			{"init <path>", "Set wallymover's path"},
			{"install", "Run wally install with wallymover"},
		}
		printStruct(cmds)
		return
	}

	command := args[1]

	if command == "init" {
		if len(args) > 2 {
			packagesPath := args[2]

			//println(packagesPath)

			config := Config{
				Path: packagesPath,
			}

			data, err := toml.Marshal(config)
			if err != nil {
				panic(err)
			}

			// Write the encoded data to a file
			err = ioutil.WriteFile("wallymover.toml", data, 0644)
			if err != nil {
				panic(err)
			}
			println("wallymover: successfully initialized wallymover.toml")
			return
		} else {
			log.Fatal("argument <path> is not given.")
		}
	} else if command == "install" {
		file, err := os.Open("wallymover.toml")
		if err != nil {
			log.Fatal("wallymover.toml is not found. please run 'wallymover init [your packages folder path]' first")
		}
		defer file.Close()

		var config Config

		b, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}

		err = toml.Unmarshal(b, &config)
		if err != nil {
			panic(err)
		}

		isEmpty, err := isDirectoryEmpty(config.Path)
		if err != nil {
			panic(err)
		}
		if isEmpty {

		} else {
			err := os.Mkdir("Packages", 0755)
			if err != nil {
				panic(err)
			}
			err = moveFiles(config.Path, "Packages")
			if err != nil {
				panic(err)
			}
			err = os.RemoveAll(config.Path)
			if err != nil {
				log.Fatal("remove wallymover.Path error:", err)
			}
		}

		err = runCommand("wally", "install")
		if err != nil {
			log.Fatal("wally error:", err) // change to Fatal after the test
		}

		err = moveFiles("Packages", config.Path)
		if err != nil {
			log.Fatal("move files error:", err)
		}

		err = os.RemoveAll("Packages")
		if err != nil {
			log.Fatal("remove Packages error:", err)
		}

		println("wallymover: successfully installed and moved with wallymover!")
		return
	}
	commands := os.Args[1:]
	err := runCommand("wally", commands...)
	if err != nil {
		log.Fatal("wally error:", err) // change to Fatal after the test
	}
}
