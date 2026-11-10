package volumes

import (
	"runtime"
	"strings"
	"testing"
)

func TestListHasSystemVolume(t *testing.T) {
	vs := List()
	if len(vs) == 0 {
		t.Fatal("no volumes found")
	}
	for _, v := range vs {
		if v.Path == "" || v.Name == "" {
			t.Errorf("incomplete volume: %+v", v)
		}
	}
	if runtime.GOOS != "windows" && vs[0].Path != "/" {
		t.Errorf("system volume should be first, got %+v", vs[0])
	}
	if runtime.GOOS != "windows" && (vs[0].Total <= 0 || vs[0].Free > vs[0].Total) {
		t.Errorf("bad capacity: %+v", vs[0])
	}
}

func TestParseMounts(t *testing.T) {
	data := `sysfs /sys sysfs rw 0 0
/dev/nvme0n1p2 / ext4 rw 0 0
/dev/loop3 /snap/core/1 squashfs ro 0 0
tmpfs /run tmpfs rw 0 0
/dev/sdb1 /media/user/My\040Disk vfat rw 0 0
server:/share /mnt/nas nfs4 rw 0 0
/dev/nvme0n1p2 / ext4 rw 0 0
`
	vs := parseMounts(strings.NewReader(data))
	want := []Volume{
		{Name: "Root", Path: "/"},
		{Name: "My Disk", Path: "/media/user/My Disk"},
		{Name: "nas", Path: "/mnt/nas"},
	}
	if len(vs) != len(want) {
		t.Fatalf("got %+v, want %+v", vs, want)
	}
	for i := range want {
		if vs[i] != want[i] {
			t.Errorf("volume %d = %+v, want %+v", i, vs[i], want[i])
		}
	}
}
