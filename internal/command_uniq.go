package internal

import (
	"context"
	"database/sql"
	_ "database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/chyroc/go-aliyundrive"
	"github.com/mattn/go-sqlite3"
	sqlite "github.com/mattn/go-sqlite3"
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
		from (select min(id) id, name, group_concat(fid, ',') file_ids, count(id) c, sum(size * 1.0) / 1024 / 1024 size, hash
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


	stmt, err := r.db.Prepare(`update item set flag = 0 where fid in (?)`)
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

	sql.Register("sqlite3_c", &sqlite.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("get_full_path",
				func(id int) (path string) {
					rows, err :=conn.Query(fmt.Sprintf(get_full_path_sql, id), []driver.Value{"id"})
					if err != nil {println("get_full_path err:" + err.Error())}
					var paths = []driver.Value{""}
					//columns := rows.Columns()
					//println("columns:", columns)
					err = rows.Next(paths)
					if err != nil {println("get_full_path 2 err:" + err.Error())}
					path = paths[0].(string)
					if id % 1000 == 0 {println(id, path)}
					return
				},
				true); err != nil {
				return err
			}
			return nil
		},
	})

	a,b,c := sqlite3.Version()
	println("sqlit3.version: %+v %+v %+v", a,b,c)
	r.db, err = sql.Open("sqlite3_c", "/Users/cn/ws/aliyundrive-cli/aliyundrive.db")
	checkErr(err)

	//r.uniq()
	//res,err :=r.db.Query("select * from (select name from item limit 1) UNION all select get_full_path(?)", 1217)
	res,err :=r.db.Exec("update item set (fpath) = get_full_path(id)")
	if res == nil || err != nil {
		println("exec get_full_path err: " + err.Error())
	}

	defer r.db.Close()
	return nil
}

// 只能在终端里完全执行，custom_function 无法创建临时表，无法删除记录
var get_full_path_sql = `
begin transaction;
-- drop table if exists Path;
-- delete from Path where 1;

PRAGMA recursive_triggers = TRUE;
CREATE TABLE Path1 (id INTEGER, pid INTEGER,
                        name text);
commit;

begin transaction;
CREATE TRIGGER if not exists find_path AFTER INSERT ON Path2 BEGIN
    INSERT INTO Path SELECT id, pid, name FROM Item WHERE
            Item.id = new.pid;
END;

-- The flaw here is that label must be unique, so when creating
-- the table there must be a unique reference for selection
-- This insert sets off the trigger find_path

INSERT INTO Path SELECT id, pid, name FROM Item WHERE id = %d;

-- Return the hierarchy in order from "root" to "c2"
select group_concat(name, '/') fpath from(SELECT *, 1 j FROM Path ORDER BY id ASC ) group by j  ;

drop table Path

commit;

`
