package main

import (
	"github.com/spf13/cobra"
	"encoding/json"
	"github.com/chaosblade-io/chaosblade/transport"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"github.com/chaosblade-io/chaosblade/util"
)

type StatusCommand struct {
	baseCommand
	exp         *expCommand
	commandType string
	target      string
	uid         string
}

func (sc *StatusCommand) Init() {
	sc.command = &cobra.Command{
		Use:     "status",
		Short:   "Get command or experiment status",
		Long:    "Get command or experiment status",
		Aliases: []string{"s"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return sc.runStatus(cmd, args)
		},
		Example: statusExample(),
	}
	sc.command.Flags().StringVar(&sc.commandType, "type", "", "command type, attach|create|destroy|detach")
	sc.command.Flags().StringVar(&sc.target, "target", "", "experiment target, for example: dubbo")
	sc.command.Flags().StringVar(&sc.uid, "uid", "", "prepare or experiment uid")

}
func (sc *StatusCommand) runStatus(command *cobra.Command, args []string) error {
	var uid = ""
	if len(args) > 0 {
		uid = args[0]
	} else {
		uid = sc.uid
	}
	var result interface{}
	var err error
	switch sc.commandType {
	case "create", "destroy", "c", "d":
		if uid != "" {
			result, err = db.QueryExperimentModelByUid(uid)
		} else if sc.target != "" {
			result, err = db.QueryExperimentModelsByCommand(sc.target)
		} else {
			result, err = db.ListExperimentModels()
		}
	case "prepare", "revoke", "p", "r":
		if uid != "" {
			result, err = db.QueryPreparationByUid(uid)
		} else {
			result, err = db.ListPreparationRecords()
		}
	default:
		if uid == "" {
			return transport.ReturnFail(transport.Code[transport.IllegalCommand], "must specify the right type or uid")
		}
		result, err = db.QueryExperimentModelByUid(uid)
		if result == nil || err != nil {
			result, err = db.QueryPreparationByUid(uid)
		}
	}
	if err != nil {
		return transport.ReturnFail(transport.Code[transport.DatabaseError], err.Error())
	}
	if util.IsNil(result) {
		return transport.Return(transport.Code[transport.DataNotFound])
	}
	response := transport.ReturnSuccess(result)

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		bytes, err := json.MarshalIndent(response, "", "\t")
		if err != nil {
			return response
		}
		sc.command.Println(string(bytes))
	} else {
		sc.command.Println(response.Print())
	}
	return nil
}

func statusExample() string {
	return `status cc015e9bd9c68406
status --type create`
}
