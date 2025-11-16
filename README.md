# go-text-utils

Collection of small Go utilities for text and string processing, with a focus on generating deterministic short keys/acronyms for filter names.

Features
- Deterministic 3-character short name generation with consonant/vowel heuristics and MD5→base36 fallback.
- Text normalization (accent removal, connector word removal) tuned for Portuguese but implemented generically.
- Small internal packages for text utilities and encoding to keep code testable and reusable.

Quick start

1. Install

    go get github.com/brunojet/go-text-utils

2. Run the demo (inside repo)

    # build and run the small demo binary (recommended)
    go build ./cmd/demo
    ./cmd/demo/demo.exe  # on Windows; on *nix use ./cmd/demo/demo

3. Use the package

Import the acronym generator:

    import "github.com/brunojet/go-text-utils/pkg/acronym"

Call the public helper:

    short, mode, err := acronym.TryExtractSignificant("Alimentação", map[string]bool{})

Notes
- The module path in `go.mod` is `github.com/brunojet/go-text-utils`.
- Logs and persisted mappings are in Portuguese for domain clarity.

License

MIT — see `LICENSE`

# Projeto go-demo-filtros-curtos

Objetivo
--------

Este projeto explora um gerador de chaves curtas (acrónios) para filtros hierárquicos usados em buscas de aplicações para terminais POS.

Motivação
----------

Queremos evitar expor chaves primárias (PK) nas consultas e, ao mesmo tempo, reduzir o comprimento das chaves usadas em junções e índices — chaves longas podem degradar desempenho ou quebrar consultas complexas.

Modelagem de filtros
--------------------

O sistema trabalha com tipos de filtros e filtros que podem ser auto-relacionados (ou seja, um filtro pode ter subfiltros). Exemplos: "Ramo" (ex.: Alimentação) e, dentro dele, sub-ramos como "Lanchonete", "Restaurante", "Bar". A hierarquia de tipos e filtros deve ser respeitada.

Especificação da chave curta
---------------------------

A ideia é que cada parte da chave tenha até 3 caracteres, e que a chave final seja composta por pelo menos duas partes: TipoFiltro + Filtro (+ opcionalmente Subfiltro). A chave curta deve ser gerada de forma determinística a partir do nome original do filtro, garantindo unicidade sempre que possível.

Em caso de colisão a estratégia do gerador é empregar heurísticas (consoantes/vogais), tentativas alternativas e um fallback determinístico (hash MD5 mapeado para base36 com sondagem linear) para produzir uma chave diferente e reproduzível.

Observações de normalização
---------------------------

Tomamos cuidado com caracteres especiais e acentuação: normalizamos para letras maiúsculas sem acentos e removemos palavras de ligação (conectores) quando apropriado. O formato genérico esperado é algo como TTT-FFF-SSS, onde cada bloco tem até 3 caracteres.

Se quiser que a saída persista as chaves geradas no arquivo de configuração, o gerador grava um mapeamento canônico em `data/nomeParaCurto.json` (este arquivo normalmente fica em `.gitignore`).

Licença
-------

MIT — ver `LICENSE`



