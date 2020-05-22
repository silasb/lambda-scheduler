package preparable

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/struCoder/pmgo/lib/process"
	"github.com/struCoder/pmgo/lib/utils"
)

// ProcPreparable is a preparable with all the necessary informations to run
// a process. To actually run a process, call the Start() method.
type ProcPreparable interface {
	PrepareBin() ([]byte, error)
	Start() (process.ProcContainer, error)
	SetupProc() (process.ProcContainer, error)
	getPath() string
	Identifier() string
	getBinPath() string
	getPidPath() string
	getOutPath() string
	getErrPath() string
}

type Preparable struct {
	Name       string
	SourcePath string
	Cmd        string
	WorkingDir string
	SysFolder  string
	Language   string
	KeepAlive  bool
	Args       []string
	Envs       []string
}

type BinaryPreparable struct {
	Name       string
	SourcePath string
	Cmd        string
	BZipFile   string
	SysFolder  string
	WorkingDir string
	Language   string
	KeepAlive  bool
	Args       []string
	Envs       []string
}

// PrepareBin will compile the Golang project from SourcePath and populate Cmd with the proper
// command for the process to be executed.
// Returns the compile command output.
func (preparable *Preparable) PrepareBin() ([]byte, error) {
	// Remove the last character '/' if present
	if preparable.SourcePath[len(preparable.SourcePath)-1] == '/' {
		preparable.SourcePath = strings.TrimSuffix(preparable.SourcePath, "/")
	}
	binPath := preparable.getBinPath()

	cmd := ""
	cmdArgs := []string{}
	if preparable.Language == "go" {
		cmd = "go"
		cmdArgs = []string{"build", "-o", binPath, preparable.SourcePath + "/."}
	}

	preparable.Cmd = binPath
	return exec.Command(cmd, cmdArgs...).Output()
}

// Start will execute the process based on the information presented on the preparable.
// This function should be called from inside the master to make sure
// all the watchers and process handling are done correctly.
// Returns a tuple with the process and an error in case there's any.
func (preparable *Preparable) Start() (process.ProcContainer, error) {
	proc := &process.Proc{
		Name:      preparable.Name,
		Cmd:       preparable.Cmd,
		Args:      preparable.Args,
		Path:      preparable.getPath(),
		Pidfile:   preparable.getPidPath(),
		Outfile:   preparable.getOutPath(),
		Errfile:   preparable.getErrPath(),
		KeepAlive: preparable.KeepAlive,
		Status:    &process.ProcStatus{},
	}

	err := proc.Start()
	return proc, err
}

func (preparable *Preparable) SetupProc() (process.ProcContainer, error) {
	proc := &process.Proc{
		Name:      preparable.Name,
		Cmd:       preparable.Cmd,
		Args:      preparable.Args,
		Path:      preparable.getPath(),
		Pidfile:   preparable.getPidPath(),
		Outfile:   preparable.getOutPath(),
		Errfile:   preparable.getErrPath(),
		KeepAlive: preparable.KeepAlive,
		Status:    &process.ProcStatus{},
	}

	// err := proc.Start()
	return proc, nil
}

// Identifier is a function that get proc name
func (preparable *Preparable) Identifier() string {
	return preparable.Name
}

func (preparable *Preparable) getPath() string {
	if preparable.SysFolder[len(preparable.SysFolder)-1] == '/' {
		preparable.SysFolder = strings.TrimSuffix(preparable.SysFolder, "/")
	}
	return preparable.SysFolder + "/" + preparable.Name
}

func (preparable *Preparable) getBinPath() string {
	return preparable.getPath() + "/" + preparable.Name
}

func (preparable *Preparable) getPidPath() string {
	return preparable.getBinPath() + ".pid"
}

func (preparable *Preparable) getOutPath() string {
	return preparable.getBinPath() + ".out"
}

func (preparable *Preparable) getErrPath() string {
	return preparable.getBinPath() + ".err"
}

// BinaryPreparable

// PrepareBin checks if the given binary path is a valid executable.
// Returns no bytes, but if there is an error, it will be returned.
func (preparable *BinaryPreparable) PrepareBin() ([]byte, error) {
	// build directory
	err := os.MkdirAll(filepath.Dir(preparable.getOutPath()), 0755)
	if err != nil {
		return make([]byte, 0), err
	}

	runtimePath := preparable.getPath() + "/runtime"

	err = os.MkdirAll(runtimePath, 0755)
	if err != nil {
		return make([]byte, 0), err
	}

	preparable.Envs = append(preparable.Envs, "LAMBDA_TASK_ROOT="+runtimePath, "PATH=/home/silas/.deno/bin:/usr/bin")

	// unzip file
	decoded, err := base64.StdEncoding.DecodeString(preparable.BZipFile)
	if err != nil {
		log.Printf("decode error:", err)
		return make([]byte, 0), err
	}

	zipPath := preparable.getPath() + "/function.zip"

	err = ioutil.WriteFile(zipPath, []byte(decoded), 0644)
	if err != nil {
		log.Printf("decode error:", err)
		return make([]byte, 0), err
	}

	err = utils.Unzip(zipPath, runtimePath)
	if err != nil {
		log.Printf("error unzipping file: %s at path %s", err, runtimePath)
	}

	// set command
	preparable.Cmd = preparable.getPath() + "/runtime/bootstrap"
	return make([]byte, 0), err
}

// Start will execute the process based on the information presented on the preparable.
// This function should be called from inside the master to make sure
// all the watchers and process handling are done correctly.
// Returns a tuple with the process and an error in case there's any.
func (preparable *BinaryPreparable) Start() (process.ProcContainer, error) {
	proc := &process.Proc{
		Name:       preparable.Name,
		Cmd:        preparable.Cmd,
		Args:       preparable.Args,
		Envs:       preparable.Envs,
		WorkingDir: preparable.getPath(),
		Path:       preparable.getPath(),
		Pidfile:    preparable.getPidPath(),
		Outfile:    preparable.getOutPath(),
		Errfile:    preparable.getErrPath(),
		KeepAlive:  preparable.KeepAlive,
		Status:     &process.ProcStatus{},
	}

	err := proc.Start()
	return proc, err
}

// Start will execute the process based on the information presented on the preparable.
// This function should be called from inside the master to make sure
// all the watchers and process handling are done correctly.
// Returns a tuple with the process and an error in case there's any.
func (preparable *BinaryPreparable) SetupProc() (process.ProcContainer, error) {
	proc := &process.Proc{
		Name:       preparable.Name,
		Cmd:        preparable.Cmd,
		Args:       preparable.Args,
		Envs:       preparable.Envs,
		WorkingDir: preparable.getPath(),
		Path:       preparable.getPath(),
		Pidfile:    preparable.getPidPath(),
		Outfile:    preparable.getOutPath(),
		Errfile:    preparable.getErrPath(),
		KeepAlive:  preparable.KeepAlive,
		Status:     &process.ProcStatus{},
		Pid:        -1,
	}

	return proc, nil
}

// Identifier is a function that get proc name
func (preparable *BinaryPreparable) Identifier() string {
	return preparable.Name
}

func (preparable *BinaryPreparable) getPath() string {
	if preparable.SysFolder[len(preparable.SysFolder)-1] == '/' {
		preparable.SysFolder = strings.TrimSuffix(preparable.SysFolder, "/")
	}
	return preparable.SysFolder + "/" + preparable.Name
}

func (preparable *BinaryPreparable) getBinPath() string {
	return preparable.SourcePath
}

func (preparable *BinaryPreparable) getPidPath() string {
	return preparable.getPath() + "/" + preparable.Name + ".pid"
}

func (preparable *BinaryPreparable) getOutPath() string {
	return preparable.getPath() + "/" + preparable.Name + ".out"
}

func (preparable *BinaryPreparable) getErrPath() string {
	return preparable.getPath() + "/" + preparable.Name + ".err"
}
