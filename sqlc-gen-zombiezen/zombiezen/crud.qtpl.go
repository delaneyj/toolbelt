// Code generated by qtc from "crud.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:1
package zombiezen

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:1
func StreamGenerateCRUD(qw422016 *qt422016.Writer, t *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:1
	qw422016.N().S(`
// Code generated by "sqlc-gen-zombiezen". DO NOT EDIT.

package `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:4
	qw422016.E().S(t.PackageName.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:4
	qw422016.N().S(`

import (
    "fmt"
    "zombiezen.com/go/sqlite"
)

type `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:11
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:11
	qw422016.N().S(`Model struct {
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:12
	for _, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:12
		qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.E().S(f.GoType.Original)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S(`json:"`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S(`"`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:13
		qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:14
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:14
	qw422016.N().S(`}


type Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:18
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:18
	qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.N().S(`(tx *sqlite.Conn) *Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.N().S(`Stmt {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:22
	qw422016.N().S(`
INSERT INTO `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:24
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:24
	qw422016.N().S(` (
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:25
	for i, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:25
		qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:26
		qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:26
		if i < len(t.Fields)-1 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:26
			qw422016.N().S(`,`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:26
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:26
		qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:27
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:27
	qw422016.N().S(`) VALUES (
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:29
	for i := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:29
		qw422016.N().S(`        ?`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:30
		if i < len(t.Fields)-1 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:30
			qw422016.N().S(`,`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:30
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:30
		qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:31
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:31
	qw422016.N().S(`)
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:31
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:31
	qw422016.N().S(`)
    return &Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:34
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:34
	qw422016.N().S(`Stmt{stmt: stmt}
}

func (ps *Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:37
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:37
	qw422016.N().S(`Stmt) Run(m *`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:37
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:37
	qw422016.N().S(`Model) error {
    defer ps.stmt.Reset()

    // Bind parameters
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:41
	streambindFields(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:41
	qw422016.N().S(`

    if _, err := ps.stmt.Step(); err != nil {
        return fmt.Errorf("failed to insert `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:44
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:44
	qw422016.N().S(`: %w", err)
    }

    return nil
}

func OnceCreate`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:50
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:50
	qw422016.N().S(`(tx *sqlite.Conn, m *`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:50
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:50
	qw422016.N().S(`Model) error {
    ps := Create`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:51
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:51
	qw422016.N().S(`(tx)
    return ps.Run(m)
}

type ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:55
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:55
	qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.N().S(`(tx *sqlite.Conn) *ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.N().S(`Stmt {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:59
	qw422016.N().S(`
SELECT
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:62
	for i, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:62
		qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:63
		qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:63
		if i < len(t.Fields)-1 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:63
			qw422016.N().S(`,`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:63
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:63
		qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:64
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:64
	qw422016.N().S(`FROM `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:65
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:65
	qw422016.N().S(`
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:65
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:65
	qw422016.N().S(`)
    return &ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:67
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:67
	qw422016.N().S(`Stmt{stmt: stmt}
}

func (ps *ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:70
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:70
	qw422016.N().S(`Stmt) Run() ([]*`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:70
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:70
	qw422016.N().S(`Model, error) {
    defer ps.stmt.Reset()

    var models []*`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:73
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:73
	qw422016.N().S(`Model
    for {
        hasRow, err := ps.stmt.Step()
        if err != nil {
            return nil, fmt.Errorf("failed to read `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:77
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:77
	qw422016.N().S(`: %w", err)
        } else if !hasRow {
            break
        }

        m := &`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:82
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:82
	qw422016.N().S(`Model{}
        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:83
	streamfillResStruct(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:83
	qw422016.N().S(`

        models = append(models, m)
    }

    return models, nil
}

func OnceReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:91
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:91
	qw422016.N().S(`(tx *sqlite.Conn) ([]*`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:91
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:91
	qw422016.N().S(`Model, error) {
    ps := ReadAll`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:92
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:92
	qw422016.N().S(`(tx)
    return ps.Run()
}

`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:96
	if t.HasID {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:96
		qw422016.N().S(`type ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:97
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:97
		qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.N().S(`(tx *sqlite.Conn) *ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.N().S(`Stmt {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:101
		qw422016.N().S(`
SELECT
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:104
		for i, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:104
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:105
			qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:105
			if i < len(t.Fields)-1 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:105
				qw422016.N().S(`,`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:105
			}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:105
			qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:106
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:106
		qw422016.N().S(`FROM `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:107
		qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:107
		qw422016.N().S(`
WHERE id = ?
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:107
		qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:107
		qw422016.N().S(`)
    return &ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:110
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:110
		qw422016.N().S(`Stmt{stmt: stmt}
}

func (ps *ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:113
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:113
		qw422016.N().S(`Stmt) Run(id int64) (*`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:113
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:113
		qw422016.N().S(`Model, error) {
    defer ps.stmt.Reset()

    ps.stmt.BindInt64(1, id)

    if hasRow, err := ps.stmt.Step(); err != nil {
        return nil, fmt.Errorf("failed to read `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:119
		qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:119
		qw422016.N().S(`: %w", err)
    } else if !hasRow {
        return nil, nil
    }

    m := &`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:124
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:124
		qw422016.N().S(`Model{}
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:125
		streamfillResStruct(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:125
		qw422016.N().S(`

    return m, nil
}

func OnceReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:130
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:130
		qw422016.N().S(`(tx *sqlite.Conn, id int64) (*`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:130
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:130
		qw422016.N().S(`Model, error) {
    ps := ReadByID`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:131
		qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:131
		qw422016.N().S(`(tx)
    return ps.Run(id)
}
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:134
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:134
	qw422016.N().S(`
func Count`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:136
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:136
	qw422016.N().S(`(tx *sqlite.Conn) (int64, error) {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:136
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:136
	qw422016.N().S(`
SELECT COUNT(*)
FROM `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:139
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:139
	qw422016.N().S(`
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:139
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:139
	qw422016.N().S(`)
    defer stmt.Reset()

    if hasRow, err := stmt.Step(); err != nil {
        return 0, fmt.Errorf("failed to count `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:144
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:144
	qw422016.N().S(`: %w", err)
    } else if !hasRow {
        return 0, nil
    }

    return stmt.ColumnInt64(0), nil
}

func OnceCount`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:152
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:152
	qw422016.N().S(`(tx *sqlite.Conn) (int64, error) {
    return Count`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:153
	qw422016.E().S(t.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:153
	qw422016.N().S(`(tx)
}

type Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:156
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:156
	qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.N().S(`(tx *sqlite.Conn) *Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.N().S(`Stmt {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:160
	qw422016.N().S(`
UPDATE `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:162
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:162
	qw422016.N().S(`
SET
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:164
	for i, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:165
		if i > 0 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:165
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			qw422016.N().S(` = ?`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			qw422016.N().D(i + 1)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			if i < len(t.Fields)-1 {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
				qw422016.N().S(`,`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:166
			qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:167
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:168
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:168
	qw422016.N().S(`WHERE id = ?1
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:168
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:168
	qw422016.N().S(`)
    return &Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:171
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:171
	qw422016.N().S(`Stmt{stmt: stmt}
}

func (ps *Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:174
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:174
	qw422016.N().S(`Stmt) Run(m *`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:174
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:174
	qw422016.N().S(`Model) error {
    defer ps.stmt.Reset()

    // Bind parameters
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:178
	streambindFields(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:178
	qw422016.N().S(`

    if _, err := ps.stmt.Step(); err != nil {
        return fmt.Errorf("failed to update `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:181
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:181
	qw422016.N().S(`: %w", err)
    }

    return nil
}

func OnceUpdate`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:187
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:187
	qw422016.N().S(`(tx *sqlite.Conn, m *`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:187
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:187
	qw422016.N().S(`Model) error {
    ps := Update`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:188
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:188
	qw422016.N().S(`(tx)
    return ps.Run(m)
}

type Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:192
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:192
	qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.N().S(`(tx *sqlite.Conn) *Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.N().S(`Stmt {
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:196
	qw422016.N().S(`
DELETE FROM `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:198
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:198
	qw422016.N().S(`
WHERE id = ?
    `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:198
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:198
	qw422016.N().S(`)
    return &Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:201
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:201
	qw422016.N().S(`Stmt{stmt: stmt}
}

func (ps *Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:204
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:204
	qw422016.N().S(`Stmt) Run(id int64) error {
    defer ps.stmt.Reset()

    ps.stmt.BindInt64(1, id)

    if _, err := ps.stmt.Step(); err != nil {
        return fmt.Errorf("failed to delete `)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:210
	qw422016.E().S(t.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:210
	qw422016.N().S(`: %w", err)
    }

    return nil
}

func OnceDelete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:216
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:216
	qw422016.N().S(`(tx *sqlite.Conn, id int64) error {
    ps := Delete`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:217
	qw422016.E().S(t.SingleName.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:217
	qw422016.N().S(`(tx)
    return ps.Run(id)
}

`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
func WriteGenerateCRUD(qq422016 qtio422016.Writer, t *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	StreamGenerateCRUD(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
func GenerateCRUD(t *GenerateCRUDTable) string {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	WriteGenerateCRUD(qb422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:221
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:223
func streambindFields(qw422016 *qt422016.Writer, tbl *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:224
	for _, f := range tbl.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:224
		qw422016.N().S(`    ps.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:225
		switch f.GoType.Original {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:226
		case "time.Time":
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:226
			qw422016.N().S(`            stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.N().D(f.Column)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.N().S(`, toolbelt.TimeToJulianDay(m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:227
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:228
		case "time.Duration":
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:228
			qw422016.N().S(`            stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.N().D(f.Column)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.N().S(`, toolbelt.DurationToMilliseconds(m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:229
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:230
		default:
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:230
			qw422016.N().S(`            stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.N().D(f.Column)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.N().S(`, m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:231
			qw422016.N().S(`)
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:232
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:233
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
func writebindFields(qq422016 qtio422016.Writer, tbl *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	streambindFields(qw422016, tbl)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
func bindFields(tbl *GenerateCRUDTable) string {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	writebindFields(qb422016, tbl)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:234
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:236
func streamfillResStruct(qw422016 *qt422016.Writer, t *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:237
	for i, f := range t.Fields {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:238
		switch f.GoType.Original {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:239
		case "time.Time":
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:239
			qw422016.N().S(`            m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.N().S(` = toolbelt.JulianDayToTime(ps.stmt.Column`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.N().D(i)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:240
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:241
		case "time.Duration":
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:241
			qw422016.N().S(`            m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.N().S(` = toolbelt.MillisecondsToDuration(ps.stmt.Column`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.N().D(i)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:242
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:243
		default:
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:243
			qw422016.N().S(`            m.`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.N().S(` = ps.stmt.Column`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.N().D(i)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:244
			qw422016.N().S(`)
`)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:245
		}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:246
	}
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
func writefillResStruct(qq422016 qtio422016.Writer, t *GenerateCRUDTable) {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	streamfillResStruct(qw422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
}

//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
func fillResStruct(t *GenerateCRUDTable) string {
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	writefillResStruct(qb422016, t)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/crud.qtpl:247
}
