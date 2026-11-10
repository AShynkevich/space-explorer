// Command scantool is a developer CLI for the scanner package: it scans a
// folder and prints the largest entries. Used to validate the scanner against
// `du` and to measure memory use.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"space-explorer/internal/scanner"
	"space-explorer/internal/tree"
)

func human(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func print(n *tree.Node, total int64, depth, maxDepth int, top int) {
	pct := 0.0
	if total > 0 {
		pct = float64(n.Size) * 100 / float64(total)
	}
	name := n.Name
	if n.IsDir && depth > 0 {
		name += "/"
	}
	fmt.Printf("%s%-40s %10s %5.1f%%\n", strings.Repeat("  ", depth), name, human(n.Size), pct)
	if depth >= maxDepth {
		return
	}
	for i, c := range n.Children {
		if i >= top {
			break
		}
		print(c, total, depth+1, maxDepth, top)
	}
	if n.Other.Items > 0 && depth < maxDepth {
		fmt.Printf("%s%-40s %10s\n", strings.Repeat("  ", depth+1),
			fmt.Sprintf("(%d other items)", n.Other.Items), human(n.Other.Size))
	}
}

func main() {
	apparent := flag.Bool("apparent", false, "use file length instead of allocated blocks")
	depth := flag.Int("depth", 2, "tree depth to print")
	top := flag.Int("top", 8, "children printed per folder")
	workers := flag.Int("workers", 0, "worker count (0 = default)")
	oneFS := flag.Bool("x", false, "stay on one file system")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: scantool [flags] <folder>")
		os.Exit(2)
	}

	o := scanner.DefaultOptions()
	o.Apparent = *apparent
	o.OneFileSystem = *oneFS
	if *workers > 0 {
		o.Workers = *workers
	}
	s := scanner.New(flag.Arg(0), o)

	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				p := s.Progress()
				fmt.Fprintf(os.Stderr, "\r%d files, %d dirs, %s   ", p.Files, p.Dirs, human(p.Bytes))
			}
		}
	}()

	res, err := s.Run(context.Background())
	close(stop)
	fmt.Fprint(os.Stderr, "\r\033[K")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	print(res.Root, res.Root.Size, 0, *depth, *top)
	fmt.Printf("\nfiles=%d dirs=%d size=%s time=%s errors=%d\n",
		res.Root.Files, res.Root.Dirs, human(res.Root.Size), res.Duration.Round(time.Millisecond), res.ErrorCount)
	fmt.Printf("retained nodes=%d heap=%s sys=%s\n", res.Root.Count(), human(int64(ms.HeapAlloc)), human(int64(ms.Sys)))
}
