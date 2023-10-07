package session

import (
	"geeorm/clause"
	"reflect"
)

// 将已经存在的对象的每一个字段的值平铺开来
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames) //多次调用clause.Set构造好每个子句
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)              //构造子句
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES) //调用一次clause.Build按照传入的顺序构造出最终的SQL语句
	result, err := s.Raw(sql, vars...).Exec()                 //调用Raw.Exec方法执行
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// 期望调用方法是：传入一个切片指针，查询的结果保存到切片中
// 根据平铺开的字段的值构造出对象。！反射！
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable() //获取表数据

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)                              // 拼接SQL语句
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT) //构造最终语句
	rows, err := s.Raw(sql, vars...).QueryRows()                                           //根据传入的sql,vars在raw构造一个Session对象，来获取数据库表的数据
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem() //dest是指向结构体的指针，需要用reflect.New(destType).Elem()创建一个新的结构体实例
		var values []interface{}
		for _, name := range table.FieldNames {
			// dest.FieldByName(name).Addr().Interface() ：获取目标结构体 dest 中指定字段名 name 的指针，并将其转换为 interface{} 类型的值，以便在 rows.Scan() 中将查询结果赋值给对应字段。
			values = append(values, dest.FieldByName(name).Addr().Interface()) //获取结构体中所有字段的指针，然后把指针传递给Scan
		}
		if err := rows.Scan(values...); err != nil {
			return err
		} //通过rows.Scan方法将查询结果映射到一个结构体实例dest中
		destSlice.Set(reflect.Append(destSlice, dest)) //利用反射机制将dest添加到destSlice中
	}
	return rows.Close()
	//通过这种方式，可以将查询结果转换为目标类型的值，而无需手动编写扫描代码或添加每个字段的 setter 方法，极大地减少了代码复杂度。这也体现了反射机制在 ORM 框架中的重要性和实际应用场景。
}
