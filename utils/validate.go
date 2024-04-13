package utils

import "errors"

func ValidateProjectId(projectId string) error {
	if len(projectId) < 6 || len(projectId) > 30 {
		return errors.New("project ID must be between 6 and 30 characters")
	}

	for _, char := range projectId {
		if !isValidCharacter(char) {
			return errors.New("project ID can only contain lowercase alphanumeric characters and hyphen (-)")
		}
	}

	return nil
}

func isValidCharacter(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-'
}
