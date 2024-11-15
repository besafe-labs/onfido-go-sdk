package utils

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv(file string) {
	if err := godotenv.Load(file); err != nil {
		fmt.Printf("\033[33m an error occurred while loading .env file\033[0m: \n\t \033[0;31m %s \033[0m\n", err)
	}
}

// IsTestRun returns true if code is running as part of a test
func IsTestRun() bool {
	return flag.Lookup("test.v") != nil || flag.Lookup("test.run") != nil
}
