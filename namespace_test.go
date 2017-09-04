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

func TestFeature(t *testing.T) {
	// generateTestFixture()
	testPath := func(path string) string {
		return filepath.Join("test-fixtures", path)
	}

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

	for _, fn := range []string{"/dogs/subdogs/B-dogs"} {
		fi, err := ns.Stat(fn)
		if err != nil {
			t.Fatal(err)
		}
		if !fi.Mode().IsRegular() {
			t.Fatal(fi.Mode().String())
		}
	}

	for _, fn := range []string{
		"/dogs/subdogs/B-dogs/nodog",
		"/dogs/subdogsnodog",
	} {
		fi, err := ns.Stat(fn)
		if err == nil {
			t.Fatal(fi.Mode().String())
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	// fmt.Println(strings.Join(results, "\n"))
	if strings.Join(results, "\n") != `dir : /
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
data: C/animals/cats/cats` {
		t.Fatal("not equal")
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
