package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/soyacen/grocer/internal/edit"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init project",
	RunE:  initRun,
}

var dstMod *string

var initProjectDir *string

func init() {
	rootCmd.AddCommand(initCmd)
	dstMod = initCmd.Flags().StringP("mod", "m", "", "go module name")
	_ = initCmd.MarkFlagRequired("mod")
	initProjectDir = initCmd.Flags().StringP("dir", "d", "", "project directory, default is current directory")
}

func initRun(_ *cobra.Command, _ []string) error {
	srcMod := "github.com/soyacen/grocer/internal/layout"
	srcModVers := srcMod + "@latest"
	srcMod, _, _ = strings.Cut(srcMod, "@")
	if err := module.CheckPath(srcMod); err != nil {
		return errors.Wrap(err, "invalid source module name")
	}

	if err := module.CheckPath(*dstMod); err != nil {
		return errors.Wrap(err, "invalid destination module name")
	}

	if *initProjectDir == "" {
		absDir, err := filepath.Abs("." + string(filepath.Separator) + path.Base(*dstMod))
		if err != nil {
			return errors.Wrap(err, "failed to get absolute path for target directory")
		}
		*initProjectDir = absDir
	} else {
		absDir, err := filepath.Abs(*initProjectDir)
		if err != nil {
			return errors.Wrap(err, "failed to get absolute path for target directory")
		}
		*initProjectDir = absDir
	}

	fmt.Println("dstMod: ", *dstMod)

	fmt.Println("dir: ", *initProjectDir)

	// Dir must not exist or must be an empty directory.
	de, err := os.ReadDir(*initProjectDir)
	if err == nil && len(de) > 0 {
		return errors.Wrap(err, "target directory exists and is non-empty")
	}
	needMkdir := err != nil

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "download", "-json", srcModVers)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Errorf("go mod download -json %s: %v\n%s%s", srcModVers, err, stderr.Bytes(), stdout.Bytes())
	}

	var info struct {
		Dir string
	}
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return errors.Errorf("go mod download -json %s: invalid JSON output: %v\n%s%s", srcMod, err, stderr.Bytes(), stdout.Bytes())
	}

	if needMkdir {
		if err := os.MkdirAll(*initProjectDir, 0o777); err != nil {
			return errors.Wrap(err, "failed to mkdir "+*initProjectDir)
		}
	}

	// Copy from module cache into new directory, making edits as needed.
	if err := filepath.WalkDir(info.Dir, func(src string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		rel, err := filepath.Rel(info.Dir, src)
		if err != nil {
			return errors.WithStack(err)
		}
		fmt.Println("rel---->", rel)
		prefixs := []string{
			"api/grpc", "api/http",
			"cmd/grpc", "cmd/http", "cmd/job", "cmd/cronjob",
			"deploy/grpc", "deploy/http", "deploy/job", "deploy/cronjob",
			"internal/grpc", "internal/http", "internal/job", "internal/cronjob",
		}
		for _, prefix := range prefixs {
			if strings.HasPrefix(rel, prefix) {
				return nil
			}
		}

		dst := filepath.Join(*initProjectDir, rel)
		if d.IsDir() {
			if err := os.MkdirAll(dst, 0o777); err != nil {
				return errors.WithStack(err)
			}
			return nil
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return errors.WithStack(err)
		}

		switch rel {
		case "cmd/root.go":
			data = fixCmdRootGo(data, *initProjectDir)
		case "go.mod":
			data = fixGoMod(data, *dstMod)
		case "Makefile":
			data = fixMakefile(data, *initProjectDir)
		}
		isRoot := !strings.Contains(rel, string(filepath.Separator))
		if strings.HasSuffix(rel, ".go") {
			data = fixGo(data, rel, srcMod, *dstMod, isRoot)
		}
		if err := os.WriteFile(dst, data, 0o666); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}); err != nil {
		return err
	}

	log.Printf("initialized %s in %s", *dstMod, *initProjectDir)
	return nil
}

func fixMakefile(data []byte, dir string) []byte {
	return bytes.ReplaceAll(data, []byte("grocer"), []byte(path.Base(dir)))
}

func fixCmdRootGo(data []byte, dir string) []byte {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "cmd/root.go", data, parser.ParseComments)
	if err != nil {
		log.Fatalf("parsing source module:\n%s", err)
	}

	buf := edit.NewBuffer(data)

	// 遍历 AST 查找 rootCmd 变量的 Use 字段
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			// 查找 cobra.Command 的结构体字面量
			if sel, ok := x.Type.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "cobra" && sel.Sel.Name == "Command" {
					// 遍历结构体字段
					for _, elt := range x.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Use" {
								// 修改 Use 字段的值
								if val, ok := kv.Value.(*ast.BasicLit); ok && val.Kind == token.STRING {
									oldVal, _ := strconv.Unquote(val.Value)
									newVal := strings.Replace(oldVal, "grocer", path.Base(dir), -1)
									if newVal != oldVal {
										buf.Replace(fset.Position(kv.Value.Pos()).Offset, fset.Position(kv.Value.End()).Offset,
											strconv.Quote(newVal))
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	return buf.Bytes()
}

// fixGo rewrites the Go source in data to replace srcMod with dstMod.
// isRoot indicates whether the file is in the root directory of the module,
// in which case we also update the package name.
func fixGo(data []byte, file string, srcMod, dstMod string, isRoot bool) []byte {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, data, parser.ImportsOnly)
	if err != nil {
		log.Fatalf("parsing source module:\n%s", err)
	}

	buf := edit.NewBuffer(data)
	at := func(p token.Pos) int {
		return fset.File(p).Offset(p)
	}

	srcName := path.Base(srcMod)
	dstName := path.Base(dstMod)
	if isRoot {
		if name := f.Name.Name; name == srcName || name == srcName+"_test" {
			dname := dstName + strings.TrimPrefix(name, srcName)
			if !token.IsIdentifier(dname) {
				log.Fatalf("%s: cannot rename package %s to package %s: invalid package name", file, name, dname)
			}
			buf.Replace(at(f.Name.Pos()), at(f.Name.End()), dname)
		}
	}

	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if path == srcMod {
			if srcName != dstName && spec.Name == nil {
				// Add package rename because source code uses original name.
				// The renaming looks strange, but template authors are unlikely to
				// create a template where the root package is imported by packages
				// in subdirectories, and the renaming at least keeps the code working.
				// A more sophisticated approach would be to rename the uses of
				// the package identifier in the file too, but then you have to worry about
				// name collisions, and given how unlikely this is, it doesn't seem worth
				// trying to clean up the file that way.
				buf.Insert(at(spec.Path.Pos()), srcName+" ")
			}
			// Change import path to dstMod
			buf.Replace(at(spec.Path.Pos()), at(spec.Path.End()), strconv.Quote(dstMod))
		}
		if strings.HasPrefix(path, srcMod+"/") {
			// Change import path to begin with dstMod
			buf.Replace(at(spec.Path.Pos()), at(spec.Path.End()), strconv.Quote(strings.Replace(path, srcMod, dstMod, 1)))
		}
	}
	return buf.Bytes()
}

// fixGoMod rewrites the go.mod content in data to add a module
// statement for dstMod.
func fixGoMod(data []byte, dstMod string) []byte {
	f, err := modfile.ParseLax("go.mod", data, nil)
	if err != nil {
		log.Fatalf("parsing source module:\n%s", err)
	}
	_ = f.AddModuleStmt(dstMod)
	new, err := f.Format()
	if err != nil {
		return data
	}
	return new
}
