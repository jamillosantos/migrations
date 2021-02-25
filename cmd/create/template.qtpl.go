// Code generated by qtc from "template.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line template.qtpl:1
package create

//line template.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line template.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line template.qtpl:1
func StreamMigration(qw422016 *qt422016.Writer, input *TemplateInput) {
//line template.qtpl:1
	qw422016.N().S(`
package `)
//line template.qtpl:2
	qw422016.E().S(input.Package)
//line template.qtpl:2
	qw422016.N().S(`

import (
	"github.com/jamillosantos/migrations"
	"github.com/jamillosantos/migrations/code"
)

var _ = migrations.DefaultSource.Add(code.MustNew(&code.Migration{
	Do: func(_ migrations.ExecutionContext) error {
		panic("not implemented")
	},
`)
//line template.qtpl:13
	if input.DontUndo {
//line template.qtpl:13
		qw422016.N().S(`
	Undo: func(_ migrations.ExecutionContext) error {
		panic("not implemented")
	},
`)
//line template.qtpl:17
	}
//line template.qtpl:17
	qw422016.N().S(`}))

`)
//line template.qtpl:20
}

//line template.qtpl:20
func WriteMigration(qq422016 qtio422016.Writer, input *TemplateInput) {
//line template.qtpl:20
	qw422016 := qt422016.AcquireWriter(qq422016)
//line template.qtpl:20
	StreamMigration(qw422016, input)
//line template.qtpl:20
	qt422016.ReleaseWriter(qw422016)
//line template.qtpl:20
}

//line template.qtpl:20
func Migration(input *TemplateInput) string {
//line template.qtpl:20
	qb422016 := qt422016.AcquireByteBuffer()
//line template.qtpl:20
	WriteMigration(qb422016, input)
//line template.qtpl:20
	qs422016 := string(qb422016.B)
//line template.qtpl:20
	qt422016.ReleaseByteBuffer(qb422016)
//line template.qtpl:20
	return qs422016
//line template.qtpl:20
}