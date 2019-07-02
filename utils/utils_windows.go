// +build windows

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"cord.stool/service/models"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// CompleteUpgrade ...
func CompleteUpgrade(fnames []string) error {

	var args []string
	args = append(args, os.Args[0])
	args = append(args, "upgrade_complete")

	for _, f := range fnames {
		args = append(args, "-f")
		args = append(args, f)
	}

	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	os.StartProcess(os.Args[0], args, procAttr)

	return nil
}

func AddRegKeys(manifest *models.ConfigManifest) error {

	for _, rk := range manifest.RegistryKeys {

		k, _, err := registry.CreateKey(registry.CURRENT_USER, rk.Key, registry.WRITE)
		if err != nil {
			return nil
		}
		defer k.Close()

		err = k.SetStringValue(rk.Name, rk.Value)
		if err != nil {
			return nil
		}
	}

	return nil
}

func CheckCompletion(regKey models.ConfigRegistryKey) (bool, error) {

	k, err := registry.OpenKey(registry.CURRENT_USER, regKey.Key, registry.QUERY_VALUE)
	if err != nil {
		return false, nil
	}
	defer k.Close()

	s, _, err := k.GetStringValue(regKey.Name)
	if err != nil {
		return false, nil
	}

	return s == regKey.Value, nil
}

func stringToUintptr(str string) uintptr {

	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str)))
}

func RunCommand(admin bool, name string, arg ...string) error {

	if !admin {
		return exec.Command(name, arg...).Run()
	}

	dll, err := syscall.LoadDLL("Shell32.dll")
	if err != nil {
		return err
	}
	defer dll.Release()

	ShellExecute, err := dll.FindProc("ShellExecuteW")
	if err != nil {
		return err
	}

	cmdLine := ""
	for _, a := range arg {

		cmdLine = cmdLine + a + " "
	}

	fpath, _ := filepath.Split(name)
	_, _, err = ShellExecute.Call(uintptr(unsafe.Pointer(nil)), stringToUintptr("runas"), stringToUintptr(name), stringToUintptr(cmdLine), stringToUintptr(fpath), 1)

	if err != windows.DS_S_SUCCESS {
		return err
	}

	return nil
}
