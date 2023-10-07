package schema

import (
	"geeorm/dialect"
	"github.com/rogpeppe/godef/go/ast"
	"reflect"
)

// 代表数据库的一栏数据
type Field struct {
	Name string //字段名
	Type string //类型
	Tag  string //约束条件
}

type Schema struct {
	Model     interface{}       //被映射的对象
	Name      string            //表名
	Fields    []*Field          //字段
	FieldName []string          //包含所有字段名（列名）
	fieldMap  map[string]*Field //记录字段名和Field的映射关系，方便之后直接使用，无需遍历Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// 将任意的对象解析为Schema实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}
	for i := 0; i < modelType.NumField(); i++ {

		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}

			schema.Fields = append(schema.Fields, field)
			schema.FieldName = append(schema.FieldName, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
