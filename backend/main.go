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
    "runtime"
    "strings"
    "time"
    // "github.com/pdfcpu/pdfcpu/pkg/api"
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

func combinePDFs(inputPDF, outputPDF string) error {
    // Gerar nomes temporários no formato DOS (8.3)
    tempInput := filepath.Join(filepath.Dir(inputPDF), "tempin.pdf")
    tempOutput := filepath.Join(filepath.Dir(outputPDF), "tempout.pdf")

    // Renomear o arquivo de entrada para o nome temporário
    if err := os.Rename(inputPDF, tempInput); err != nil {
        return fmt.Errorf("Erro ao renomear arquivo de entrada: %v", err)
    }

    // Comando para combinar a capa com o arquivo PDF usando pdftk
    cmd := exec.Command("pdftk", "./bg/capa.pdf", tempInput, "cat", "output", tempOutput)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Erro ao combinar PDFs: %v\n", err)
        log.Printf("Saída do comando: %s\n", string(output))
        return fmt.Errorf("Erro ao combinar %s com capa: %v", inputPDF, err)
    }

    // Renomear o arquivo de saída para o nome original
    if err := os.Rename(tempOutput, outputPDF); err != nil {
        return fmt.Errorf("Erro ao renomear arquivo de saída: %v", err)
    }

    // Remover o arquivo temporário de entrada
    if err := os.Remove(tempInput); err != nil {
        log.Printf("Erro ao remover arquivo temporário de entrada: %v\n", err)
    }

    return nil
}

// writeErrorLog writes conversion errors to a log file
func writeErrorLog(pdfDir, mdFile string, err error, cmdOutput string) error {
    logFile := filepath.Join(pdfDir, "conversion_errors.log")
    f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    timestamp := time.Now().Format("2006-01-02 15:04:05")
    logEntry := fmt.Sprintf("[%s] Erro ao converter arquivo: %s\nErro: %v\nSaída do Pandoc:\n%s\n-------------------\n", 
        timestamp, mdFile, err, cmdOutput)
    if _, err := f.WriteString(logEntry); err != nil {
        return err
    }
    return nil
}

// convertMarkdownToPDF converte um arquivo Markdown em PDF usando Pandoc.
func convertMarkdownToPDF(mdFilePath, pdfDir string) error {
    baseName := strings.TrimSuffix(filepath.Base(mdFilePath), ".md")
    pdfFilePath := filepath.Join(pdfDir, baseName+".pdf")
    mdPath := filepath.Dir(mdFilePath)
    bgPath := "./bg"
    resourcesPath := bgPath + ":" + mdPath
    headerPath := bgPath + "/header.md"
    cmd := exec.Command("pandoc", 
        "--filter" , "pandoc-crossref",
        "--pdf-engine", "xelatex", 
        "--template", bgPath + "/cpa", 
        "--listings", 
        "--file-scope", 
        "--verbose",
        "--resource-path", resourcesPath,
        headerPath, mdFilePath,
        "-o", 
        pdfFilePath)   
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Erro ao converter PDF: %v\n", err)
        log.Printf("Saída do comando: %s\n", string(output))
        return fmt.Errorf("Erro ao converter %s para PDF: %v\n%s", mdFilePath, err, string(output))
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

// copyLogFile copia o arquivo de log para o diretório de destino
func copyLogFile(pdfDir, destDir string) error {
    logFile := filepath.Join(pdfDir, "conversion_errors.log")
    if _, err := os.Stat(logFile); os.IsNotExist(err) {
        return nil // arquivo não existe, não é um erro
    }

    destFile := filepath.Join(destDir, "conversion_errors.log")
    input, err := os.ReadFile(logFile)
    if err != nil {
        return err
    }

    return os.WriteFile(destFile, input, 0644)
}

// uploadHandler lida com o upload do arquivo e processa a conversão.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
        return
    }

    r.ParseMultipartForm(200 << 20) // 10 MB

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
                log.Printf("Erro ao converter arquivo %s: %v\n", path, err)
                cmdOutput := ""
                if strings.Contains(err.Error(), "Saída do comando:") {
                    parts := strings.SplitN(err.Error(), "\n", 2)
                    if len(parts) > 1 {
                        cmdOutput = parts[1]
                    }
                }
                if logErr := writeErrorLog(pdfDir, path, err, cmdOutput); logErr != nil {
                    log.Printf("Erro ao gravar log: %v\n", logErr)
                }
                return nil // Continue processando outros arquivos
            }
            runtime.GC() // Força a execução do garbage collector após cada conversão
        }
        return nil
    }); err != nil {
        http.Error(w, "Erro ao acessar arquivos Markdown", http.StatusInternalServerError)
        return
    }

    // Combina cada PDF gerado com a capa
    combinedDir := "../data/combined_pdf"
    if err := os.MkdirAll(combinedDir, os.ModePerm); err != nil {
        http.Error(w, "Erro ao criar diretório de PDFs combinados", http.StatusInternalServerError)
        return
    }

    if err := filepath.Walk(pdfDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".pdf" {
            combinedPath := filepath.Join(combinedDir, filepath.Base(path))
            if err := combinePDFs(path, combinedPath); err != nil {
                return err
            }
        }
        return nil
    }); err != nil {
        http.Error(w, "Erro ao combinar PDFs com capa", http.StatusInternalServerError)
        return
    }

    // Antes de zipar, copie o arquivo de log
    if err := copyLogFile(pdfDir, combinedDir); err != nil {
        log.Printf("Erro ao copiar arquivo de log: %v\n", err)
    }

    // Gerar nome do arquivo com timestamp
    timestamp := time.Now().Format("20060102_150405")
    zipFilePath := fmt.Sprintf("../data/arquivoPDF_%s.zip", timestamp)
    if err := zipDirectory(combinedDir, zipFilePath); err != nil {
        http.Error(w, "Erro ao comprimir a pasta PDF", http.StatusInternalServerError)
        return
    }

    if err := os.RemoveAll(dest); err != nil {
        http.Error(w, "Erro ao remover a pasta descomprimida", http.StatusInternalServerError)
        return
    }

    // Nova linha para remover o diretório de PDFs
    if err := os.RemoveAll(pdfDir); err != nil {
        http.Error(w, "Erro ao remover a pasta de PDFs", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(zipFilePath)))
    http.ServeFile(w, r, zipFilePath)
    
    // Remove the combined PDF directory after serving the file
    if err := os.RemoveAll(combinedDir); err != nil {
        log.Printf("Erro ao remover a pasta de PDFs combinados: %v\n", err)
    }
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
