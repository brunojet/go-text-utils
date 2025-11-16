Demo: cmd/demo

Este diretório contém um pequeno binário de demonstração que mostra como usar a API pública `pkg/acronym`.

Construir e executar (Windows):

    go build ./cmd/demo
    .\cmd\demo\demo.exe "Alimentação" "Educação a distância"

No Linux/macOS:

    go build ./cmd/demo
    ./cmd/demo "Alimentação" "Educação a distância"

O demo reusa um mapa `usedKeys` entre chamadas para evidenciar como colisões são evitadas em uma execução única.
