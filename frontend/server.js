const express = require('express');
const multer = require('multer');
const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');
const path = require('path');

const app = express();

// Middleware para servir arquivos estáticos
app.use(express.static(path.join(__dirname, './public')));

// Configuração de upload
const upload = multer({ dest: '../data/uploads/' });

// Diretório de resultados
const resultDir = path.join(__dirname, '../data/resultados');
if (!fs.existsSync(resultDir)) {
    fs.mkdirSync(resultDir, { recursive: true });
}

// Rota para página inicial
app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, './public/index.html'));
});

app.post('/upload', upload.single('file'), async (req, res) => {
    const filePath = req.file ? `../data/uploads/${req.file.filename}` : null;

    try {
        // Validações de arquivo
        if (!req.file) {
            return res.status(400).send('Nenhum arquivo enviado');
        }

        if (req.file.mimetype !== 'application/zip') {
            return res.status(400).send('Apenas arquivos ZIP são permitidos');
        }

        const apiUrl = process.env.API_URL;

        // Configuração do FormData
        const form = new FormData();
        form.append('file', fs.createReadStream(filePath), {
            filename: req.file.originalname,
            contentType: 'application/zip'
        });

        // Envio para API com timeout de 30 minutos
        const response = await axios.post(apiUrl, form, {
            headers: {
                ...form.getHeaders(),
                'Content-Type': 'multipart/form-data'
            },
            responseType: 'arraybuffer', 
            maxBodyLength: Infinity,
            maxContentLength: Infinity,
            timeout: 30 * 60 * 1000, 
        });

        // Gerar nome do arquivo de resultado
        const timestamp = new Date().toISOString().replace(/:/g, '-');
        const resultFilename = `resultado_${timestamp}.zip`;
        const resultFilePath = path.join(resultDir, resultFilename);

        // Salvar arquivo de resultado
        fs.writeFileSync(resultFilePath, response.data);

        // Configurar cabeçalhos para download
        res.set({
            'Content-Type': 'application/zip',
            'Content-Disposition': `attachment; filename="${resultFilename}"`
        });

        // Enviar arquivo ZIP
        res.send(response.data);

    } catch (error) {
        console.error('Erro no processamento:', error);

        if (error.code === 'ECONNABORTED') {
            return res.status(504).send('Tempo limite excedido');
        }

        if (error.response) {
            return res.status(error.response.status).send(error.response.data);
        } else {
            return res.status(500).send('Erro interno');
        }
    } finally {
        // Limpar arquivo temporário
        if (filePath && fs.existsSync(filePath)) {
            fs.unlink(filePath, (err) => {
                if (err) console.error('Erro ao excluir arquivo:', err);
            });
        }
    }
});

// Rota para listar resultados
app.get('/resultados', (req, res) => {
    try {
        const files = fs.readdirSync(resultDir);
        const fileDetails = files.map(file => {
            const filePath = path.join(resultDir, file);
            const stats = fs.statSync(filePath);
            
            return {
                name: file,
                path: `/download/${file}`,
                size: stats.size,
                sizeFormatted: formatFileSize(stats.size),
                createdAt: stats.birthtime.toLocaleString('pt-BR', {
                    dateStyle: 'short',
                    timeStyle: 'short'
                })
            };
        })
        // Ordenar por data de criação decrescente
        .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

        res.json(fileDetails);
    } catch (error) {
        console.error('Erro ao listar resultados:', error);
        res.status(500).send('Erro ao listar arquivos');
    }
});

// Função para formatar tamanho do arquivo
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Rota para download de arquivos
app.get('/download/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(resultDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).send('Arquivo não encontrado');
    }

    res.download(filePath, filename, (err) => {
        if (err) {
            console.error('Erro no download:', err);
            res.status(500).send('Erro ao baixar arquivo');
        }
    });
});

// Rota para limpar arquivos antigos
app.delete('/limpar-resultados', (req, res) => {
    try {
        const files = fs.readdirSync(resultDir);
        const now = Date.now();

        files.forEach(file => {
            const filePath = path.join(resultDir, file);
            const stats = fs.statSync(filePath);
            
            // Excluir arquivos com mais de 24 horas
            if (now - stats.birthtimeMs > 24 * 60 * 60 * 1000) {
                fs.unlinkSync(filePath);
            }
        });

        res.send('Arquivos antigos removidos');
    } catch (error) {
        console.error('Erro ao limpar resultados:', error);
        res.status(500).send('Erro ao limpar arquivos');
    }
});

// Iniciar servidor
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Servidor rodando na porta ${PORT}`);
});