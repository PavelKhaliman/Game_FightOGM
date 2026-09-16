package main

import (
	"fightogm/internal/app"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	root := flag.String("assets-root", "", "Папка игры с assets/sprites")
	smoke := flag.Bool("smoke-test", false, "Проверить графику и анимации в скрытом окне и выйти")
	output := flag.String("output", "test-output", "Папка отчёта smoke-test")
	flag.Parse()
	if assetRoot, err := app.FindRoot(*root); err == nil {
		logDir := filepath.Join(assetRoot, "logs")
		if os.MkdirAll(logDir, 0755) == nil {
			if file, err := os.Create(filepath.Join(logDir, "fightogm.log")); err == nil {
				defer file.Close()
				// GUI executables may have no stderr handle. Always write the file first.
				log.SetOutput(io.MultiWriter(file, os.Stderr))
			}
		}
	}
	if err := app.Run(app.Options{Root: *root, Smoke: *smoke, Output: *output}); err != nil {
		log.Printf("[FightOGM] ОШИБКА: %v", err)
		os.Exit(1)
	}
}
