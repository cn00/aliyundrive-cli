package internal

import (
	"context"
	"database/sql"
	_ "database/sql"
	"fmt"
	"github.com/chyroc/go-aliyundrive"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type CommandUniq struct {
	cli *Cli
	db  *sql.DB
}

type ItemView struct {
	id                   int64
	name, file_ids, hash string
	c, size              int64
}

func (r *CommandUniq) uniq() {
	sqls := `select *
		from (select min(id) id, name, group_concat(file_id, ',') file_ids, count(id) c, sum(size * 1.0) / 1024 / 1024 size, hash
			  from item
			  where hash <> ':'
			  group by hash)
		where c > 1`
	query, err := r.db.Query(sqls)

	if err != nil {
		println(err.Error())
	}
	var rows []ItemView
	for query.Next() {
		row := ItemView{}
		query.Scan(&row.id, &row.name, &row.file_ids, &row.c, &row.size, &row.hash)
		fmt.Printf("%+v\n\n", row)
		rows = append(rows, row)
	}
	query.Close()

	stmt, err := r.db.Prepare(`update item set flag = 0 where file_id = ?`)
	for _, row := range rows {
		file_ids := strings.Split(row.file_ids, ",")
		for i := 1; i < len(file_ids); i++ {
			_, err = r.cli.ali.File.DeleteFile(context.Background(), &aliyundrive.DeleteFileReq{
				DriveID: r.cli.driveID,
				FileID:  file_ids[i],
			})
			if err == nil {
				_, err = stmt.Exec(file_ids[i])
				println("delete " + row.name + file_ids[i])
				if err != nil {
					println("upate flag err:" + err.Error())
				}
			}
		}
	}
}

func (r *CommandUniq) Run() (err error) {
	if err := r.cli.setupDrive(); err != nil {
		return err
	}

	if err := r.cli.setupFiles(); err != nil {
		return err
	}

	r.db, err = sql.Open("sqlite3", "/Users/cn/ws/aliyundrive-cli/aliyundrive.db")
	checkErr(err)

	r.uniq()

	defer r.db.Close()
	return nil
}
