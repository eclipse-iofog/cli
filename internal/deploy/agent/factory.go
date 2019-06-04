package deployagent

import (
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Namespace string
	Name      string
	User      string
	Host      string
	KeyFile   string
	Local     bool
	AgentName string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	_, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check Agent Name is specified
	if opt.AgentName == "" {
		return nil, util.NewInputError("Must specify agent name")
	}

	// Check controller already exists
	_, err = config.GetAgent(opt.Namespace, opt.Name)
	if err == nil {
		return nil, util.NewConflictError(opt.Namespace + "/" + opt.Name)
	}

	// Local executor
	if opt.Local == true {
		return newLocalExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
