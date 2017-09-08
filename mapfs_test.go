package vfs

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

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

func TestSafeMapFS(t *testing.T) {
	assertIsSafe(t, SafeMap(map[string]string{
		"1/2/3/4/5/6":    "test-fixtures/C/animals/cats/cats",
		"1/2/3/A/4/5/6":  "test-fixtures/C/animals/cats/cats",
		"/1/2/3/B/4/5/6": "test-fixtures/C/animals/cats/C-cats",
		"2":              "test-fixtures/C/animals/cats/cats",
	}))

	{
		ns := NewNameSpace()
		ns.Bind("/", Map(map[string]string{
			"1/2/3":   "test-fixtures/C/animals/cats/cats",
			"1/A/2/3": "test-fixtures/C/animals/cats/cats",
			"1/B/2/4": "test-fixtures/C/animals/cats/C-cats",
			"2":       "test-fixtures/C/animals/cats/cats",
			"/3":      "test-fixtures/C/animals/cats/cats",
		}), "/", BindReplace)
		// only SafeBind fixes bad input
		assertIsNotExist(t, ns,
			"/3",
		)
	}
	{
		ns := NewNameSpace()
		bindOrDie(t, ns.BindSafe("/", SafeMap(map[string]string{
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
