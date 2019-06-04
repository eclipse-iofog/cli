package deploycontroller

import (
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Name       string
	Namespace  string
	User       string
	Host       string
	KeyFile    string
	Local      bool
	KubeConfig string
	Password   string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	_, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check controller already exists
	_, err = config.GetController(opt.Namespace, opt.Name)
	if err == nil {
		return nil, util.NewConflictError(opt.Namespace + "/" + opt.Name)
	}

	// Check password provided
	if opt.Password == "" {
		return nil, util.NewInputError("Must specify a password")
	}

	// Local executor
	if opt.Local == true {
		return newLocalExecutor(opt), nil
	}

	// Kubernetes executor
	if opt.KubeConfig != "" {
		return newKubernetesExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
