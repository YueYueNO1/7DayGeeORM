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
	Model      interface{}       //被映射的对象
	Name       string            //表名
	Fields     []*Field          //字段
	FieldNames []string          //包含所有字段名（列名）
	fieldMap   map[string]*Field //记录字段名和Field的映射关系，方便之后直接使用，无需遍历Field
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
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

// 用于从一个目标对象中提取字段值并返回一个包含这些字段值的interface{}切片
// dest interface{} 参数表示目标对象，可以是任意类型的指针，在函数内部使用了反射的机制来获取目标对象的值
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest)) //valueOf获取目标对象的反射值，Indirect方法获取目标对象的实际值，即去除指针的指向。
	var fieldValues []interface{}                        //用于存储字段值
	for _, field := range schema.Fields {                //遍历
		//通过fieldByName获取目标对象中对应字段的反射值，通过interface（）方法获取该字段的具体值，然后将该字段值添加到fieldValues
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues //包含目标中所有字段值的一个interface{}切片
}
