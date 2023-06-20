package file

import (
	"fmt"
	"os"
)

func Create(dir, filename, content string) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return err
	}

	_, err = file.WriteString(content)
	return err
}
