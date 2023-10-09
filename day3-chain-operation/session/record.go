package session

import (
	"errors"
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

// 用于处理传递给Update方法的可变参数kv
// 目的是为了兼容不同的调用方式，既可以接收一个显式传递的map，也可以接受一组键值对作为参数，
// 并将它们转换为同一个的map格式
func (s *Session) Update(kv ...interface{}) (int64, error) {
	//首先通过强制转换，如果转换成功则直接赋值给变量m
	//如果转换失败，则表示第一个参数不是map类型，需要通过遍历参数切片kv构建一个新的map
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		//以偶数索引为键，奇数索引作为对应键的值
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()

}

func (s *Session) Delete() (int64, error) {
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	//scan方法：将查询结果存储到指定变量的方法；处理多行查询结果；处理不同数据类型的查询结果；错误处理
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem() //为了将查询结果存储到切片中并返回
	//限制最多只返回一个指向切片的指针作为Find方法的参数
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("Not Found")
	}
	dest.Set(destSlice.Index(0))
	return nil
	//实现原理：根据传入的类型，利用反射构造切片，调用：Limit（1）限制返回的行数，调用Find方法获取到查询结果
}
