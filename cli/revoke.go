package main

import (
	"fmt"
	"context"
	"github.com/spf13/cobra"
	"github.com/chaosblade-io/chaosblade/transport"
	"github.com/chaosblade-io/chaosblade/exec"
	"github.com/chaosblade-io/chaosblade/exec/jvm"
)

type RevokeCommand struct {
	baseCommand
}

func (rc *RevokeCommand) Init() {
	rc.command = &cobra.Command{
		Use:     "revoke [PREPARE UID]",
		Aliases: []string{"r"},
		Short:   "Revoke deploy",
		Long:    "Revoke deploy",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runRevoke(args)
		},
		Example: revokeExample(),
	}
}

func (rc *RevokeCommand) runRevoke(args []string) error {
	uid := args[0]
	record, err := db.QueryPreparationByUid(uid)
	if err != nil {
		return transport.ReturnFail(transport.Code[transport.DatabaseError],
			fmt.Sprintf("query record err, %s", err.Error()))
	}
	if record == nil {
		return transport.ReturnFail(transport.Code[transport.DataNotFound],
			fmt.Sprintf("the uid record not found"))
	}
	var response *transport.Response
	var channel = exec.NewLocalChannel()
	switch record.ProgramType {
	case PrepareJvmType:
		response = jvm.Detach(record.Port)
	case PrepareK8sType:
		args := fmt.Sprintf("delete ns chaosblade")
		response = channel.Run(context.Background(), "kubectl", args)
	default:
		return transport.ReturnFail(transport.Code[transport.IllegalParameters],
			fmt.Sprintf("not support the %s type", record.ProgramType))
	}
	if !response.Success {
		checkError(db.UpdatePreparationRecordByUid(uid, "Running", fmt.Sprintf("revoke failed. %s", response.Err)))
		return response
	}
	checkError(db.UpdatePreparationRecordByUid(uid, "Revoked", ""))
	rc.command.Println(response.Print())
	return nil
}

func revokeExample() string {
	return `blade revoke <UID>`
}
