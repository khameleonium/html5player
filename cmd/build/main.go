package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	
	switch command {
	case "pack":
		packCmd()
	case "unpack":
		unpackCmd()
	case "batch":
		batchCmd()
	case "build":
		fallthrough
	default:
		if command == "build" {
			os.Args = append(os.Args[:1], os.Args[2:]...)
		}
		buildCmd()
	}
}

func printUsage() {
	fmt.Println(`Утилита сборки HTML5 Game Player

Команды:
  build   - Собрать исполняемый файл игры
  pack    - Упаковать папку в зашифрованный XOR архив
  unpack  - Распаковать XOR архив в обычный ZIP
  batch   - Собрать сразу несколько игр (пакетная сборка)

Примеры:
  go run ./cmd/build build -mode embed -out game.exe
  go run ./cmd/build batch -mode embed -games "C:\games\game1,C:\games\game2"
`)
}

func buildCmd() {
	mode := flag.String("mode", "embed", "Режим сборки: embed, dir, xor")
	out := flag.String("out", "game.exe", "Имя выходного файла")
	key := flag.String("key", "", "Ключ для XOR (только для mode=xor)")
	dirName := flag.String("dirName", "game_data", "Имя папки для dir режима")
	pakName := flag.String("pakName", "game_res.pak", "Имя pak файла для xor режима")
	flag.Parse()

	if *mode == "xor" && *key == "" {
		fmt.Println("Ошибка: для режима xor необходимо указать -key")
		os.Exit(1)
	}

	fmt.Printf("Собираем игру (режим: %s)...\n", *mode)
	buildGame(*out, *mode, *key, *dirName, *pakName)
}

func buildGame(out string, mode string, key string, dirName string, pakName string) {
	ldflags := fmt.Sprintf("-X main.RunMode=%s -X main.XorKey=%s -X main.DirName=%s -X main.PakName=%s", mode, key, dirName, pakName)
	ldflags += " -H windowsgui"
	
	generateWinRes()
	
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", out, "./cmd/game")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Ошибка сборки %s: %v\n", out, err)
		os.Exit(1)
	}
	fmt.Printf("Сборка успешно завершена: %s\n", out)
}

func batchCmd() {
	batchFlags := flag.NewFlagSet("batch", flag.ExitOnError)
	mode := batchFlags.String("mode", "embed", "Режим сборки: embed, dir, xor")
	games := batchFlags.String("games", "", "Пути к папкам игр через запятую")
	key := batchFlags.String("key", "", "Ключ для XOR (если xor)")
	batchFlags.Parse(os.Args[2:])

	if *games == "" {
		fmt.Println("Ошибка: необходимо указать папки через -games")
		os.Exit(1)
	}
	
	err := os.MkdirAll("builds", 0755)
	if err != nil {
		fmt.Println("Ошибка создания папки builds:", err)
		os.Exit(1)
	}

	gamePaths := strings.Split(*games, ",")
	
	if *mode == "embed" {
		if _, err := os.Stat("game_data"); err == nil {
			os.Rename("game_data", "game_data_bak")
			defer os.Rename("game_data_bak", "game_data")
		}
	}

	for _, gPath := range gamePaths {
		gPath = strings.TrimSpace(gPath)
		if gPath == "" { continue }
		
		gameName := filepath.Base(filepath.Clean(gPath))
		fmt.Printf("\n=== Обработка игры: %s ===\n", gameName)
		
		outExe := filepath.Join("builds", gameName+".exe")
		
		if *mode == "embed" {
			os.RemoveAll("game_data")
			err := copyDir(gPath, "game_data")
			if err != nil {
				fmt.Println("Ошибка копирования:", err)
				continue
			}
			buildGame(outExe, *mode, *key, "game_data", "game_res.pak")
			os.RemoveAll("game_data")
		} else if *mode == "xor" {
			outPak := filepath.Join("builds", gameName+".pak")
			packGame(gPath, outPak, *key)
			buildGame(outExe, *mode, *key, "game_data", gameName+".pak")
		} else if *mode == "dir" {
			buildGame(outExe, *mode, *key, gameName, "game_res.pak")
		}
	}
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil { return err }
		return os.WriteFile(target, data, 0644)
	})
}

func packCmd() {
	packFlags := flag.NewFlagSet("pack", flag.ExitOnError)
	in := packFlags.String("in", "game_data", "Папка с ресурсами")
	out := packFlags.String("out", "game_res.pak", "Выходной файл")
	key := packFlags.String("key", "", "Ключ для шифрования XOR")
	packFlags.Parse(os.Args[2:])

	if *key == "" {
		fmt.Println("Ошибка: необходимо указать -key")
		os.Exit(1)
	}

	packGame(*in, *out, *key)
}

func packGame(in string, out string, key string) {
	fmt.Printf("Упаковываем папку %s в %s (XOR)...\n", in, out)

	tempZip, err := os.CreateTemp("", "game_res_*.zip")
	if err != nil {
		fmt.Println("Ошибка создания временного файла:", err)
		os.Exit(1)
	}
	defer os.Remove(tempZip.Name())

	zipWriter := zip.NewWriter(tempZip)
	err = filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }

		relPath, _ := filepath.Rel(in, path)
		relPath = filepath.ToSlash(relPath)

		f, err := os.Open(path)
		if err != nil { return err }
		defer f.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil { return err }

		_, err = io.Copy(w, f)
		return err
	})
	
	if err != nil {
		fmt.Println("Ошибка при создании ZIP:", err)
		os.Exit(1)
	}
	zipWriter.Close()
	tempZip.Close()

	data, err := os.ReadFile(tempZip.Name())
	if err != nil {
		fmt.Println("Ошибка чтения временного ZIP:", err)
		os.Exit(1)
	}

	applyXor(data, []byte(key))

	if err := os.WriteFile(out, data, 0644); err != nil {
		fmt.Println("Ошибка сохранения выходного файла:", err)
		os.Exit(1)
	}
	fmt.Println("Успешно упаковано!")
}

func unpackCmd() {
	unpackFlags := flag.NewFlagSet("unpack", flag.ExitOnError)
	in := unpackFlags.String("in", "game_res.pak", "Файл с зашифрованными ресурсами")
	out := unpackFlags.String("out", "unpacked.zip", "Выходной обычный ZIP файл")
	key := unpackFlags.String("key", "", "Ключ для расшифровки XOR")
	unpackFlags.Parse(os.Args[2:])

	if *key == "" {
		fmt.Println("Ошибка: необходимо указать -key")
		os.Exit(1)
	}

	fmt.Printf("Распаковываем %s в %s...\n", *in, *out)

	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		os.Exit(1)
	}

	applyXor(data, []byte(*key))

	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Println("Ошибка записи ZIP файла:", err)
		os.Exit(1)
	}
	fmt.Println("Успешно распаковано! Теперь вы можете открыть", *out, "любым архиватором.")
}

func applyXor(data []byte, key []byte) {
	keyLen := len(key)
	if keyLen == 0 { return }
	for i := 0; i < len(data); i++ {
		data[i] ^= key[i%keyLen]
	}
}

func generateWinRes() {
	if _, err := os.Stat("winres/winres.json"); os.IsNotExist(err) {
		return
	}
	
	fmt.Println("Генерация ресурсов Windows (иконка, мета-данные)...")
	cmd := exec.Command("go", "run", "github.com/tc-hib/go-winres@latest", "make", "--out", "cmd/game/rsrc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Внимание: Не удалось сгенерировать ресурсы (go-winres): %v\n", err)
		fmt.Println("Сборка продолжится без иконки и мета-данных.")
	}
}
