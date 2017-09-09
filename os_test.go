package vfs

import "testing"

func TestOSFS(t *testing.T) {
	ns := NewNameSpace()
	ns.Bind("/", OS("test-fixtures/A/animals"), "/", BindAfter)
	ns.Bind("/", OS("test-fixtures/B/animals"), "/", BindAfter)
	ns.Bind("/wood", OS("test-fixtures/B/things/wood"), "/", BindAfter)
	ns.Bind("/wood2", OS("test-fixtures/B/things"), "/wood", BindAfter)
	assertIsNotExist(t, ns,
		"/1/2/B",
		"/3",
	)
	assertOSPather(t, ns, map[string]string{
		"/":            "",
		"/dogs/A-dogs": "test-fixtures/A/animals/dogs/A-dogs",
		"/dogs/B-dogs": "test-fixtures/B/animals/dogs/B-dogs",
		"/dogs":        "test-fixtures/A/animals/dogs",
		"/dogs/dogs":   "test-fixtures/A/animals/dogs/dogs",
		"/wood":        "test-fixtures/B/things/wood",
		"/wood2":       "test-fixtures/B/things/wood",
	})

	assertWalk(t, ns, `dir : /
dir : /dogs
file: /dogs/A-dogs
data: A/animals/dogs/A-dogs
file: /dogs/B-dogs
data: B/animals/dogs/B-dogs
file: /dogs/dogs
data: A/animals/dogs/dogs
dir : /wood
dir : /wood/table
file: /wood/table/B-table
data: B/things/wood/table/B-table
file: /wood/table/table
data: B/things/wood/table/table
dir : /wood/tree
file: /wood/tree/B-tree
data: B/things/wood/tree/B-tree
file: /wood/tree/tree
data: B/things/wood/tree/tree
dir : /wood2
dir : /wood2/table
file: /wood2/table/B-table
data: B/things/wood/table/B-table
file: /wood2/table/table
data: B/things/wood/table/table
dir : /wood2/tree
file: /wood2/tree/B-tree
data: B/things/wood/tree/B-tree
file: /wood2/tree/tree
data: B/things/wood/tree/tree`)
}

func TestSafeOSFS(t *testing.T) {
	assertIsSafe(t,
		SafeOS("test-fixtures/A/animals"),
		SafeOS("."),
		SafeOS(".."),
	)
	assertNotSafe(t,
		SafeOS("doesnotexist"),
	)
}
