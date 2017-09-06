package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/karrick/godirwalk"
	"github.com/pivotal-golang/bytefmt"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Node struct {
	name     string
	size     int64
	children nodemap
}

func NewNode(name string) *Node {
	return &Node{
		name:     name,
		children: nodemap{},
	}
}

type nodes []*Node

func (s nodes) Len() int      { return len(s) }
func (s nodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type BySize struct{ nodes }

func (s BySize) Less(i, j int) bool { return s.nodes[i].size < s.nodes[j].size }

type nodemap map[string]*Node

func (mp nodemap) slice() nodes {
	nodeslice := nodes{}
	for _, n := range mp {
		nodeslice = append(nodeslice, n)
	}

	return nodeslice
}

var rootnode *Node
var lines int
var files int

func strhashsForPct(a, b int64) string {
	t := math.Ceil((float64(a) / float64(b)) * 20)
	return strings.Repeat("#", int(t))
}

const ansiClearLine = "\033[2K"
const ansiCursorUp = "\033[%dA"

func clearOutput(l int) {
	fmt.Printf(ansiCursorUp, lines)
	lines = 0
}

func printNode(node *Node, tot int64) {
	fmt.Printf("%s%21s %8s   %s\n", ansiClearLine, strhashsForPct(node.size, tot), bytefmt.ByteSize(uint64(node.size)), node.name)
	lines++
}

func printTotals() {
	fmt.Printf("%s%21s %8s   (%d files)\n", ansiClearLine, "TOTAL:", bytefmt.ByteSize(uint64(rootnode.size)), files)
	lines++
}

func printHistogram(node *Node) {
	nodeSlice := node.children.slice()
	sort.Sort(BySize{nodeSlice})
	largestSize := nodeSlice[len(nodeSlice)-1].size

	termHeight, _ := terminal.Height()

	startNode := len(nodeSlice) - int(termHeight) + 3
	if startNode <= 0 || termHeight <= 0 {
		startNode = 0
	}
	totalLines := len(nodeSlice) - startNode

	clearOutput(totalLines)

	for _, n := range nodeSlice[startNode:] {
		printNode(n, largestSize)
	}
	printTotals()
}

func addFile(path string) {
	fi, err := os.Lstat(path)
	if err != nil {
		return
	}
	filesize := fi.Size()

	files++

	relpath, _ := filepath.Rel(rootnode.name, path)
	pathparts := strings.Split(relpath, string(filepath.Separator))

	curnode := rootnode
	curnode.size += filesize
	for _, part := range pathparts {
		childnode, exists := curnode.children[part]
		if !exists {
			childnode = NewNode(part)
			curnode.children[part] = childnode
		}
		curnode = childnode
		curnode.size += filesize
	}
}

func isDir(dir string) bool {
	fi, _ := os.Stat(dir)
	return fi.IsDir()
}

func walkPathAndPrintResults(rootpath string) {
	doneChan := make(chan bool)

	rootnode = NewNode(rootpath)

	go func() {
		godirwalk.Walk(rootpath, &godirwalk.Options{
			Unsorted: true,
			Callback: func(path string, de *godirwalk.Dirent) error {
				if de.ModeType().IsRegular() {
					addFile(path)
				}
				return nil
			},
		})
		doneChan <- true
	}()

	ticker := time.NewTicker(time.Millisecond * 500)
	go func() {
		for _ = range ticker.C {
			printHistogram(rootnode)
		}
	}()

	for {
		select {
		case <-doneChan:
			ticker.Stop()
			printHistogram(rootnode)
			return
		}
	}
}

func parseArgsForRootPath() string {
	flag.Parse()

	rootpath := flag.Arg(0)
	if rootpath == "" {
		rootpath, _ = os.Getwd()
	}
	rootpath, _ = filepath.Abs(rootpath)
	if !isDir(rootpath) {
		fmt.Printf("%s is not a directory", rootpath)
		os.Exit(1)
	}

	return rootpath
}

func main() {
	walkPathAndPrintResults(parseArgsForRootPath())
}
