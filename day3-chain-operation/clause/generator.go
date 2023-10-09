package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

// 实现各个字句的生成规则
func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

// 用于生成一组问号字符（“？”）
// 用于表示SQL语句中需要绑定的参数部分（num int代表需要生成的问号个数）
func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, " ?")

	}
	return strings.Join(vars, ", ")
}

// 用于SQL操作中INSERT语句的字符串和绑定参数值
// 实现了INSERT类型的SQL语句生成规则
// 第一个参数values 能够接收任意数量、任意类型的参数。
func _insert(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	//将字段名使用“，”连接串连成一个字符串然后使用fmt.Sprintf函数将“INSERT INTO 表名 （字段名）”字符串格式化成一条完整的SQL语句，并返回该SQL语句的字符串值
	fields := strings.Join(values[1].([]string), ",")
	//返回值： interface切片是用于存储SQL语句中绑定的参数值的，interface{}是用来传递位置的数据类型，[]interface{}主要用于处理多个不同类型的数据
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{} //表示一个空的interface{}切片
}

// 用于SQL操作中VALUES语句的字符串和绑定参数值

func _values(values ...interface{}) (string, []interface{}) {

	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" { //判断当前是否已经生成了绑定参数部分
			bindStr = genBindVars(len(v)) //如果没有就调用函数生成该部分，然后赋值给bindStr
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) { //判断是否为最后一条数据
			sql.WriteString(", ") //格式化成一个values元组
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars //绑定参数部分对应的参数切片为vars
}

func _select(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _limit(values ...interface{}) (string, []interface{}) {
	return "LIMIT ?", values
}

func _where(values ...interface{}) (string, []interface{}) {
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}

}

func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range m {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ",")), vars
}

func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"counts(*)"})
}
