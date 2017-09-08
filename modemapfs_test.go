package vfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestModeMap(t *testing.T) {
	ns := NewNameSpace()

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
			// fmt.Println("dir", p, info.Mode())
			return nil
		}
		// fmt.Println("file", p)
		f, err := ns.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		_ = data
		// fmt.Println("data", string(data))
		return nil
	})

}
