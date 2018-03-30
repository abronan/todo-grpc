package orm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-pg/pg/internal"
)

func Update(db DB, model ...interface{}) error {
	res, err := NewQuery(db, model...).Update()
	if err != nil {
		return err
	}
	return internal.AssertOneRow(res.RowsAffected())
}

type updateQuery struct {
	q *Query

	omitZero bool
}

var _ QueryAppender = (*updateQuery)(nil)

func (q updateQuery) Copy() QueryAppender {
	return updateQuery{
		q: q.q.Copy(),
	}
}

func (q updateQuery) Query() *Query {
	return q.q
}

func (q updateQuery) AppendQuery(b []byte) ([]byte, error) {
	if q.q.stickyErr != nil {
		return nil, q.q.stickyErr
	}

	var err error

	if len(q.q.with) > 0 {
		b, err = q.q.appendWith(b)
		if err != nil {
			return nil, err
		}
	}

	b = append(b, "UPDATE "...)
	b = q.q.appendFirstTableWithAlias(b)

	b, err = q.mustAppendSet(b)
	if err != nil {
		return nil, err
	}

	if q.q.hasOtherTables() || q.modelHasData() {
		b = append(b, " FROM "...)
		b = q.q.appendOtherTables(b)
		b, err = q.appendModelData(b)
		if err != nil {
			return nil, err
		}
	}

	b, err = q.q.mustAppendWhere(b)
	if err != nil {
		return nil, err
	}

	if len(q.q.returning) > 0 {
		b = q.q.appendReturning(b)
	}

	return b, nil
}

func (q *updateQuery) modelHasData() bool {
	if !q.q.hasModel() {
		return false
	}
	v := q.q.model.Value()
	return v.Kind() == reflect.Slice && v.Len() > 0
}

func (q *updateQuery) mustAppendSet(b []byte) ([]byte, error) {
	if len(q.q.set) > 0 {
		b = q.q.appendSet(b)
		return b, nil
	}

	if q.q.model == nil {
		return nil, errors.New("pg: Model is nil")
	}

	b = append(b, " SET "...)

	value := q.q.model.Value()
	var err error
	if value.Kind() == reflect.Struct {
		b, err = q.appendSetStruct(b, value)
	} else {
		if q.modelHasData() {
			b, err = q.appendSetSlice(b, value)
		} else {
			err = fmt.Errorf("pg: can't bulk-update empty slice %s", value.Type())
		}
	}
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (q updateQuery) appendSetStruct(b []byte, strct reflect.Value) ([]byte, error) {
	fields, err := q.q.getFields()
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		fields = q.q.model.Table().DataFields
	}

	pos := len(b)
	for _, f := range fields {
		if q.omitZero && f.OmitZero(strct) {
			continue
		}

		if len(b) != pos {
			b = append(b, ", "...)
			pos = len(b)
		}

		b = append(b, f.Column...)
		b = append(b, " = "...)

		app, ok := q.q.values[f.SQLName]
		if ok {
			b = app.AppendFormat(b, q.q)
		} else {
			b = f.AppendValue(b, strct, 1)
		}
	}
	return b, nil
}

func (q updateQuery) appendSetSlice(b []byte, slice reflect.Value) ([]byte, error) {
	fields, err := q.q.getFields()
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		fields = q.q.model.Table().DataFields
	}

	for i, f := range fields {
		if i > 0 {
			b = append(b, ", "...)
		}

		b = append(b, f.Column...)
		b = append(b, " = "...)
		b = append(b, "_data."...)
		b = append(b, f.Column...)
	}
	return b, nil
}

func (q updateQuery) appendModelData(b []byte) ([]byte, error) {
	if !q.q.hasModel() {
		return b, nil
	}

	v := q.q.model.Value()
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return b, nil
	}

	columns, err := q.q.getDataFields()
	if err != nil {
		return nil, err
	}

	if len(columns) > 0 {
		columns = append(columns, q.q.model.Table().PKs...)
	} else {
		columns = q.q.model.Table().Fields
	}

	return appendSliceValues(b, columns, v), nil
}

func appendSliceValues(b []byte, fields []*Field, slice reflect.Value) []byte {
	b = append(b, "(VALUES ("...)
	for i := 0; i < slice.Len(); i++ {
		el := slice.Index(i)
		if el.Kind() == reflect.Interface {
			el = el.Elem()
		}
		b = appendValues(b, fields, reflect.Indirect(el))
		if i != slice.Len()-1 {
			b = append(b, "), ("...)
		}
	}
	b = append(b, ")) AS _data("...)
	b = appendColumns(b, "", fields)
	b = append(b, ")"...)
	return b
}

func appendValues(b []byte, fields []*Field, v reflect.Value) []byte {
	for i, f := range fields {
		if i > 0 {
			b = append(b, ", "...)
		}
		if f.OmitZero(v) {
			b = append(b, "NULL"...)
		} else {
			b = f.AppendValue(b, v, 1)
		}
		if f.HasFlag(customTypeFlag) {
			b = append(b, "::"...)
			b = append(b, f.SQLType...)
		}
	}
	return b
}
