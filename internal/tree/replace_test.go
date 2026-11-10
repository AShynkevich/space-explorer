package tree

import (
	"path/filepath"
	"testing"
)

// build: /r (300) -> big (200: deep 150 + other 50), small (100)
func sample() (r, big, deep, small *Node) {
	r = &Node{Name: filepath.FromSlash("/r"), IsDir: true, Size: 300, Files: 6, Dirs: 3}
	big = &Node{Name: "big", Parent: r, IsDir: true, Size: 200, Files: 4, Dirs: 1, Other: Other{Items: 1, Size: 50}}
	deep = &Node{Name: "deep", Parent: big, IsDir: true, Size: 150, Files: 3}
	small = &Node{Name: "small", Parent: r, IsDir: true, Size: 100, Files: 2}
	big.Children = []*Node{deep}
	r.Children = []*Node{big, small}
	return
}

func TestReplaceUpdatesAncestors(t *testing.T) {
	r, big, deep, small := sample()
	// deep grew from 150 to 450 and now holds one more file and one folder
	fresh := &Node{Name: "tmp-root", IsDir: true, Size: 450, Files: 5, Dirs: 1}
	child := &Node{Name: "x", Parent: fresh, Size: 400, Files: 1}
	fresh.Children = []*Node{child}
	Replace(deep, fresh)

	if fresh.Name != "deep" || fresh.Parent != big {
		t.Errorf("fresh not adopted: name %q parent %v", fresh.Name, fresh.Parent)
	}
	if big.Size != 500 || r.Size != 600 {
		t.Errorf("sizes: big %d (want 500), root %d (want 600)", big.Size, r.Size)
	}
	if big.Files != 6 || r.Files != 8 || big.Dirs != 2 || r.Dirs != 4 {
		t.Errorf("counts: big %d/%d root %d/%d", big.Files, big.Dirs, r.Files, r.Dirs)
	}
	if r.Children[0] != big || r.Children[1] != small {
		t.Error("root children order wrong")
	}
	if got, want := child.Path(), filepath.FromSlash("/r/big/deep/x"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestReplaceResortsAndPropagatesUnreadable(t *testing.T) {
	r, big, _, small := sample()
	// small becomes larger than big
	fresh := &Node{IsDir: true, Size: 900, Files: 9, Unreadable: true}
	Replace(small, fresh)
	if r.Children[0] != fresh || r.Children[1] != big {
		t.Error("children not re-sorted after replace")
	}
	if !r.Unreadable {
		t.Error("unreadable flag should reach the root")
	}
	if r.Size != 1100 {
		t.Errorf("root size = %d, want 1100", r.Size)
	}
}

func TestReplaceRootIsNoop(t *testing.T) {
	r, _, _, _ := sample()
	Replace(r, &Node{Size: 1})
	if r.Size != 300 {
		t.Error("replacing the root must do nothing")
	}
}
