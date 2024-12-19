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
    "strings"
)

// unzip extrai o conteúdo de um arquivo ZIP para um diretório de destino.
func unzip(src, dest string) error {
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

// convertMarkdownToPDF converte um arquivo Markdown em PDF usando Pandoc.
func convertMarkdownToPDF(mdFilePath, pdfDir string) error {
    pdfFilePath := filepath.Join(pdfDir, strings.TrimSuffix(filepath.Base(mdFilePath), ".md")+".pdf")
    mdPath := filepath.Dir(mdFilePath)
    bgPath := "./bg"
    resourcesPath := bgPath + ":" + mdPath
    // cmd := exec.Command("pandoc '" + mdFilePath + "' -o '" + pdfFilePath + "' --template 'eisvogel' --listings --file-scope --resource-path '" + mdPath + "'")
    cmd := exec.Command("pandoc", mdFilePath, "-o", pdfFilePath, "--template", "eisvogel", "--listings", "--file-scope", "--resource-path", resourcesPath)
    
    if err := cmd.Run(); err != nil {
        log.Println(cmd)
        return fmt.Errorf("erro ao converter %s para PDF: %v", mdFilePath, err)
    }
    log.Printf("Arquivo convertido para PDF: %s\n", pdfFilePath)
    return nil
}

// zipDirectory comprime um diretório em um arquivo ZIP.
func zipDirectory(src, dest string) error {
    zipFile, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer zipFile.Close()

    writer := zip.NewWriter(zipFile)
    defer writer.Close()

    return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if fi.Mode().IsRegular() {
            f, err := writer.Create(strings.TrimPrefix(file, src+"/"))
            if err != nil {
                return err
            }
            srcFile, err := os.Open(file)
            if err != nil {
                return err
            }
            defer srcFile.Close()
            if _, err = io.Copy(f, srcFile); err != nil {
                return err
            }
        }
        return nil
    })
}

// uploadHandler lida com o upload do arquivo e processa a conversão.
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

    dest := "../data/unzipped"
    if err := os.MkdirAll(dest, os.ModePerm); err != nil {
        http.Error(w, "Erro ao criar diretório de destino", http.StatusInternalServerError)
        return
    }

    if err := unzip(tmpFile.Name(), dest); err != nil {
        http.Error(w, "Erro ao descomprimir o arquivo", http.StatusInternalServerError)
        return
    }

    pdfDir := "../data/pdf"
    if err := os.MkdirAll(pdfDir, os.ModePerm); err != nil {
        http.Error(w, "Erro ao criar diretório PDF", http.StatusInternalServerError)
        return
    }

    if err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".md" {
    
            if err := convertMarkdownToPDF(path, pdfDir); err != nil {
                return err
            }
        }
        return nil
    }); err != nil {
        http.Error(w, "Erro ao converter arquivos Markdown para PDF", http.StatusInternalServerError)
        return
    }

    zipFilePath := "../data/arquivoPDF.zip"
    if err := zipDirectory(pdfDir, zipFilePath); err != nil {
        http.Error(w, "Erro ao comprimir a pasta PDF", http.StatusInternalServerError)
        return
    }

    if err := os.RemoveAll(dest); err != nil {
        http.Error(w, "Erro ao remover a pasta descomprimida", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(zipFilePath)))
    http.ServeFile(w, r, zipFilePath)
}

func main() {
    if _, err := exec.LookPath("pandoc"); err != nil {
        log.Fatal("Pandoc não está instalado. Por favor, instale o Pandoc para usar este serviço.")
    }

    http.HandleFunc("/upload", uploadHandler)
    log.Println("Servidor rodando na porta 5000...")
    if err := http.ListenAndServe(":5000", nil); err != nil {
        log.Fatal(err)
    }
}