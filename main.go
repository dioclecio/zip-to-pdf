package main

import (
    "archive/zip"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
)

// Função para descomprimir o arquivo ZIP
func unzip(src string, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        fpath := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
                return err
            }
            continue
        }

        if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        outFile, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }
        defer outFile.Close()

        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer rc.Close()

        if _, err = io.Copy(outFile, rc); err != nil {
            return err
        }
    }
    return nil
}

// Função para converter Markdown em PDF usando Pandoc
func convertMarkdownToPDF(mdFilePath string, pdfDir string) error {
    pdfFilePath := filepath.Join(pdfDir, filepath.Base(mdFilePath[:len(mdFilePath)-2]+"pdf"))
	fmt.Printf("Arquivo: %s\n", pdfFilePath)
    cmd := exec.Command("pandoc", mdFilePath, "-o", pdfFilePath)
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("erro ao converter %s para PDF: %v", mdFilePath, err)
    }
    fmt.Printf("Arquivo convertido para PDF: %s\n", pdfFilePath)
    return nil
}

// Handler para o upload do arquivo
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
        return
    }

    r.ParseMultipartForm(10 << 20) // 10 MB

    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Erro ao obter o arquivo", http.StatusBadRequest)
        return
    }
    defer file.Close()

    tmpFile, err := os.CreateTemp("", "*.zip")
    if err != nil {
        http.Error(w, "Erro ao criar arquivo temporário", http.StatusInternalServerError)
        return
    }
    defer os.Remove(tmpFile.Name())

    if _, err := io.Copy(tmpFile, file); err != nil {
        http.Error(w, "Erro ao salvar o arquivo", http.StatusInternalServerError)
        return
    }

    dest := "./unzipped"
    if err := os.MkdirAll(dest, os.ModePerm); err != nil {
        http.Error(w, "Erro ao criar diretório de destino", http.StatusInternalServerError)
        return
    }

    if err := unzip(tmpFile.Name(), dest); err != nil {
        http.Error(w, "Erro ao descomprimir o arquivo", http.StatusInternalServerError)
        return
    }

    pdfDir := filepath.Join(dest, "PDF")
    if err := os.MkdirAll(pdfDir, os.ModePerm); err != nil {
        http.Error(w, "Erro ao criar diretório PDF", http.StatusInternalServerError)
        return
    }

    err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".md" {
            if err := convertMarkdownToPDF(path, pdfDir); err != nil {
                return err
            }
        }
        return nil
    })

    if err != nil {
        http.Error(w, "Erro ao converter arquivos Markdown para PDF", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Arquivo descomprimido e arquivos Markdown convertidos com sucesso em %s", dest)
}

func main() {
    // Verifica se o Pandoc está instalado
    if _, err := exec.LookPath("pandoc"); err != nil {
        log.Fatal("Pandoc não está instalado. Por favor, instale o Pandoc para usar este serviço.")
    }

    http.HandleFunc("/upload", uploadHandler)
    fmt.Println("Servidor rodando na porta 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}