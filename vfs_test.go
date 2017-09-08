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

// this file contains helper functions for all the tests

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

func assertIsSafe(t *testing.T, ffs ...FileSystemFunc) {
	t.Helper()
	for _, ff := range ffs {
		_, err := ff()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func assertNotSafe(t *testing.T, ffs ...FileSystemFunc) {
	t.Helper()
	for _, ff := range ffs {
		_, err := ff()
		if err == nil {
			t.Fatal("not safe")
		}
	}
}

func bindOrDie(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}

}
func safeOrDie(t *testing.T, f FileSystemFunc) FileSystem {
	t.Helper()
	fs, err := f()
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func assertIsRegular(t *testing.T, ns NameSpace, paths ...string) {
	t.Helper()
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
	t.Helper()
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
	t.Helper()
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

func assertOSPather(t *testing.T, ns NameSpace, paths map[string]string) {
ps:
	for fn, ospath := range paths {
		fi, err := ns.Stat(fn)
		if err != nil {
			t.Fatalf("expected path to be readable: %s : %v, %v ", fn, fi.Mode().String(), err)
		}
		if opr, ok := fi.(OSPather); ok {
			op := opr.OSPath()
			if ospath == "" {
				t.Fatalf("did not expect vpath '%s' to be OSPather: '%s' != '%s'", fn, ospath, op)
			}
			if op != ospath {
				t.Fatalf("expected ospath to be equal: '%s' != '%s'", ospath, op)
			}
			continue ps
		}
		if ospath != "" {
			t.Fatalf("expected path to be ospath: %s", fn, ospath)
		}
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
