package vfs

import "testing"

func TestFileMapFS(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/", FileMap(map[string]string{
		"1/2/3/4/5/6":   "test-fixtures/C/animals/cats/cats",
		"1/2/3/A/4/5/6": "test-fixtures/C/animals/cats/cats",
		"1/2/3/B/4/5/6": "test-fixtures/C/animals/cats/C-cats",
		"2":             "test-fixtures/C/animals/cats/cats",
	}), "/", BindReplace)

	assertIsNotExist(t, ns,
		"/1/2/B",
		"/3",
	)
	assertOSPather(t, ns, map[string]string{
		"/1/2":           "",
		"/1/2/3/A/4/5/6": "test-fixtures/C/animals/cats/cats",
		"2":              "test-fixtures/C/animals/cats/cats",
	})

	assertWalk(t, ns, `dir : /
dir : /1
dir : /1/2
dir : /1/2/3
dir : /1/2/3/4
dir : /1/2/3/4/5
file: /1/2/3/4/5/6
data: C/animals/cats/cats
dir : /1/2/3/A
dir : /1/2/3/A/4
dir : /1/2/3/A/4/5
file: /1/2/3/A/4/5/6
data: C/animals/cats/cats
dir : /1/2/3/B
dir : /1/2/3/B/4
dir : /1/2/3/B/4/5
file: /1/2/3/B/4/5/6
data: C/animals/cats/C-cats
file: /2
data: C/animals/cats/cats`)
}

func TestSafeFileMapFS(t *testing.T) {
	assertIsSafe(t, SafeFileMap(map[string]string{
		"1/2/3/4/5/6":    "test-fixtures/C/animals/cats/cats",
		"1/2/3/A/4/5/6":  "test-fixtures/C/animals/cats/cats",
		"/1/2/3/B/4/5/6": "test-fixtures/C/animals/cats/C-cats",
		"2":              "test-fixtures/C/animals/cats/cats",
	}))

	assertNotSafe(t, SafeFileMap(map[string]string{
		"1/2/3/4/5/6":   "test-fixtures/C/animals/cats/cats",
		"1/2/3/B/4/5/6": "test-fixtures/C/animals/cats/NO",
	}))
	{
		ns := FileMap(map[string]string{
			"1/2/3":   "test-fixtures/C/animals/cats/cats",
			"1/A/2/3": "test-fixtures/C/animals/cats/cats",
			"1/B/2/4": "test-fixtures/C/animals/cats/C-cats",
			"2":       "test-fixtures/C/animals/cats/cats",
			"/3":      "test-fixtures/C/animals/cats/cats",
		})
		// only SafeBind fixes bad input
		assertIsNotExist(t, ns,
			"/3",
		)
	}
	{
		ns := NewNameSpace()
		bindOrDie(t, ns.BindSafe("/", SafeFileMap(map[string]string{
			"1/f":  "test-fixtures/C/animals/cats/cats",
			"1/f2": "test-fixtures/C/animals/cats/cats",
			"2":    "test-fixtures/C/animals/cats/cats",
			"/3/f": "test-fixtures/C/animals/cats/cats",
			"4/f":  "test-fixtures/C/animals/cats/C-cats",
		}), "/", BindReplace))
		assertIsDir(t, ns,
			"/1",
			"/3",
			"/4",
		)
		assertIsRegular(t, ns,
			"/2",
			"/1/f2",
			"/1/f",
			"/3/f",
			"/4/f",
		)
		assertIsNotExist(t, ns,
			"/1/2",
		)
	}
}
