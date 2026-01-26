package cmd

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
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

type initFlags struct {
	Mod string
	Dir string
}

var initFlag initFlags

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&initFlag.Mod, "mod", "m", "", "go module name")
	_ = initCmd.MarkFlagRequired("mod")
	initCmd.Flags().StringVarP(&initFlag.Dir, "dir", "d", "", "project directory, default is current directory")
}

func initRun(_ *cobra.Command, _ []string) error {
	srcMod, srcModVers, err := getSrcModInfo()
	if err != nil {
		return err
	}
	info, err := getGoModInfo(srcMod, srcModVers)
	if err != nil {
		return err
	}

	if err := module.CheckPath(initFlag.Mod); err != nil {
		return errors.Wrap(err, "invalid destination module name")
	}

	dir, err := getProjectDir(initFlag.Dir, initFlag.Mod)
	if err != nil {
		return err
	}

	// Dir must not exist or must be an empty directory.
	de, err := os.ReadDir(dir)
	if err == nil && len(de) > 0 {
		return errors.New("target directory exists and is non-empty")
	}
	if err != nil {
		// need make directory
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return errors.Wrap(err, "failed to mkdir "+dir)
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

		dst := filepath.Join(dir, rel)
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
			data, err = fixCmdRootGo(data, dir)
		case "go.mod":
			data, err = fixGoMod(data, initFlag.Mod)
		case "Makefile":
			data = fixMakefile(data, dir)
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".go") {
			isRoot := !strings.Contains(rel, string(filepath.Separator))
			data, err = fixGo(data, rel, srcMod, initFlag.Mod, isRoot)
			if err != nil {
				return err
			}
		}
		if err := os.WriteFile(dst, data, 0o666); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}); err != nil {
		return err
	}

	log.Printf("initialized %s in %s", initFlag.Mod, dir)
	return nil
}

func fixMakefile(data []byte, dir string) []byte {
	return bytes.ReplaceAll(data, []byte("grocer"), []byte(path.Base(dir)))
}

func fixCmdRootGo(data []byte, dir string) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "cmd/root.go", data, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse source module")
	}

	buf := edit.NewBuffer(data)

	// 遍历 AST 查找 rootCmd 变量的 Use 字段
	ast.Inspect(f, func(n ast.Node) bool {
		x, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		sel, ok := x.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "cobra" || sel.Sel.Name != "Command" {
			return true
		}

		// 遍历结构体字段
		for _, elt := range x.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Use" {
				continue
			}

			val, ok := kv.Value.(*ast.BasicLit)
			if !ok || val.Kind != token.STRING {
				continue
			}

			oldVal, _ := strconv.Unquote(val.Value)
			newVal := strings.Replace(oldVal, "grocer", path.Base(dir), -1)
			if newVal != oldVal {
				buf.Replace(fset.Position(kv.Value.Pos()).Offset, fset.Position(kv.Value.End()).Offset,
					strconv.Quote(newVal))
			}
		}
		return true
	})

	return buf.Bytes(), nil
}

func fixGo(data []byte, file string, srcMod, dstMod string, isRoot bool) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, data, parser.ImportsOnly)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", file)
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
				return nil, errors.Errorf("%s: cannot rename package %s to package %s: invalid package name", file, name, dname)
			}
			buf.Replace(at(f.Name.Pos()), at(f.Name.End()), dname)
		}
	}

	for _, spec := range f.Imports {
		pathStr, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		if pathStr != srcMod && !strings.HasPrefix(pathStr, srcMod+"/") {
			continue
		}

		if pathStr == srcMod {
			if srcName != dstName && spec.Name == nil {
				buf.Insert(at(spec.Path.Pos()), srcName+" ")
			}
			buf.Replace(at(spec.Path.Pos()), at(spec.Path.End()), strconv.Quote(dstMod))
		} else if strings.HasPrefix(pathStr, srcMod+"/") {
			buf.Replace(at(spec.Path.Pos()), at(spec.Path.End()), strconv.Quote(strings.Replace(pathStr, srcMod, dstMod, 1)))
		}
	}
	return buf.Bytes(), nil
}

func fixGoMod(data []byte, dstMod string) ([]byte, error) {
	filename := "go.mod"
	f, err := modfile.ParseLax(filename, data, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", filename)
	}
	_ = f.AddModuleStmt(dstMod)
	newData, err := f.Format()
	if err != nil {
		return data, nil
	}
	return newData, nil
}
