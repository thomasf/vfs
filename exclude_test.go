package vfs

import "testing"

func TestExclude(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/1", OS(testPath("B")), "/things", BindAfter)
	ns.Bind("/2", Exclude(OS(testPath("B")), "/things/wood/table"), "/things", BindAfter)
	assertWalk(t, ns, `dir : /
dir : /1
dir : /1/wood
dir : /1/wood/table
file: /1/wood/table/B-table
data: B/things/wood/table/B-table
file: /1/wood/table/table
data: B/things/wood/table/table
dir : /1/wood/tree
file: /1/wood/tree/B-tree
data: B/things/wood/tree/B-tree
file: /1/wood/tree/tree
data: B/things/wood/tree/tree
dir : /2
dir : /2/wood
dir : /2/wood/tree
file: /2/wood/tree/B-tree
data: B/things/wood/tree/B-tree
file: /2/wood/tree/tree
data: B/things/wood/tree/tree`)

	assertIsDir(t, ns,
		"/1/wood/table/",
		"/2/wood/",
		"/2/wood/tree/",
	)

	assertIsNotExist(t, ns,
		"/2/wood/table/",
	)

}

func TestExcludeFiles(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/1", OS(testPath("B")), "/things", BindAfter)
	ns.Bind("/2", Exclude(OS(testPath("B")),
		"/things/wood/tree/B-tree",
		"/things/wood/table/NOTAFILE",
		"/things/wood/table/B-table",
	), "/things", BindAfter)
	assertOSPather(t, ns, map[string]string{
		"/2/wood/table/table": "test-fixtures/B/things/wood/table/table",
		"/1/wood":             "test-fixtures/B/things/wood",
	})
	assertWalk(t, ns, `dir : /
dir : /1
dir : /1/wood
dir : /1/wood/table
file: /1/wood/table/B-table
data: B/things/wood/table/B-table
file: /1/wood/table/table
data: B/things/wood/table/table
dir : /1/wood/tree
file: /1/wood/tree/B-tree
data: B/things/wood/tree/B-tree
file: /1/wood/tree/tree
data: B/things/wood/tree/tree
dir : /2
dir : /2/wood
dir : /2/wood/table
file: /2/wood/table/table
data: B/things/wood/table/table
dir : /2/wood/tree
file: /2/wood/tree/tree
data: B/things/wood/tree/tree`)
}
