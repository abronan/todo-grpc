package orm

import (
	"github.com/go-pg/pg/internal"
	"github.com/go-pg/pg/types"
)

type join struct {
	Parent     *join
	BaseModel  tableModel
	JoinModel  tableModel
	Rel        *Relation
	ApplyQuery func(*Query) (*Query, error)

	Columns []string
}

func (j *join) Select(db DB) error {
	switch j.Rel.Type {
	case HasManyRelation:
		return j.selectMany(db)
	case Many2ManyRelation:
		return j.selectM2M(db)
	}
	panic("not reached")
}

func (j *join) selectMany(db DB) error {
	q, err := j.manyQuery(db)
	if err != nil {
		return err
	}
	if q == nil {
		return nil
	}
	return q.Select()
}

func (j *join) manyQuery(db DB) (*Query, error) {
	manyModel := newManyModel(j)
	if manyModel == nil {
		return nil, nil
	}

	q := NewQuery(db, manyModel)
	if j.ApplyQuery != nil {
		var err error
		q, err = j.ApplyQuery(q)
		if err != nil {
			return nil, err
		}
	}

	q.columns = append(q.columns, hasManyColumnsAppender{j})

	baseTable := j.BaseModel.Table()
	var where []byte
	where = append(where, "("...)
	where = appendColumns(where, j.JoinModel.Table().Alias, j.Rel.FKs)
	where = append(where, ") IN ("...)
	where = appendChildValues(
		where, j.JoinModel.Root(), j.JoinModel.ParentIndex(), j.Rel.FKValues)
	where = append(where, ")"...)
	q = q.Where(internal.BytesToString(where))

	if j.Rel.Polymorphic != nil {
		q = q.Where(
			`? IN (?, ?)`,
			j.Rel.Polymorphic.Column,
			baseTable.ModelName, baseTable.TypeName,
		)
	}

	return q, nil
}

func (j *join) selectM2M(db DB) error {
	q, err := j.m2mQuery(db)
	if err != nil {
		return err
	}
	if q == nil {
		return nil
	}
	return q.Select()
}

func (j *join) m2mQuery(db DB) (*Query, error) {
	m2mModel := newM2MModel(j)
	if m2mModel == nil {
		return nil, nil
	}

	q := NewQuery(db, m2mModel)
	if j.ApplyQuery != nil {
		var err error
		q, err = j.ApplyQuery(q)
		if err != nil {
			return nil, err
		}
	}

	q.columns = append(q.columns, hasManyColumnsAppender{j})

	index := j.JoinModel.ParentIndex()
	baseTable := j.BaseModel.Table()
	var join []byte
	join = append(join, "JOIN "...)
	if db != nil {
		join = db.FormatQuery(join, string(j.Rel.M2MTableName))
	} else {
		join = append(join, j.Rel.M2MTableName...)
	}
	join = append(join, " AS "...)
	join = append(join, j.Rel.M2MTableAlias...)
	join = append(join, " ON ("...)
	for i, col := range j.Rel.BaseFKs {
		if i > 0 {
			join = append(join, ", "...)
		}
		join = append(join, j.Rel.M2MTableAlias...)
		join = append(join, '.')
		join = types.AppendField(join, col, 1)
	}
	join = append(join, ") IN ("...)
	join = appendChildValues(join, j.BaseModel.Root(), index, baseTable.PKs)
	join = append(join, ")"...)
	q = q.Join(internal.BytesToString(join))

	joinTable := j.JoinModel.Table()
	for i, col := range j.Rel.JoinFKs {
		if i >= len(joinTable.PKs) {
			break
		}
		pk := joinTable.PKs[i]
		q = q.Where(
			"?.? = ?.?",
			joinTable.Alias, pk.Column,
			j.Rel.M2MTableAlias, types.F(col),
		)
	}

	return q, nil
}

func (j *join) hasParent() bool {
	if j.Parent != nil {
		switch j.Parent.Rel.Type {
		case HasOneRelation, BelongsToRelation:
			return true
		}
	}
	return false
}

func (j *join) appendAlias(b []byte) []byte {
	b = append(b, '"')
	b = appendAlias(b, j, true)
	b = append(b, '"')
	return b
}

func (j *join) appendAliasColumn(b []byte, column string) []byte {
	b = append(b, '"')
	b = appendAlias(b, j, true)
	b = append(b, "__"...)
	b = types.AppendField(b, column, 2)
	b = append(b, '"')
	return b
}

func (j *join) appendBaseAlias(b []byte) []byte {
	if j.hasParent() {
		b = append(b, '"')
		b = appendAlias(b, j.Parent, true)
		b = append(b, '"')
		return b
	}
	return append(b, j.BaseModel.Table().Alias...)
}

func appendAlias(b []byte, j *join, topLevel bool) []byte {
	if j.hasParent() {
		b = appendAlias(b, j.Parent, topLevel)
		topLevel = false
	}
	if !topLevel {
		b = append(b, "__"...)
	}
	b = append(b, j.Rel.Field.SQLName...)
	return b
}

func (j *join) appendHasOneColumns(b []byte) []byte {
	if j.Columns == nil {
		for i, f := range j.JoinModel.Table().Fields {
			if i > 0 {
				b = append(b, ", "...)
			}
			b = j.appendAlias(b)
			b = append(b, '.')
			b = append(b, f.Column...)
			b = append(b, " AS "...)
			b = j.appendAliasColumn(b, f.SQLName)
		}
		return b
	}

	for i, column := range j.Columns {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = j.appendAlias(b)
		b = append(b, '.')
		b = types.AppendField(b, column, 1)
		b = append(b, " AS "...)
		b = j.appendAliasColumn(b, column)
	}

	return b
}

func (j *join) appendHasOneJoin(db DB, b []byte) []byte {
	b = append(b, "LEFT JOIN "...)
	if db != nil {
		b = db.FormatQuery(b, string(j.JoinModel.Table().Name))
	} else {
		b = append(b, j.JoinModel.Table().Name...)
	}
	b = append(b, " AS "...)
	b = j.appendAlias(b)

	b = append(b, " ON "...)
	if j.Rel.Type == HasOneRelation {
		joinTable := j.Rel.JoinTable
		for i, fk := range j.Rel.FKs {
			if i > 0 {
				b = append(b, " AND "...)
			}
			b = j.appendAlias(b)
			b = append(b, '.')
			b = append(b, joinTable.PKs[i].Column...)
			b = append(b, " = "...)
			b = j.appendBaseAlias(b)
			b = append(b, '.')
			b = append(b, fk.Column...)
		}
	} else {
		baseTable := j.BaseModel.Table()
		for i, fk := range j.Rel.FKs {
			if i > 0 {
				b = append(b, " AND "...)
			}
			b = j.appendAlias(b)
			b = append(b, '.')
			b = append(b, fk.Column...)
			b = append(b, " = "...)
			b = j.appendBaseAlias(b)
			b = append(b, '.')
			b = append(b, baseTable.PKs[i].Column...)
		}
	}

	return b
}

type hasManyColumnsAppender struct {
	*join
}

func (q hasManyColumnsAppender) AppendFormat(b []byte, f QueryFormatter) []byte {
	if q.Rel.M2MTableAlias != "" {
		b = append(b, q.Rel.M2MTableAlias...)
		b = append(b, ".*, "...)
	}

	joinTable := q.JoinModel.Table()

	if q.Columns != nil {
		for i, column := range q.Columns {
			if i > 0 {
				b = append(b, ", "...)
			}
			b = append(b, joinTable.Alias...)
			b = append(b, '.')
			b = types.AppendField(b, column, 1)
		}
		return b
	}

	return appendColumns(b, joinTable.Alias, joinTable.Fields)
}
