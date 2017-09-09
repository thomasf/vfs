package vfs

import "testing"

func TestExclude(t *testing.T) {

	{
		ns := NewNameSpace()
		ns.Bind("/1", Exclude(OS(testPath("B")), "/things/wood/tree"), "/", BindAfter)
		assertIsDir(t, ns,
			"/1/things/wood/table",
			"/1/things",
		)
		assertIsNotExist(t, ns,
			"/1/things/wood/tree",
		)
		assertWalk(t, ns, ``)

	}

	{
		ns := NewNameSpace()
		ns.Bind("/1", Exclude(OS(testPath("B/things")),
			"/wood/table",
			"/wood/tree/B-tree",
		), "/wood/", BindAfter)

		assertIsDir(t, ns,
			"/1/tree/",
		)
		assertIsNotExist(t, ns,
			"/1/table",
		)

		assertWalk(t, ns, `dir : /
dir : /1
dir : /1/tree
file: /1/tree/tree
data: B/things/wood/tree/tree`)

		assertIsNotExist(t, ns,
			"/2/wood/table/",
		)

	}
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

func TestSafeExclude(t *testing.T) {
	ns := NewNameSpace()
	bindOrDie(t, ns.BindSafe("/",
		SafeExclude(SafeOS(testPath("B")),
			"/things/wood/table",
			"things/wood/tree",
		),
		"/things", BindAfter))
	assertWalk(t, ns, "s")

	bindOrDie(t, ns.BindSafe("/a",
		SafeExclude(SafeOS(testPath("B")),
			"/things/wood/table",
			"things/wood/tree",
			"/animals/dogs/dogs",
		),
		"/animals", BindAfter))

	assertIsDir(t, ns,
		"/wood/",
		"/aninmals",
		"/dogs",
	)
	assertIsNotExist(t, ns,
		"/2/tree/",
		"/2/table/",
	)

	assertWalk(t, ns, `dir   : /
dir   : /1
dir   : /1/table
file  : /1/table/B-table
data  : B/things/wood/table/B-table
file  : /1/table/table
data  : B/things/wood/table/table
dir   : /1/tree
file  : /1/tree/B-tree
data  : B/things/wood/tree/B-tree
file  : /1/tree/tree
data  : B/things/wood/tree/tree
dir   : /2
dir   : /2/wood`)

}
