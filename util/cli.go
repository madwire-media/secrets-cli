package util

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// CliQuestionYesNo prints a question prompt and allows either yes or no answers
// to be entered, returning "y" or "yes" as true and "n" or "no" as false
func CliQuestionYesNo(question string) bool {
	for {
		answer := CliQuestion(question + " (y/n)")

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please only enter 'y' or 'n'")
		}
	}
}

// CliQuestionYesNoDefault prints a question prompt and allows either yes or no
// answers to be entered, or no answer at all. If no answer is entered, the
// defaultValue is returned, otherwise "y" or "yes" returns true and "n" or "no"
// returns false
func CliQuestionYesNoDefault(question string, defaultValue bool) bool {
	var yn string

	if defaultValue == true {
		yn = " (Y/n)"
	} else {
		yn = " (y/N)"
	}

	for {
		answer := CliQuestion(question + yn)

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "":
			return defaultValue
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter 'y', 'n', or leave the line empty")
		}
	}
}

// CliQuestion prints a prompt to stdout and reads a line of input from stdin,
// returning that read string
func CliQuestion(question string) string {
	var answer string

	fmt.Printf("%s: ", question)
	fmt.Scanf("%s", &answer)

	return answer
}

// CliQuestionHidden prints a prompt to stdout and reads a hidden line of input
// from stdin. This is meant to be used for passwords where you don't want them
// printed to the screen
func CliQuestionHidden(question string) (string, error) {
	fmt.Printf("%s (hidden): ", question)

	rawAnswer, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(rawAnswer), nil
}
