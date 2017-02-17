package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/mattn/go-adodb"
	"jiang/logout"
	"strconv"
	"strings"
)

type Db struct {
	Name   string
	Source string
	Log    *logout.Logout
}

func (d *Db) GetDbBy(name, source string) *sql.DB {
	d.Name = name
	d.Source = source
	return d.GetDb()
}

//连接数据库
func (d *Db) GetDb() *sql.DB {
	db, err := sql.Open(d.Name, d.Source)
	if err != nil {
		d.Log.Out("sql.Open is err", err, d.Source)
		return nil
	}
	d.Log.Out("sql.Open is ok", d.Source)
	return db
}

//批处理，key为主键，空时为id
func (d *Db) BatchStatement(table, key, data string, col []string) string {
	inset := fmt.Sprintf("`%s`", strings.Join(col, "`, `"))
	if key == "" {
		key = "id"
	}
	update := ""
	for _, v := range col {
		if v == key {
			continue
		}
		update += fmt.Sprintf("%s=VALUES(%s)", v, v)
	}

	if update != "" {
		update = "ON DUPLICATE KEY UPDATE " + update
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES %s %s", table, inset, data, update)
}

func (d *Db) InsertUpdate(table string, re map[string]interface{}, col []string) string {
	var (
		key    = make([]string, 0)
		value  = make([]string, 0)
		update = make([]string, 0)
	)
	for k, val := range re {
		b := true
		v := ToString(val)
		key = append(key, k)
		value = append(value, v)
		for _, c := range col {
			if k == c {
				b = false
				break
			}
		}
		if b {
			s := fmt.Sprintf("%s=VALUES(%s)", k, k)
			update = append(update, s)
		}
	}
	return fmt.Sprintf("INSERT INTO %s (`%s`) VALUES ('%s') ON DUPLICATE KEY UPDATE %s", table, strings.Join(key, "`, `"), strings.Join(value, "', '"), strings.Join(update, ","))
}

//回滚备份SQL语句
func (d *Db) RollStatement(table string, col []string, id int, roll bool) string {
	set := make([]string, 0)
	for _, v := range col {
		a := "a."
		if !roll {
			a = ""
		}
		s := fmt.Sprintf("%s`%s` = b.`%s`", a, v, v)
		set = append(set, s)
	}
	column := strings.Join(col, "`,`")

	if !roll {
		return fmt.Sprintf("INSERT INTO %s (`edit`, `%s`) SELECT `id`, `%s` FROM %s b WHERE id = '%v' ON DUPLICATE KEY UPDATE %s", table, column, column, table, id, strings.Join(set, ", "))
	}
	return fmt.Sprintf("UPDATE %s a INNER JOIN %s b ON a.id = b.edit SET %s WHERE a.id = '%v'", table, table, strings.Join(set, ", "), id)
}

//执行语句
func (d *Db) ExecStatement(table string, re map[string]interface{}) string {

	var (
		key   string
		value string
		set   = make([]string, 0)
		i     = 0
	)

	//SQL更新的语句
	id := ToString(re["id"])
	for k, v := range re {
		//update
		if k != "id" {
			s := fmt.Sprintf("`%s` = '%s'", k, ToString(v))
			set = append(set, s)
			i++
		}
		//insert
		if key == "" {
			key = fmt.Sprintf("`%s`", k)
			value = fmt.Sprintf("'%s'", ToString(v))
			continue
		}
		key = fmt.Sprintf("%s, `%s`", key, k)
		value = fmt.Sprintf("%s,'%s'", value, ToString(v))
	}

	if id != "" && id != "0" {
		return fmt.Sprintf("UPDATE %s SET %s WHERE id = '%s'", table, strings.Join(set, ", "), id)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUE (%s) ", table, key, value)
}

func (d *Db) Feach(db *sql.DB, statement string) []interface{} {
	rows, err := db.Query(statement)
	if err != nil {
		d.Log.Out("db.Query is err", err, statement)
		return nil
	}
	d.Log.Out("db.Query is ok", statement)
	defer rows.Close()

	columns, err := rows.Columns()
	result := make([]interface{}, 0)
	for rows.Next() {
		//初始Scan接收值
		d := make(map[string]interface{})
		re := make([]sql.RawBytes, len(columns))
		args := make([]interface{}, len(columns))
		for key := range columns {
			//使用&re[key],不能直接用 key, val := range re{} 中的&val
			args[key] = &re[key]
		}

		rows.Scan(args...)
		//由于args用的是rer的指针所以直接使用re
		for k, v := range re {
			s := string(v)
			i, err := strconv.Atoi(s)
			if err == nil {
				d[columns[k]] = i
			} else {
				d[columns[k]] = s
			}
		}
		result = append(result, d)
	}
	return result
}

func (d *Db) Exec(db *sql.DB, statement string) string {
	re, err := db.Exec(statement)
	if err != nil {
		d.Log.Out("db.Exec is err", err, statement)
		return "0"
	}
	d.Log.Out("db.Exec is ok", statement)

	i, err := re.LastInsertId()
	if err != nil {
		return "0"
	}

	if i == 0 {
		i = 1
	}
	return fmt.Sprintf("%v", i)
}

//转各类型为字符串
func ToString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return strconv.Itoa(v.(int))
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	default:
		return ""
	}
}
