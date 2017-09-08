package vfs

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestNewNameSpace(t *testing.T) {
	// We will mount this filesystem under /fs1
	mount := Map(map[string]string{"fs1file": "abcdefgh"})

	// Existing process. This should give error on Stat("/")
	t1 := NameSpace{}
	t1.Bind("/fs1", mount, "/", BindReplace)

	// using NewNameSpace. This should work fine.
	t2 := NewNameSpace()
	t2.Bind("/fs1", mount, "/", BindReplace)

	testcases := map[string][]bool{
		"/":            {false, true},
		"/fs1":         {true, true},
		"/fs1/fs1file": {true, true},
	}

	fss := []FileSystem{t1, t2}

	for j, fs := range fss {
		for k, v := range testcases {
			_, err := fs.Stat(k)
			result := err == nil
			if result != v[j] {
				t.Errorf("fs: %d, testcase: %s, want: %v, got: %v, err: %s", j, k, v[j], result, err)
			}
		}
	}

	fi, err := t2.Stat("/")
	if err != nil {
		t.Fatal(err)
	}

	if fi.Name() != "/" {
		t.Errorf("t2.Name() : want:%s got:%s", "/", fi.Name())
	}

	if !fi.ModTime().IsZero() {
		t.Errorf("t2.Modime() : want:%v got:%v", time.Time{}, fi.ModTime())
	}
}

func TestIntermediateEmtpyDirs(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/1/2/3/4/5/6", OneFile(testPath("C/animals/cats/cats"), "fake-dog1"), "/", BindBefore)
	ns.Bind("/1/2/3/4/5/6", OneFile(testPath("C/animals/cats/cats"), "fake-dog2"), "/", BindBefore)
	ns.Bind("/1/2/3/A/4/5/6", OneFile(testPath("C/animals/cats/cats"), "fake-dog3"), "/", BindBefore)
	ns.Bind("/1/2/3/A/4/5/6", OneFile(testPath("C/animals/cats/cats"), "fake-dog4"), "/", BindBefore)
	ns.Bind("/1/2/3/B/4/5/6", OneFile(testPath("C/animals/cats/cats"), "fake-dog5"), "/", BindBefore)
	ns.Bind("/1", OS(testPath("C")), "/", BindAfter)
	assertIsDir(t, ns,
		"/1/2",
		"/1/2/3",
		"/1/2/3/4",
		"/1/2/3/A",
		"/1/2/3/A/4",
		"/1/2/3/B/4",
		"/1/2/3/4/5/6",
		"/1/animals/",
		"/1/animals/cats",
	)
	assertIsRegular(t, ns,
		"/1/2/3/4/5/6/fake-dog1",
		"/1/2/3/4/5/6/fake-dog2",
		"/1/2/3/A/4/5/6/fake-dog3",
		"/1/2/3/A/4/5/6/fake-dog4",
		"/1/2/3/B/4/5/6/fake-dog5",
		"/1/animals/cats/cats",
	)
	assertIsNotExist(t, ns,
		"/1/2/3/4/5/7",
		"/2",
		"/1/3",
		"/1/animals/cats/dogs",
	)
	assertOSPather(t, ns, map[string]string{
		"/1/2/3/4/5":   "",
		"/1/2/3/4/5/6": "",
	})
}
func TestSafeBind(t *testing.T) {
	m := map[string]string{
		"1/6":   "test-fixtures/C/animals/cats/cats",
		"1/A/6": "test-fixtures/C/animals/cats/cats",
		"1/B/6": "test-fixtures/C/animals/cats/C-cats",
		"2":             "test-fixtures/C/animals/cats/cats",
	}
	// ensure that the filesytem itself is safe
	_ = safeOrDie(t, SafeFileMap(m))

	const (
		nx = "file does not exist"
		nd = "not a directory"
		ok = ""
	)
	testCases := [][]string{
		{"/", ok},
		{"/boo", nx},
		{"/", ok},
		{"/things", nx},
		{"/1/A/", ok},
		{"/1", ok},
		{"/1/", ok},
		{"/1/", ok},
		{"/1/A/6", nd},
		{"/1/B/6", nd},
		{"/1/B/", ok},
		{"/2", nd},
	}

	ns := NewNameSpace()
tests:
	for i, tc := range testCases {
		new := tc[0]
		estr := tc[1]
		err := ns.BindSafe(fmt.Sprintf("/%v", i), SafeMap(m), new, BindAfter)
		if err != nil {
			if estr != "" && strings.Contains(err.Error(), estr) {
				continue tests
			}
			spew.Dump(err, new, estr)
			t.Fatal(err)
		}
	}
	assertWalk(t, ns, `dir : /
dir : /0
dir : /0/1
file: /0/1/6
data: test-fixtures/C/animals/cats/cats
dir : /0/1/A
file: /0/1/A/6
data: test-fixtures/C/animals/cats/cats
dir : /0/1/B
file: /0/1/B/6
data: test-fixtures/C/animals/cats/C-cats
file: /0/2
data: test-fixtures/C/animals/cats/cats
dir : /10
file: /10/6
data: test-fixtures/C/animals/cats/C-cats
dir : /2
dir : /2/1
file: /2/1/6
data: test-fixtures/C/animals/cats/cats
dir : /2/1/A
file: /2/1/A/6
data: test-fixtures/C/animals/cats/cats
dir : /2/1/B
file: /2/1/B/6
data: test-fixtures/C/animals/cats/C-cats
file: /2/2
data: test-fixtures/C/animals/cats/cats
dir : /4
file: /4/6
data: test-fixtures/C/animals/cats/cats
dir : /5
file: /5/6
data: test-fixtures/C/animals/cats/cats
dir : /5/A
file: /5/A/6
data: test-fixtures/C/animals/cats/cats
dir : /5/B
file: /5/B/6
data: test-fixtures/C/animals/cats/C-cats
dir : /6
file: /6/6
data: test-fixtures/C/animals/cats/cats
dir : /6/A
file: /6/A/6
data: test-fixtures/C/animals/cats/cats
dir : /6/B
file: /6/B/6
data: test-fixtures/C/animals/cats/C-cats
dir : /7
file: /7/6
data: test-fixtures/C/animals/cats/cats
dir : /7/A
file: /7/A/6
data: test-fixtures/C/animals/cats/cats
dir : /7/B
file: /7/B/6
data: test-fixtures/C/animals/cats/C-cats`)
}

func TestFprint(t *testing.T) {
	var buf bytes.Buffer
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/excl", Exclude(OS(testPath("B")), "/things/wood/table"), "/things", BindAfter)
	ns.Bind("/dogs", OS(testPath("A/animals/dogs")), "/", BindAfter)
	ns.Bind("/dogs", OS(testPath("B/animals/dogs")), "/", BindBefore)
	ns.Bind("/new/dogs", OneFile(testPath("C/animals/cats/cats"), "fake-dog"), "/", BindBefore)
	ns.Bind("/fm/dogs", FileMap(map[string]string{"dogs/fake-dog": testPath("C/animals/cats/cats")}), "/", BindBefore)
	ns.Bind("/mapdogs", Map(map[string]string{"fake-dog": ""}), "/", BindBefore)

	ns.Fprint(&buf)
	s := buf.String()
	expected := `name space {
	/:
		ns /
	/dogs:
		os(test-fixtures/B/animals/dogs) /
		ns /dogs
		os(test-fixtures/A/animals/dogs) /
	/excl:
		ns /excl
		exclude(os(test-fixtures/B)) /things
	/fm/dogs:
		filemap(1) /
		ns /fm/dogs
	/mapdogs:
		filemap(1) /
		ns /mapdogs
	/new/dogs:
		onefile(test-fixtures/C/animals/cats/cats:fake-dog) /
		ns /new/dogs
}
`
	if s != expected {
		fmt.Println(s)
		t.Log("GOT")
		t.Log(s)
		t.Log("EXPECTED")
		t.Log(expected)
		t.Fatal()
	}

}

func TestComplicated(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/dogs", OS(testPath("A/animals/dogs")), "/", BindAfter)
	ns.Bind("/dogs", OS(testPath("B/animals/dogs")), "/", BindAfter)
	ns.Bind("/dogs/subdogs", OS(testPath("B/animals/dogs")), "/", BindAfter)
	ns.Bind("/dogs/subdogs/sub2/sub3", OS(testPath("A/animals/dogs")), "/", BindAfter)
	ns.Bind("/dogs", OneFile(testPath("C/animals/cats/cats"), "fake-dog"), "/", BindBefore)
	ns.Bind("/alt/dogs", OS(testPath("A/animals/dogs")), "/", BindAfter)
	ns.Bind("/new/dogs", OneFile(testPath("C/animals/cats/cats"), "fake-dog"), "/", BindBefore)
	ns.Bind("/all", OS(testPath("A")), "/", BindBefore)
	ns.Bind("/all", OS(testPath("B")), "/", BindBefore)
	ns.Bind("/all", OS(testPath("C")), "/", BindAfter)

	assertIsDir(t, ns,
		"/dogs/subdogs/",
	)

	assertIsRegular(t, ns,
		"/dogs/subdogs/B-dogs",
	)
	assertIsNotExist(t, ns,
		"/dogs/subdogs/B-dogs/nodog",
		"/dogs/subdogsnodog",
	)

	assertWalk(t, ns, `dir : /
dir : /all
dir : /all/animals
dir : /all/animals/cats
file: /all/animals/cats/C-cats
data: C/animals/cats/C-cats
file: /all/animals/cats/cats
data: C/animals/cats/cats
dir : /all/animals/dogs
file: /all/animals/dogs/A-dogs
data: A/animals/dogs/A-dogs
file: /all/animals/dogs/B-dogs
data: B/animals/dogs/B-dogs
file: /all/animals/dogs/dogs
data: B/animals/dogs/dogs
dir : /all/ships
dir : /all/ships/battleships
file: /all/ships/battleships/A-battleships
data: A/ships/battleships/A-battleships
file: /all/ships/battleships/battleships
data: A/ships/battleships/battleships
dir : /all/things
dir : /all/things/wood
dir : /all/things/wood/table
file: /all/things/wood/table/B-table
data: B/things/wood/table/B-table
file: /all/things/wood/table/table
data: B/things/wood/table/table
dir : /all/things/wood/tree
file: /all/things/wood/tree/B-tree
data: B/things/wood/tree/B-tree
file: /all/things/wood/tree/tree
data: B/things/wood/tree/tree
dir : /alt
dir : /alt/dogs
file: /alt/dogs/A-dogs
data: A/animals/dogs/A-dogs
file: /alt/dogs/dogs
data: A/animals/dogs/dogs
dir : /dogs
file: /dogs/A-dogs
data: A/animals/dogs/A-dogs
file: /dogs/B-dogs
data: B/animals/dogs/B-dogs
file: /dogs/dogs
data: A/animals/dogs/dogs
file: /dogs/fake-dog
data: C/animals/cats/cats
dir : /dogs/subdogs
file: /dogs/subdogs/B-dogs
data: B/animals/dogs/B-dogs
file: /dogs/subdogs/dogs
data: B/animals/dogs/dogs
dir : /dogs/subdogs/sub2
dir : /dogs/subdogs/sub2/sub3
file: /dogs/subdogs/sub2/sub3/A-dogs
data: A/animals/dogs/A-dogs
file: /dogs/subdogs/sub2/sub3/dogs
data: A/animals/dogs/dogs
dir : /new
dir : /new/dogs
file: /new/dogs/fake-dog
data: C/animals/cats/cats`)
}
