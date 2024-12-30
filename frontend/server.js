require('dotenv').config(); // Carrega as variáveis do arquivo .env
const express = require('express');
const multer = require('multer');
const axios = require('axios');
const fs = require('fs');
const path = require('path');
const FormData = require('form-data'); // npm install --save form-data

const app = express();

// Adicione esta rota para servir o index.html
app.get('/', (req, res) => {
    console.log('Requisição recebida para /');
    res.sendFile(path.join(__dirname, 'public/index.html'));
});

// Configuração do multer para armazenar arquivos na pasta uploads
const upload = multer({ dest: '../data/uploads/' }); // Usando caminho relativo

app.post('/upload', upload.single('file'), async (req, res) => {
    // Declarar filePath antes do bloco try
    const filePath = req.file ? `../data/uploads/${req.file.filename}` : null;

    try {
        // Validações iniciais
        if (!req.file) {
            return res.status(400).send('Nenhum arquivo enviado');
        }

        // Verificação de tipo de arquivo
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

        // Resto do código permanece igual
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

        // Gerar nome para o arquivo de resultado
        const timestamp = new Date().toISOString().replace(/:/g, '-');
        const resultFilename = `resultado_${timestamp}.zip`;

        // Configurar cabeçalhos para download
        res.set({
            'Content-Type': 'application/zip',
            'Content-Disposition': `attachment; filename="${resultFilename}"`
        });

        // Enviar o arquivo ZIP de volta para o cliente
        res.send(response.data);

    } catch (error) {
        // Tratamento de erro detalhado
        console.error('Erro no processamento:', error);

        if (error.code === 'ECONNABORTED') {
            return res.status(504).send('Tempo limite de processamento excedido');
        }

        if (error.response) {
            return res.status(error.response.status).send(error.response.data);
        } else {
            return res.status(500).send('Erro interno no processamento');
        }
    } finally {
        // Verificação adicional antes de excluir
        if (filePath && fs.existsSync(filePath)) {
            fs.unlink(filePath, (err) => {
                if (err) console.error('Erro ao excluir arquivo temporário:', err);
            });
        }
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Servidor rodando na porta ${PORT}`);
});