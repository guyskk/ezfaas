package common

import (
	"fmt"
	"os"
	gofilepath "path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
)

func ReadUserFile(filepath string) ([]byte, error) {
	filepath, err := homedir.Expand(filepath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(gofilepath.ToSlash(filepath))
}

var ErrCanceled = fmt.Errorf("canceled")

func Comfirm(label string) bool {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	return err == nil
}

func ComfirmDeploy() bool {
	return Comfirm("Confirm Deploy")
}
