package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "add cronjob",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := flag.IsValid(); err != nil {
			return err
		}
		return jobRun(cmd, args, "cronjob")
	},
}

func init() {
	rootCmd.AddCommand(cronjobCmd)
	cronjobCmd.Flags().StringVarP(&flag.Name, "name", "n", "", "cron job name, must consist of alphanumeric characters and underscores, and start with a letter, required")
	_ = cronjobCmd.MarkFlagRequired("name")
	cronjobCmd.Flags().StringVarP(&flag.Dir, "dir", "d", "", "project directory, default is current directory")
}

func jobRun(_ *cobra.Command, _ []string, kind string) error {
	srcMod, srcModVers, err := getSrcModInfo()
	if err != nil {
		return err
	}

	info, err := getGoModInfo(srcMod, srcModVers)
	if err != nil {
		return err
	}

	dir, err := getProjectDir(flag.Dir, "")
	if err != nil {
		return err
	}

	// Dir must exist and must be non-empty.
	de, err := os.ReadDir(dir)
	if err != nil || len(de) == 0 {
		return fmt.Errorf("target directory %s does not exist or is empty", dir)
	}

	files := []string{
		dir + "/cmd/" + flag.Name + ".go",
		dir + "/deploy/values/" + flag.Name + ".yaml",
		dir + "/internal/" + flag.Name,
	}

	// 检查所有可能的目标路径是否已存在
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			if stat.IsDir() {
				return fmt.Errorf("target directory already exists: %s", file)
			} else {
				return fmt.Errorf("target file already exists: %s", file)
			}
		}
	}

	dstMod, err := readMod(dir)
	if err != nil {
		return err
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
			"cmd/" + kind,
			"deploy/values/" + kind,
			"internal/" + kind,
		}
		var matched bool
		for _, prefix := range prefixs {
			if strings.HasPrefix(rel, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
		dst := strings.Replace(filepath.Join(dir, rel), kind, flag.Name, 1)
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
		case "deploy/values/" + kind + ".yaml":
			data = bytes.ReplaceAll(data, []byte("grocer-"+kind), []byte(path.Base(dstMod)+"-"+flag.Name))
			dst = strings.TrimSuffix(dst, flag.Name+".yaml")
			for _, env := range envs {
				dir := filepath.Join(dst, env)
				if err := os.MkdirAll(dir, 0o777); err != nil {
					return errors.WithStack(err)
				}
				file := filepath.Join(dir, flag.Name+".yaml")
				if err := os.WriteFile(file, data, 0o666); err != nil {
					return errors.WithStack(err)
				}
			}
			_ = data
			return nil
		case "cmd/" + kind + ".go",
			"internal/" + kind + "/fx.go",
			"internal/" + kind + "/model.go",
			"internal/" + kind + "/repo.go",
			"internal/" + kind + "/repository.go",
			"internal/" + kind + "/service.go":
			data = bytes.ReplaceAll(data, []byte(kind), []byte(flag.Name))
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".go") {
			isRoot := !strings.Contains(rel, string(filepath.Separator))
			data, err = fixGo(data, rel, srcMod, dstMod, isRoot)
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

	log.Printf("add %s %s in %s\n", kind, flag.Name, dir)
	return nil
}
