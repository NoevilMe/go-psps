package psps

// Process is the generic interface that is implemented on every platform
// and provides common operations for processes.
type Process interface {
	// Pid is the process ID for this process.
	Pid() int

	// PPid is the parent process ID for this process.
	PPid() int

	// PGid is the parent process group lID for this process.
	PGid() int

	// Executable name running this process. This is not a path to the executable.
	Name() string

	// Executable path running this process.
	ImagePath() string

	// Current work directory
	Cwd() string

	// command line arguments
	CmdLine() []string
}