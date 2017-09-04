package vfs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const verbose = false

func TestIncludeExclude(t *testing.T) {
	t.Skip() // TODO: include/exclude has bugs
	ns := NameSpace{}
	ns.Bind("/", NewNameSpace(), "/", BindReplace)
	ns.Bind("/1", Include(OS(testPath("A")), "/animals/cats/cats"), "/", BindAfter)
	// ns.Bind("/", OS(testPath("B")), "/", BindAfter)
	// ns.Bind("/", OS(testPath("C")), "/", BindAfter)
	expected := "NOT THIS"
	assertWalk(t, ns, expected)
	assertIsDir(t, ns,
		"/",
	)
	assertIsRegular(t, ns)
	assertIsNotExist(t, ns,
		"/1/2/3/4/5/7",
	)
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

func TestComplicated(t *testing.T) {
	// generateTestFixture()
	ns := NameSpace{}
	// ns.Bind("/", OS(testPath(".")), "/", BindReplace)
	// ns.Bind("/", OS("."), "/", BindReplace)
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

	expected := `dir : /
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
data: C/animals/cats/cats`

	assertWalk(t, ns, expected)

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
	}
}

func assertIsDir(t *testing.T, ns NameSpace, paths ...string) {
	for _, fn := range paths {
		fi, err := ns.Stat(fn)
		if err != nil {
			t.Fatal(err)
		}
		if !fi.Mode().IsDir() {
			t.Fatal(fi.Mode().String())
		}
		if _, err := ns.ReadDir(fn); err != nil {
			t.Fatal(err)
		}
	}
}

func assertIsNotExist(t *testing.T, ns NameSpace, paths ...string) {
	for _, fn := range paths {
		fi, err := ns.Stat(fn)
		if err == nil {
			t.Fatal(fi.Mode().String())
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}

}

func assertWalk(t *testing.T, ns NameSpace, expected string) {
	var results []string
	addRes := func(kind, data string) {
		s := fmt.Sprintf("%-4s: %s", kind, data)
		results = append(results, s)

		if verbose {
			switch kind {
			case "dir", "file":
				fmt.Print("\n")
			}
			fmt.Printf("%-6s: %-30s", kind, data)
		}

	}
	err := Walk("/", ns, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("ERROR: %s : %v !!!", p, err)
			return nil
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
		t.Fatal(err)
	}

	s := strings.Join(results, "\n")
	if s != expected {
		fmt.Printf("\n===========\n\nEXPECTED:\n\n%s\n\nGOT:\n\n%s\n", expected, s)
		t.Fatal("not equal")
	}

}
