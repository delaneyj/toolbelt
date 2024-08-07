// Code generated by qtc from "queries.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:1
package zombiezen

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:1
import "github.com/delaneyj/toolbelt"

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:3
func StreamGenerateQuery(qw422016 *qt422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:3
	qw422016.N().S(`
// Code generated by "sqlc-gen-zombiezen". DO NOT EDIT.

package `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:6
	qw422016.E().S(q.PackageName.Lower)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:6
	qw422016.N().S(`

import (
    "fmt"
    "zombiezen.com/go/sqlite"

    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:12
	if q.NeedsTimePackage {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:12
		qw422016.N().S(`
    "time"
    "github.com/delaneyj/toolbelt"
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:15
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:15
	qw422016.N().S(`
)


`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:19
	if q.HasResponse && !q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:19
		qw422016.N().S(`
type `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:20
		qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:20
		qw422016.N().S(`Res struct {
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:21
		for _, f := range q.ResponseFields {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:21
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.E().S(f.GoType.Original)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S(`json:"`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S(`"`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:22
			qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:23
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:23
		qw422016.N().S(`}
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:25
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:25
	qw422016.N().S(`

`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:27
	if q.HasParams && !q.ParamsIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:27
		qw422016.N().S(`
type `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:28
		qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:28
		qw422016.N().S(`Params struct {
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:29
		for _, f := range q.Params {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:29
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.E().S(f.GoType.Original)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S(`json:"`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.E().S(f.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S(`"`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:30
			qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:31
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:31
		qw422016.N().S(`}
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:33
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:33
	qw422016.N().S(`

type `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:35
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:35
	qw422016.N().S(`Stmt struct {
    stmt *sqlite.Stmt
}

func `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.N().S(`(tx *sqlite.Conn) *`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.N().S(`Stmt {
    // Prepare the statement into connection cache
    stmt := tx.Prep(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:39
	qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:42
	qw422016.N().S(q.SQL)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:42
	qw422016.N().S(`
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:42
	qw422016.N().S("`")
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:42
	qw422016.N().S(`)
    ps := &`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:44
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:44
	qw422016.N().S(`Stmt{stmt: stmt}
    return ps
}

func (ps *`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:48
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:48
	qw422016.N().S(`Stmt) Run(
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:49
	streamfillReqParams(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:49
	qw422016.N().S(`) (
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:51
	streamfillReturns(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:51
	qw422016.N().S(`) {
    defer ps.stmt.Reset()

`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:55
	if len(q.Params) > 0 {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:55
		qw422016.N().S(`
    // Bind parameters
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:57
		for _, p := range q.Params {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:57
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:58
			streamfillParams(qw422016, p.GoType, p.SQLType, p.Name, p.Column, q.ParamsIsSingularField)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:58
			qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:59
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:60
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:60
	qw422016.N().S(`
    // Execute the query
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:63
	if q.HasResponse {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:64
		if q.ResponseHasMultiple {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:64
			qw422016.N().S(`        for {
            if hasRow, err := ps.stmt.Step(); err != nil {
                return res, fmt.Errorf("failed to execute {{.Name.Lower}} SQL: %w", err)
            } else if !hasRow {
                break
            }
            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:71
			streamfillResponse(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:71
			qw422016.N().S(`
            res = append(res, row)
        }
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:75
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:75
			qw422016.N().S(`        if hasRow, err := ps.stmt.Step(); err != nil {
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:77
			if q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:77
				qw422016.N().S(`            return `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:78
				qw422016.E().S(q.ResponseFields[0].Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:78
				qw422016.N().S(`, fmt.Errorf("failed to execute {{.Name.Lower}} SQL: %w", err)
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:79
			} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:79
				qw422016.N().S(`            return res, err
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:81
			}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:81
			qw422016.N().S(`        } else if hasRow {
            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:83
			streamfillResponse(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:84
			if !q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:84
				qw422016.N().S(`            res = &row
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:86
			}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:86
			qw422016.N().S(`        }
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:88
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:89
	} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:89
		qw422016.N().S(`    if _, err := ps.stmt.Step(); err != nil {
        return fmt.Errorf("failed to execute `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:91
		qw422016.E().S(q.Name.Lower)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:91
		qw422016.N().S(` SQL: %w", err)
    }
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:93
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:93
	qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:95
	if q.HasResponse {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:96
		if q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:96
			qw422016.N().S(`    return `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:97
			qw422016.E().S(q.ResponseFields[0].Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:97
			qw422016.N().S(`, nil
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:98
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:98
			qw422016.N().S(`    return res, nil
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:100
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:101
	} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:101
		qw422016.N().S(`    return nil
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:103
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:103
	qw422016.N().S(`}

func Once`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:106
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:106
	qw422016.N().S(`(
    tx *sqlite.Conn,
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:108
	streamfillReqParams(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:108
	qw422016.N().S(`) (
    `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:110
	streamfillReturns(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:110
	qw422016.N().S(`) {
    ps := `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:112
	qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:112
	qw422016.N().S(`(tx)

    return ps.Run(
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:115
	if q.HasParams {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:115
		qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:116
		if q.ParamsIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:116
			qw422016.N().S(`            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:117
			qw422016.E().S(q.Params[0].Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:117
			qw422016.N().S(`,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:118
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:118
			qw422016.N().S(`            params,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:120
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:121
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:121
	qw422016.N().S(`
    )
}

`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
func WriteGenerateQuery(qq422016 qtio422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	StreamGenerateQuery(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
func GenerateQuery(q *GenerateQueryContext) string {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	WriteGenerateQuery(qb422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:125
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:127
func streamfillResponse(qw422016 *qt422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:127
	qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:128
	if q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:130
		f := q.ResponseFields[0]

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:132
		switch f.GoType.Original {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:133
		case "time.Time":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:133
			qw422016.N().S(`            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:134
			qw422016.E().S(f.Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:134
			qw422016.N().S(` = toolbelt.JulianDayToTime(ps.stmt.ColumnFloat(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:134
			qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:134
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:135
		case "time.Duration":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:135
			qw422016.N().S(`            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:136
			qw422016.E().S(f.Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:136
			qw422016.N().S(` = toolbelt.MillisecondsToDuration(ps.stmt.ColumnInt64(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:136
			qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:136
			qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:137
		default:
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:137
			qw422016.N().S(`            `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.E().S(f.Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.N().S(` = ps.stmt.Column`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:138
			qw422016.N().S(`)
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:139
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:140
	} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:140
		qw422016.N().S(`    row := `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:141
		qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:141
		qw422016.N().S(`Res{
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:142
		for _, f := range q.ResponseFields {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:143
			switch f.GoType.Original {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:144
			case "time.Time":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:144
				qw422016.N().S(`                `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:145
				qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:145
				qw422016.N().S(`: toolbelt.JulianDayToTime(ps.stmt.ColumnFloat(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:145
				qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:145
				qw422016.N().S(`)),
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:146
			case "time.Duration":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:146
				qw422016.N().S(`                `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:147
				qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:147
				qw422016.N().S(`: toolbelt.MillisecondsToDuration(ps.stmt.ColumnInt64(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:147
				qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:147
				qw422016.N().S(`)),
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:148
			default:
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:148
				qw422016.N().S(`                `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.E().S(f.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.N().S(`: ps.stmt.Column`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.E().S(f.SQLType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.N().D(f.Offset)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:149
				qw422016.N().S(`),
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:150
			}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:151
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:151
		qw422016.N().S(`    }
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:153
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:153
	qw422016.N().S(`
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
func writefillResponse(qq422016 qtio422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	streamfillResponse(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
func fillResponse(q *GenerateQueryContext) string {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	writefillResponse(qb422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:154
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:156
func streamfillReqParams(qw422016 *qt422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:157
	if q.HasParams {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:158
		if q.ParamsIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:158
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:159
			qw422016.E().S(q.Params[0].Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:159
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:159
			qw422016.E().S(q.Params[0].GoType.Original)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:159
			qw422016.N().S(`,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:160
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:160
			qw422016.N().S(`        params `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:161
			qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:161
			qw422016.N().S(`Params,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:162
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:163
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
func writefillReqParams(qq422016 qtio422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	streamfillReqParams(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
func fillReqParams(q *GenerateQueryContext) string {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	writefillReqParams(qb422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:164
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:167
func streamfillReturns(qw422016 *qt422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:168
	if q.HasResponse {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:169
		if q.ResponseIsSingularField {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:169
			qw422016.N().S(`        `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:170
			qw422016.E().S(q.ResponseFields[0].Name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:170
			qw422016.N().S(` `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:170
			qw422016.E().S(q.ResponseFields[0].GoType.Original)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:170
			qw422016.N().S(`,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:171
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:171
			qw422016.N().S(`        res `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
			if q.ResponseHasMultiple {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
				qw422016.N().S(`[]`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
			} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
				qw422016.N().S(`*`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
			}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
			qw422016.E().S(q.Name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:172
			qw422016.N().S(`Res,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:173
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:174
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:174
	qw422016.N().S(`    err error,
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
func writefillReturns(qq422016 qtio422016.Writer, q *GenerateQueryContext) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	streamfillReturns(qw422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
func fillReturns(q *GenerateQueryContext) string {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	writefillReturns(qb422016, q)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:176
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:178
func streamfillParams(qw422016 *qt422016.Writer, goType, sqlType, name toolbelt.CasedString, col int, isSingle bool) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:179
	switch goType.Original {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:180
	case "time.Time":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:180
		qw422016.N().S(`        ps.stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		qw422016.E().S(sqlType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		qw422016.N().D(col)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		qw422016.N().S(`, toolbelt.TimeToJulianDay(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		if isSingle {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
			qw422016.E().S(name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
			qw422016.N().S(`params.`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
			qw422016.E().S(name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:181
		qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:182
	case "time.Duration":
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:182
		qw422016.N().S(`        ps.stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		qw422016.E().S(sqlType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		qw422016.N().D(col)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		qw422016.N().S(`, toolbelt.DurationToMilliseconds(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		if isSingle {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
			qw422016.E().S(name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
			qw422016.N().S(`params.`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
			qw422016.E().S(name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:183
		qw422016.N().S(`))
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:184
	default:
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:184
		qw422016.N().S(`        ps.stmt.Bind`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		qw422016.E().S(sqlType.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		qw422016.N().S(`(`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		qw422016.N().D(col)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		qw422016.N().S(`, `)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		if isSingle {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
			qw422016.E().S(name.Camel)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		} else {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
			qw422016.N().S(`params.`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
			qw422016.E().S(name.Pascal)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:185
		qw422016.N().S(`)
`)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:186
	}
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
func writefillParams(qq422016 qtio422016.Writer, goType, sqlType, name toolbelt.CasedString, col int, isSingle bool) {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	qw422016 := qt422016.AcquireWriter(qq422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	streamfillParams(qw422016, goType, sqlType, name, col, isSingle)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	qt422016.ReleaseWriter(qw422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
}

//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
func fillParams(goType, sqlType, name toolbelt.CasedString, col int, isSingle bool) string {
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	qb422016 := qt422016.AcquireByteBuffer()
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	writefillParams(qb422016, goType, sqlType, name, col, isSingle)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	qs422016 := string(qb422016.B)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	qt422016.ReleaseByteBuffer(qb422016)
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
	return qs422016
//line sqlc-gen-zombiezen/zombiezen/queries.qtpl:187
}

// `
