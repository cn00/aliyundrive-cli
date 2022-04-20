package internal

import (
	"database/sql"
	_ "database/sql"
	"github.com/chyroc/go-aliyundrive"
	_ "github.com/mattn/go-sqlite3"
)

type CommandDump struct {
	cli *Cli
	db  *sql.DB
}

func checkErr(err error) {
	if err != nil {
		println(err.Error())
	}
}

func (r *CommandDump) doOnDir(v *aliyundrive.File, pid int64, stmt *sql.Stmt) (err error) {
	r.cli.checkoutDir(v.FileID, v.Name)
	if err := r.cli.setupFiles(); err != nil {
		return err
	}
	println(v.Name, len(r.cli.files))
	for _, v := range r.cli.files {
		itype := 0
		if v.Type == "file" {
			itype = 1
		}
		//fmt.Printf("%s:%+v\n\n",v.Name, v)
		res, err := stmt.Exec(v.Name, itype, v.Size, v.URL, pid, v.ContentHashName+":"+v.ContentHash, v.FileExtension, v.FileID)
		checkErr(err)
		if err != nil {
			continue
		}
		if v.Type == "folder" {
			pid, err := res.LastInsertId()
			r.doOnDir(v, pid, stmt)
			checkErr(err)
		}
	}
	return
}

func (r *CommandDump) Run() (err error) {
	if err := r.cli.setupDrive(); err != nil {
		return err
	}

	if err := r.cli.setupFiles(); err != nil {
		return err
	}

	r.db, err = sql.Open("sqlite3", "/Users/cn/ws/aliyundrive-cli/aliyundrive.db")
	checkErr(err)

	// init table
	r.db.Exec(`create table if not exists item 
		(
			id   integer not null
				constraint item_pk primary key autoincrement,
			p_id integer,
			type int,
			name text,
			size int,
			url  text,
			hash text,
			constraint item_pk_2 unique (p_id, name)
		);
		
		create unique index item_id_uindex on item (id);
		`)

	//r.cli.PrintFiles(r.cli.files)
	stmt, err := r.db.Prepare("INSERT INTO item(name, type, size, url, p_id, hash, surfix, file_id) values(?,?,?,?,?,?,?,?)")
	checkErr(err)
	for _, v := range r.cli.files {
		r.doOnDir(v, 0, stmt)
	}
	defer r.db.Close()
	return nil
}
