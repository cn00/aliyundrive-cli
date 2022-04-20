package internal

import (
	"context"
	"database/sql"
	_ "database/sql"
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
		println("db.Query err:", err.Error())
	}
	var rows []ItemView
	for query.Next() {
		row := ItemView{}
		query.Scan(&row.id, &row.name, &row.file_ids, &row.c, &row.size, &row.hash)
		//fmt.Printf("query.Scan: %+v\n\n", row)
		rows = append(rows, row)
	}
	query.Close()

	stmt, err := r.db.Prepare(`update item set flag = 0 where file_id in (?)`)
	var batch []string
	for _, row := range rows {
		file_ids := strings.Split(row.file_ids, ",")
		batch = append(batch, file_ids[1:]...)
		if len(batch) < 200 {
			continue
		}
		r.cli.ali.File.Batch(context.Background(), aliyundrive.BatchDelete, r.cli.driveID, batch)
		batchs := strings.Join(batch, ",")
		_, err = stmt.Exec(batchs)
		//for i := 1; i < len(batch); i++ {
		//	_, err = stmt.Exec(batch[i])
		println("delete " + batchs + " == " + row.file_ids)
		if err != nil {
			println("upate flag err:" + err.Error())
		}
		//}
		batch = make([]string, 1)
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
