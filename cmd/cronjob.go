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
	"golang.org/x/mod/module"
)

// cronjobCmd represents the cronjob command
var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "add cronjob",
	RunE:  cronjobRun,
}

var cronjobName *string

var cronjobProjectDir *string

func init() {
	rootCmd.AddCommand(cronjobCmd)
	cronjobName = cronjobCmd.Flags().StringP("name", "n", "", "cron job name")
	_ = cronjobCmd.MarkFlagRequired("name")
	cronjobProjectDir = cronjobCmd.Flags().StringP("dir", "d", "", "project directory, default is current directory")
}

func cronjobRun(_ *cobra.Command, _ []string) error {
	srcMod := "github.com/soyacen/grocer/internal/layout"
	srcModVers := srcMod + "@latest"
	srcMod, _, _ = strings.Cut(srcMod, "@")
	if err := module.CheckPath(srcMod); err != nil {
		return errors.Wrap(err, "invalid source module name")
	}

	if *cronjobProjectDir == "" {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return errors.Wrap(err, "failed to get absolute path for target directory")
		}
		*cronjobProjectDir = absDir
	} else {
		absDir, err := filepath.Abs(*cronjobProjectDir)
		if err != nil {
			return errors.Wrap(err, "failed to get absolute path for target directory")
		}
		*cronjobProjectDir = absDir
	}

	// Dir must exist and must be non-empty.
	de, err := os.ReadDir(*cronjobProjectDir)
	if err != nil || len(de) == 0 {
		return errors.New("target directory does not exist or is empty")
	}

	// Remove the needMkdir variable and mkdir logic since we're not creating directories
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
			"cmd/cronjob",
			"deploy/cronjob",
			"internal/cronjob",
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
		case "cmd/cronjob.go":
			data = fixCmdCronjobGo(data, *initProjectDir)
		default:

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

func fixCmdCronjobGo(data []byte, dir string) []byte {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "cmd/cronjob.go", data, parser.ParseComments)
	if err != nil {
		log.Fatalf("parsing source module:\n%s", err)
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
			newVal := strings.Replace(oldVal, "cronjob", path.Base(dir), -1)
			if newVal != oldVal {
				buf.Replace(fset.Position(kv.Value.Pos()).Offset, fset.Position(kv.Value.End()).Offset,
					strconv.Quote(newVal))
			}
		}
		return true
	})

	return buf.Bytes()
}
