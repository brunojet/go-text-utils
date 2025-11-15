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

    cd src
    go run .

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

O objeto desse projeto é explorarmos um código para geração de chaves para busca de filtros combinados para uma loja de aplicativos de terminais POS.

Hoje temos como desafio de não expormos a PK das entidades e ao mesmo tempo evitar chaves com nomes muito longos tendo risco de quebra de consultas com muitas junções de filtros.

O aplicativo espera que tenhamos filtros configuráveis para buscas de aplicativos, como por exemplo: região, ramo, subramo, funcionalidades e etc. O filtro pode ser autorelacionado, ou seja, um filtro pode ter subfiltros. Exemplo podemos ter o ramo alimentação e dentro dele termos relacionados subramos como lanchonete, restaurante, bar e etc. Os tipos de filtros e filtros devem corresponder a essa hierarquia.

A deia é que a chave de busca seja composta de ate 3 caracteres cada parte, onde teremos no minimo a combinacao de 2 partes: TipoFiltro + Filtro + (opcional) Subfiltro. Essa chave "curta" deve ser calculada a partir do nome original do filtro, de forma que seja unica e reproduzivel, no caso de colisao, devemos ter uma estrategia para gerar chaves diferentes. Tomar cuidado com carecteres especiais, acentos e etc. Vamos padronizar sempre em maiusculas e sem acentos no formato TTT-FFF-SSS-SSS onde TTT é o tipo do filtro, FFF o filtro e SSSS o subfiltro (opcional).



