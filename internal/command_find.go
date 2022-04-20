package internal

import (
	"context"
	"fmt"
	"github.com/chyroc/go-aliyundrive"
)

type CommandFind struct {
	cli    *Cli
	fileID string
}

func (r *CommandFind) Run() (err error) {
	if err := r.cli.setupDrive(); err != nil {
		return err
	}

	res, err := r.cli.ali.File.GetFile(context.Background(), &aliyundrive.GetFileReq{
		DriveID: r.cli.driveID,
		FileID:  r.fileID,
	})

	fmt.Printf("%+v\n", res)

	return
}
