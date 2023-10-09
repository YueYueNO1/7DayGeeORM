package session

import (
	"database/sql"
	"geeorm/clause"
	"geeorm/dialect"
	"geeorm/log"
	"geeorm/schema"
	"strings"
)

// 用于实现数据库交互
type Session struct {
	db      *sql.DB         //sql.Open()方法连接数据库成功之后返回的指针
	sql     strings.Builder //23变量用来拼接SQL语句和sql语句中占位符的对应位置
	sqlVars []interface{}   //用户调用Raw()方法即可改变这两个变量的值

	dialect  dialect.Dialect
	refTable *schema.Schema

	clause clause.Clause
	//新增对事务的支持
	tx *sql.Tx
}

// 用于描述数据库操作的最小功能集合
type CommonDB interface {
	//用于执行查询语句，并返回查询结果的行集合和一个可能的错误
	Query(query string, args ...interface{}) (*sql.Rows, error)
	//用于执行查询语句，并返回查询结果的单行数据
	QueryRow(query string, args ...interface{}) *sql.Row
	//用于执行非查询语句，返回关于执行结果的一些信息，如受影响的行数等
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// 实现对sql.DB 和 sql.Tx 的显式声明，用于确保它们都实现了CommonDB接口。
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

func (s *Session) DB() CommonDB {
	//新增！=nil判断，用于事务
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// 封装三个函数，使用log统一打印日志，而且每次操作执行之后清空sql的两个变量，这样Session可以复用
// 开启一次会话可以执行多次SQL
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
