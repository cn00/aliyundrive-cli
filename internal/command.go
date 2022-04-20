package internal

import (
	"errors"
	"fmt"
	"strings"
)

type Command interface {
	Run() error
}

const CommandUsage = `
Only support the follow sub command:
	2tv      	 send video to tv
	cd       	 chdir
	download 	 download file
	dump     	 dump all file info to sqlite
	exit     	 exit program
	find     	 find by file id
	help or ?	 print help usage
	ls       	 list files
	mkdir    	 create directory
	mv       	 file or directory
	rename   	 rename file or directory
	rm       	 remove file or directory
	uniq     	 uniq files
	upload   	 upload file
`

func (r *Cli) ParseCommand(input string) (Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	if input == "ls" || strings.HasPrefix(input, "ls ") {
		return &CommandLs{cli: r}, nil
	}
	if input == "exit" {
		Exit(0)
	}
	if strings.HasPrefix(input, "cd ") {
		return &CommandCd{cli: r, dir: strings.TrimSpace(input[len("cd "):])}, nil
	}
	if strings.HasPrefix(input, "dump") {
		return &CommandDump{cli: r}, nil
	}
	if strings.HasPrefix(input, "uniq") {
		return &CommandUniq{cli: r}, nil
	}
	if strings.HasPrefix(input, "find") {
		return &CommandFind{cli: r, fileID: strings.TrimSpace(input[len("find "):])}, nil
	}
	if strings.HasPrefix(input, "mkdir ") {
		return &CommandMkdir{cli: r, dir: strings.TrimSpace(input[len("mkdir "):])}, nil
	}
	if strings.HasPrefix(input, "rm ") {
		return &CommandRm{cli: r, name: strings.TrimSpace(input[len("rm "):])}, nil
	}
	if strings.HasPrefix(input, "2tv ") {
		return &CommandToTv{cli: r, name: strings.TrimSpace(input[len("2tv "):])}, nil
	}
	if strings.HasPrefix(input, "upload ") {
		return &CommandUpload{cli: r, file: strings.TrimSpace(input[len("upload "):])}, nil
	}
	if strings.HasPrefix(input, "download ") {
		return &CommandDownload{cli: r, name: strings.TrimSpace(input[len("download "):])}, nil
	}
	if strings.HasPrefix(input, "mv ") {
		l := splitSpace(strings.TrimSpace(input[len("mv "):]))
		if len(l) != 2 {
			return nil, fmt.Errorf("mv 命令不合法，需要两个以空格区分的参数，如: mv old new")
		}
		return &CommandMv{cli: r, from: l[0], to: l[1]}, nil
	}
	if strings.HasPrefix(input, "rename ") {
		l := splitSpace(strings.TrimSpace(input[len("rename "):]))
		if len(l) != 2 {
			return nil, fmt.Errorf("rename 命令不合法，需要两个以空格区分的参数，如: rename old new")
		}
		return &CommandRename{cli: r, from: l[0], to: l[1]}, nil
	}
	if strings.HasPrefix(input, "help") || strings.HasPrefix(input, "?") {
		return nil, errors.New(CommandUsage)
	}
	return nil, errors.New(CommandUsage)
}
