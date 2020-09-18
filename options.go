package ggs

//ExitCallback is a function which would be invoked when service shutdown
type ExitCallback func()

// InitOptions struct having information about init parameters
type InitOptions struct {
	configDir string
}

// InitOption used by the Init
type InitOption func(*InitOptions)

// WithConfigDir is option to set config dir.
func WithConfigDir(dir string) InitOption {
	return func(o *InitOptions) {
		o.configDir = dir
	}
}

// getInitOpts is to get the init options
func getInitOpts(options ...InitOption) InitOptions {
	opts := InitOptions{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}

// RunOptions struct having information about run parameters
type RunOptions struct {
	exitCb ExitCallback
}

// RunOption used by the Run
type RunOption func(*RunOptions)

// WithExitCallback is option to set exit callback for graceful shutdown.
func WithExitCallback(cb ExitCallback) RunOption {
	return func(o *RunOptions) {
		o.exitCb = cb
	}
}

// getRunOpts is to get the run options
func getRunOpts(options ...RunOption) RunOptions {
	opts := RunOptions{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}
