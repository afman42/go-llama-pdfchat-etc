# GO LLAMA PDFCHAT and TXT

## Kebutuhan

- Install Golang Terbaru
- Install NodeJS Terbaru
- Install Ollama Models Terbaru
- Install Cmake Terbaru
- UPX (compress file)
- Model Embedding
  - How to chat pdf, use model `nomic-embed-text` - convert txt
  - How to chat txt, use model `nomic-embed-text`
- Model Chat, use model `llama3`, `qwen`, `deepseek`
- lama atau cepat pemrosesan tergantung spesifikasi sistem pc, laptop maupun server
- File example,bentukan harus dokumen, disclaimer utk di testing
  - text.txt : `https://github.com/jonathanhecl/gollama/blob/main/examples/rag/text.txt`
  - Lorem_ipsum.pdf : searching google
- put file `pdf2txt` into `/usr/local/bin` or `~/.local/bin` to run file globally
- should upload file inside text, not image or anything, just pure file text
 

## Todo

- [x] Add remove upload file
- [x] Make Responsive web in mobile and pc
- [] Need Ui Box chat improved
- [x] Add log from backend

## Menjalankan

- sebelum jalankan, `go mod tidy` perlu internet
- di dev, `make run/api`, `make run/web` di beda terminal
- di preview, `make run/preview_linux`
- di prod, `make deploy/prod` mau masukin ke `caprover`

## Kendala

- di prod, saat mengirimkan text terjadi 504 Gateway Time out nginx jadi naikan upstream nginx nya, alternatif jalan local atau intranet
