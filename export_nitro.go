package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	destDir := "nitro-dualite-export"

	// LISTA DE MÓDULOS PARA IGNORAR (Adicione ou remova nomes aqui)
	// Estes nomes devem bater exatamente com as pastas em src/components/
	ignoreModules := []string{
		"mod-tools",
		"wired",
		"camera",
		"groups",
		"achievements",
		"inventory",
		"navigator",
		"avatar-editor",
		"chat-history",
		"friends",
		"user-profile",
		"room-tools",
	}

	// 1. Limpeza do diretório de destino
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		fmt.Printf("Removendo diretório existente: %s\n", destDir)
		err := os.RemoveAll(destDir)
		if err != nil {
			fmt.Printf("Erro ao remover: %v\n", err)
			return
		}
	}

	// Pré-processamento dos caminhos ignorados para facilitar a comparação
	ignorePaths := make(map[string]bool)
	for _, mod := range ignoreModules {
		// Caminho relativo padrão: src/components/nome-do-modulo
		path := filepath.Join("src", "components", mod)
		ignorePaths[path] = true
	}

	// 2. Iterar sobre todos os arquivos
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(".", path)
		if err != nil {
			return err
		}

		// --- FILTROS DE DIRETÓRIOS PESADOS ---
		if d.IsDir() {
			name := d.Name()
			// Pular infraestrutura e pastas de build
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" || name == destDir {
				return filepath.SkipDir
			}

			// Pular módulos específicos de componentes solicitados
			if ignorePaths[rel] {
				fmt.Printf("Pulpando módulo pesado: %s\n", rel)
				return filepath.SkipDir
			}
		}

		// --- FILTROS DE ARQUIVOS (LOCKS/LOGS) ---
		if !d.IsDir() {
			name := d.Name()
			if name == "yarn.lock" || name == "package-lock.json" || strings.HasSuffix(name, ".log") {
				return nil
			}
		}

		// --- LÓGICA DE CÓPIA ---
		// Copiaremos tudo o que sobrar (re-permitindo styles e assets)
		// Mas apenas se estiver em src/, public/ ou for um arquivo-chave da raiz

		shouldCopy := false

		// 1. Tudo em src/ (já filtrado pelos ignoreModules acima)
		if strings.HasPrefix(rel, "src"+string(os.PathSeparator)) || rel == "src" {
			shouldCopy = true
		}

		// 2. Tudo em public/
		if strings.HasPrefix(rel, "public"+string(os.PathSeparator)) || rel == "public" {
			shouldCopy = true
		}

		// 3. Arquivos essenciais da raiz
		if !d.IsDir() && rel == d.Name() { // Arquivo está na raiz
			switch d.Name() {
			case "package.json", "tsconfig.json", "vite.config.ts", "vite.config.js", "craco.config.js", ".eslintrc.js", ".eslintrc.json", "index.html":
				shouldCopy = true
			}
		}

		if shouldCopy {
			targetPath := filepath.Join(destDir, rel)
			if d.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			}

			// Garantir diretório pai
			parent := filepath.Dir(targetPath)
			if err := os.MkdirAll(parent, 0755); err != nil {
				return err
			}

			return copyFile(path, targetPath)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nProcesso finalizado com sucesso!")
	fmt.Printf("Os arquivos foram exportados para: %s\n", destDir)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}
