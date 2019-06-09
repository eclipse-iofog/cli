package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Delete an Agent",
		Long:    `Delete an Agent`,
		Example: `iofogctl delete agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, err := delete.NewExecutor(namespace, name)
			util.Check(err)

			// Run the command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
