require('dotenv').config(); // Carrega as variáveis do arquivo .env
const express = require('express');
const multer = require('multer');
const axios = require('axios');
const fs = require('fs');
const path = require('path');
const FormData = require('form-data'); // npm install --save form-data
const form = new FormData();

const app = express();

// Adicione esta rota para servir o index.html
app.get('/', (req, res) => {
    console.log('Requisição recebida para /');
    res.sendFile(path.join(__dirname, 'public/index.html'));
});

const upload = multer({ dest: 'uploads/' });

app.post('/upload', upload.single('entrada'), async (req, res) => {
    console.log('Arquivo recebido:', req.file.originalname); // Log do nome do arquivo recebido
    try {
        const filePath = path.join(__dirname, req.file.path);
        const apiUrl = process.env.API_URL; // Obtém a URL da API da variável de ambiente

        // Verifique se a variável de ambiente API_URL está definida
        if (!apiUrl) {
            console.error('API_URL não está definida.');
            return res.status(500).send('API_URL não está definida.');
        }

        console.log('Enviando arquivo para a API:', apiUrl);
        form.append('entrada', fs.createReadStream(filePath));
        const request_config = {
            headers: {
                'Content-Type': 'application/zip'
            },
            responseType: 'arraybuffer'
        };
        try {
            const response = await axios.post(apiUrl, form, request_config);

            // Verifique se a resposta é um arquivo ZIP
            const contentType = response.headers['content-type'];
            if (!contentType || !contentType.includes('application/zip')) {
                console.error('A resposta da API não é um arquivo ZIP. Tipo de conteúdo:', contentType);
                return res.status(500).send('A resposta da API não é um arquivo ZIP.');
            }

            // Verifique se a resposta contém dados
            if (!response.data || response.data.length === 0) {
                console.error('A resposta da API está vazia.');
                return res.status(500).send('A resposta da API está vazia.');
            }

            console.log('Arquivo recebido da API, tamanho:', response.data.length); // Log do tamanho do arquivo recebido

            // Salvar o arquivo ZIP recebido
            const outputFilePath = path.join(__dirname, 'output.zip');
            fs.writeFileSync(outputFilePath, response.data);
            console.log('Arquivo ZIP salvo como output.zip');

            // Remover o arquivo da pasta uploads
            fs.unlink(filePath, (err) => {
                if (err) {
                    console.error('Erro ao remover o arquivo:', err);
                } else {
                    console.log('Arquivo removido da pasta uploads:', filePath);
                }
            });

            // Enviar o arquivo para o cliente
            return res.download(outputFilePath, 'output.zip', (err) => {
                if (err) {
                    console.error('Erro ao enviar o arquivo:', err);
                    return res.status(500).send('Erro ao enviar o arquivo');
                }
                console.log('Arquivo enviado com sucesso:', outputFilePath);
            });
        } catch (error) {
            console.error('Erro ao processar o arquivo:', error.message);
            console.error('Erro detalhes:', error.response ? error.response.data : error);
            return res.status(500).send('Erro ao processar o arquivo');
        }
    } catch (error) {
        console.error('Erro ao processar o arquivo:', error); // Log do erro para depuração
        return res.status(500).send('Erro ao processar o arquivo');
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Servidor rodando na porta ${PORT}`);
});