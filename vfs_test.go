package vfs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestModeMap(t *testing.T) {
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)

	fm := map[string]string{
		"1/2/a": "test-fixtures/C/animals/cats/cats",
		"1/2/b": "test-fixtures/C/animals/cats/cats",
		"c":     "test-fixtures/C/animals/cats/cats",
	}
	mm := map[string]os.FileMode{
		"":      0777,
		"1":     0767,
		"1/2":   0766,
		"1/2/a": 0077,
		"c":     0737,
	}

	ns.Bind("/", ModeMap(FileMap(fm), mm), "/", BindReplace)

	for k, v := range map[string]os.FileMode{
		"":      0777,
		"1":     0767,
		"1/2":   0766,
		"1/2/a": 0077,
		"1/2/b": 0660,
		"c":     0737,
	} {
		fi, err := ns.Stat(k)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode() != v {
			t.Fatalf("not equal modes %s (%o) %s (%o): %s", fi.Mode(), fi.Mode(), v, v, k)
		}

	}
	Walk("/", ns, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			t.Logf("path:%s err:%v (%T)", p, err, err)
			return fmt.Errorf("ERROR: %s : %v !!!", p, err)
		}
		if info.IsDir() {
			fmt.Println("dir", p, info.Mode())
			return nil
		}
		fmt.Println("file", p)
		f, err := ns.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("data", string(data))
		return nil
	})

}

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
	)
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

func TestMapFSOpenRoot(t *testing.T) {
	fs := Map(map[string]string{
		"foo/bar/three.txt": "a",
		"foo/bar.txt":       "b",
		"top.txt":           "c",
		"other-top.txt":     "d",
	})
	tests := []struct {
		path string
		want string
	}{
		{"/foo/bar/three.txt", "a"},
		{"foo/bar/three.txt", "a"},
		{"foo/bar.txt", "b"},
		{"top.txt", "c"},
		{"/top.txt", "c"},
		{"other-top.txt", "d"},
		{"/other-top.txt", "d"},
	}
	for _, tt := range tests {
		rsc, err := fs.Open(tt.path)
		if err != nil {
			t.Errorf("Open(%q) = %v", tt.path, err)
			continue
		}
		slurp, err := ioutil.ReadAll(rsc)
		if err != nil {
			t.Error(err)
		}
		if string(slurp) != tt.want {
			t.Errorf("Read(%q) = %q; want %q", tt.path, tt.want, slurp)
		}
		rsc.Close()
	}

	_, err := fs.Open("/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("ReadDir /xxxx = %v; want os.IsNotExist error", err)
	}
}

func TestMapFSReaddir(t *testing.T) {
	fs := Map(map[string]string{
		"foo/bar/three.txt": "333",
		"foo/bar.txt":       "22",
		"top.txt":           "top.txt file",
		"other-top.txt":     "other-top.txt file",
	})
	tests := []struct {
		dir  string
		want []os.FileInfo
	}{
		{
			dir: "/",
			want: []os.FileInfo{
				mapFI{name: "foo", dir: true},
				mapFI{name: "other-top.txt", size: len("other-top.txt file")},
				mapFI{name: "top.txt", size: len("top.txt file")},
			},
		},
		{
			dir: "/foo",
			want: []os.FileInfo{
				mapFI{name: "bar", dir: true},
				mapFI{name: "bar.txt", size: 2},
			},
		},
		{
			dir: "/foo/",
			want: []os.FileInfo{
				mapFI{name: "bar", dir: true},
				mapFI{name: "bar.txt", size: 2},
			},
		},
		{
			dir: "/foo/bar",
			want: []os.FileInfo{
				mapFI{name: "three.txt", size: 3},
			},
		},
	}
	for _, tt := range tests {
		fis, err := fs.ReadDir(tt.dir)
		if err != nil {
			t.Errorf("ReadDir(%q) = %v", tt.dir, err)
			continue
		}
		if !reflect.DeepEqual(fis, tt.want) {
			t.Errorf("ReadDir(%q) = %#v; want %#v", tt.dir, fis, tt.want)
			continue
		}
	}

	_, err := fs.ReadDir("/xxxx")
	if !os.IsNotExist(err) {
		t.Errorf("ReadDir /xxxx = %v; want os.IsNotExist error", err)
	}
}

// if they need to be regenerated
func generateTestFixture() {
	testPath := func(path string) string {
		return filepath.Join("test-fixtures", path)
	}

	for _, f := range []string{
		"A/animals/dogs",
		// "A/ships/battleships",
		// "B/animals/dogs",
		"B/things/wood/table",
		"B/things/wood/tree",
		"C/animals/cats",
	} {
		root := string(f[0])
		if err := os.MkdirAll(testPath(f), 0770); err != nil {
			log.Fatal(err)
		}

		for _, fn := range []string{
			filepath.Join(f, filepath.Base(f)),                             // common between all directories
			filepath.Join(f, fmt.Sprintf("%s-%s", root, filepath.Base(f))), // unique from every root
		} {
			if err := ioutil.WriteFile(testPath(fn), []byte(fn), 0660); err != nil {
				log.Fatal(err)
			}
			log.Println(fn)
		}
	}
}

func testPath(path string) string {
	return filepath.Join("test-fixtures", path)
}

func assertIsRegular(t *testing.T, ns NameSpace, paths ...string) {
	for _, fn := range paths {
		fi, err := ns.Stat(fn)
		if err != nil {
			t.Fatal(err)
		}
		if !fi.Mode().IsRegular() {
			t.Fatal(fi.Mode().String())
		}
		f, err := ns.Open(fn)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

	}
}

func assertIsDir(t *testing.T, ns NameSpace, paths ...string) {
	for _, fn := range paths {
		fi, err := ns.Stat(fn)
		if err != nil {
			t.Fatal(err)
		}
		if !fi.Mode().IsDir() {
			t.Fatalf("expected path as directory: %s : %v ", fn, fi.Mode().String())
		}
		if _, err := ns.ReadDir(fn); err != nil {
			t.Fatalf("expected path as directory: %s : %v ", fn, err)
		}
	}
}

func assertIsNotExist(t *testing.T, ns NameSpace, paths ...string) {
loop:
	for _, fn := range paths {
		fi, err := ns.Stat(fn)
		if err == nil {
			t.Fatalf("expected path to not exist: %s : %v ", fn, fi.Mode().String())
		}
		if !os.IsNotExist(err) {
			t.Fatalf("expected path to not exist: %s : %v ", fn, err)
		}
		_, err = ns.Open(fn)
		if os.IsNotExist(err) {
			continue loop
		}

		if _, ok := err.(*os.PathError); ok {
			continue loop
		}
		t.Logf("%T", err)
		t.Fatal(err)
	}
}

func assertWalk(t *testing.T, ns NameSpace, expected string) {
	// walkEntry .
	type walkEntry struct {
		kind, data string
	}

	var results []walkEntry
	addRes := func(kind, data string) {
		results = append(results, walkEntry{kind, data})
		// fmt.Printf("%-6s: %-30s", kind, data)
	}
	getAssertString := func() string {
		var strs []string
		for _, v := range results {
			s := fmt.Sprintf("%-4s: %s", v.kind, v.data)
			strs = append(strs, s)
		}
		return strings.Join(strs, "\n")
	}
	getPrintString := func() string {
		var strs []string
		for _, v := range results {
			s := fmt.Sprintf("%-6s: %-30s", v.kind, v.data)
			strs = append(strs, s)
		}
		return strings.Join(strs, "\n")
	}
	err := Walk("/", ns, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			t.Logf("path:%s err:%v (%T)", p, err, err)
			return fmt.Errorf("ERROR: %s : %v !!!", p, err)
		}
		if info.IsDir() {
			addRes("dir", p)
			return nil
		}
		addRes("file", p)
		f, err := ns.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		addRes("data", string(data))
		return nil
	})

	if err != nil {
		t.Logf("WALKLOG: \n%s ", getPrintString())
		t.Fatal(err)
	}
	{
		s := getAssertString()
		if s != expected {
			fmt.Printf("\n===========\n\nEXPECTED:\n\n%s\n\nGOT:\n\n%s\n", expected, s)
			t.Fatal("not equal")
		}
	}

}
