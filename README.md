# cobolc — compilador COBOL → C

**Português** | [English](README.en.md)

Trabalho (EP) da disciplina de **Introdução à Compiladores**. O objetivo é construir, em Go, um compilador para COBOL em formato fixo (*fixed format*) que gere um programa C equivalente, seguindo a especificação do curso (*Especificação Textual do Compilador COBOL*).

O trabalho é entregue em fases. **A fase atual é a primeira entrega: a análise léxica.** Além do lexer, o programa já reconhece a IDENTIFICATION DIVISION (§3.1 da especificação) e extrai o nome do programa.

---

## Sumário

- [O que o programa faz hoje](#o-que-o-programa-faz-hoje)
- [Como usar](#como-usar)
- [Exemplo de saída](#exemplo-de-saída)
- [Arquitetura](#arquitetura)
- [Análise léxica](#análise-léxica)
- [Decisões de projeto](#decisões-de-projeto)
- [Testes](#testes)
- [Estrutura de pastas](#estrutura-de-pastas)
- [Próximos passos](#próximos-passos)
- [Licença](#licença)

---

## O que o programa faz hoje

Dado um arquivo COBOL, o `cobolc`:

1. **Lê o arquivo** respeitando o layout de colunas do COBOL.
2. **Gera o fluxo de tokens**: converte o código em uma sequência ordenada de tokens (palavras reservadas, identificadores, literais, operadores, símbolos), cada um com linha e coluna.
3. **Reporta erros léxicos** com linha, coluna e uma mensagem clara. A análise não para no primeiro erro: todos são listados de uma vez.
4. **Identifica o programa**: valida `IDENTIFICATION DIVISION.` + `PROGRAM-ID. <nome>.` e extrai o nome.
5. **Mostra tudo no terminal** e, opcionalmente, **grava o mesmo relatório em um `.txt`**.

Ainda **não** há geração de código C. Ela vem nas próximas fases (veja [Próximos passos](#próximos-passos)).

---

## Como usar

### Requisitos

- Go **1.26** ou superior.

### Compilar

```bash
go build -o cobolc ./cmd/cobolc
```

### Executar

```bash
./cobolc [-debug] [-o saida.txt] <arquivo.cob>
```

Ou, sem compilar antes:

```bash
go run ./cmd/cobolc -o saida.txt tests/input_programs/prog001_divisoes.cob
```

| Flag | Descrição |
|---|---|
| `-o <arquivo.txt>` | Grava o relatório (fluxo de tokens + erros) neste arquivo. A pasta é criada se não existir. O arquivo é gravado mesmo se houver erros. |
| `-debug` | Inclui as linhas de depuração (coluna 7 = `D`). Sem a flag, elas são ignoradas como comentário. |

**Código de saída:** `0` se não houver erros; `1` se houver erro léxico, erro na identificação do programa, ou falha ao ler/gravar arquivos.

### Rodar todos os testes

```bash
./scripts/run-tests.sh
```

Compila o `cobolc` uma vez e roda em cada programa de `tests/input_programs/`, gravando o resultado de cada um em `tests/outputs/<nome>.txt`. Ao final mostra o resumo (`Resumo: 7/7 passaram`) e sai com código diferente de zero se algum falhar. Pode ser executado de qualquer pasta.

---

## Exemplo de saída

Trecho real de `tests/outputs/prog001_divisoes.txt`:

```
Arquivo: tests/input_programs/prog001_divisoes.cob

=== Fluxo de tokens (77) ===
#      LINHA  COL   TIPO             LEXEMA
1      15     8     KEYWORD          IDENTIFICATION
2      15     23    KEYWORD          DIVISION
3      15     31    PERIOD           .
4      16     8     KEYWORD          PROGRAM-ID
5      16     18    PERIOD           .
6      16     20    IDENT            PROG0001
7      16     28    PERIOD           .
...

=== Erros léxicos (0) ===
nenhum

=== Identificação (§3.1) ===
PROGRAM-ID: PROG0001

Resultado: SUCESSO
```

Com erros, a seção de erros lista cada um com linha e coluna:

```
=== Erros léxicos (2) ===
[E] linha 3, col 16: caractere inválido '@'
[E] linha 4, col 20: literal alfanumérico não terminado: falta a aspa de fechamento (")

=== Identificação (§3.1) ===
PROGRAM-ID: EXEMPLO

Resultado: FALHA (2 erros léxicos)
```

---

## Arquitetura

O programa é um pipeline em camadas, seguindo o layout padrão de projetos Go (`cmd/` para o executável, `internal/` para os pacotes):

```
arquivo .cob ──► source ──► lexer ──► parser ──► report ──► terminal / .txt
                              │
                              └─► TokenStream (tokens em ordem + erros)
```

| Pacote | Papel |
|---|---|
| `cmd/cobolc` | Ponto de entrada: lê as flags, orquestra o pipeline, imprime e grava o relatório. |
| `internal/source` | Leitura do arquivo fonte. |
| `internal/token` | Tipos de token, palavras reservadas e a struct `Token` (tipo, lexema, linha, coluna). |
| `internal/lexer` | Análise léxica. Produz a `TokenStream`, que guarda o fluxo de tokens **em ordem** e os erros encontrados. |
| `internal/diagnostic` | Representação dos erros, com os níveis de severidade da especificação (§6: `I`, `W`, `E`, `S`). |
| `internal/ast` | Nós da árvore sintática (por enquanto, `Program` e `IdentificationDivision`). |
| `internal/parser` | Análise sintática da IDENTIFICATION DIVISION (§3.1). |
| `internal/report` | Monta o relatório em texto, o mesmo para terminal e `.txt`. |

---

## Análise léxica

### Layout de colunas (§2.1)

| Colunas | Área | Tratamento |
|---|---|---|
| 1–6 | Sequência | Ignorada. |
| 7 | Indicador | `*` ou `/` = comentário · `-` = continuação de literal · `D` = linha de depuração · espaço = linha normal. Qualquer outro caractere é erro. |
| 8–11 | Área A | Lida como código. |
| 12–72 | Área B | Lida como código. |
| 73–80 | Programa | Ignorada. |

### Categorias de token

| Tipo | O que é | Exemplo |
|---|---|---|
| `KEYWORD` | Palavra reservada | `IDENTIFICATION`, `MOVE`, `PIC` |
| `IDENT` | Identificador (nome definido pelo programador) | `WRK-TOTAL` |
| `NUMBER` | Literal numérico, com sinal e decimal opcionais | `10`, `-123.45` |
| `STRING` | Literal alfanumérico, entre `"` ou `'` | `"OLA, MUNDO!"` |
| `PICTURE-STRING` | Máscara da cláusula PICTURE | `S9(05)V99` |
| `OPERATOR` | Operador aritmético ou relacional | `=`, `>`, `**` |
| `PERIOD`, `COMMA`, `SEMICOLON`, `LPAREN`, `RPAREN` | Símbolos `.` `,` `;` `(` `)` | `.` |
| `EOF` | Fim do arquivo | |

**Palavras reservadas:** as do §2.2 (`IDENTIFICATION`, `ENVIRONMENT`, `DATA`, `PROCEDURE`, `DIVISION`, `SECTION`, `DISPLAY`, `ACCEPT`, `PERFORM`, `STOP`, `RUN`, `IF`, `ELSE`, `MOVE`), mais `PROGRAM-ID` (§3.1) e `PIC`/`PICTURE` (§3.3).

### Erros léxicos

Todos com severidade `E`, no formato `[E] linha L, col C: mensagem`:

| Situação | Mensagem |
|---|---|
| Caractere fora do alfabeto da linguagem | `caractere inválido '@'` |
| `_` fora do nome do programa | `caractere inválido '_'` |
| Literal sem a aspa de fechamento | `literal alfanumérico não terminado: falta a aspa de fechamento (")` |
| Linha de continuação de literal sem aspa | `continuação de literal deve iniciar com aspa (")` |
| Identificador com mais de 30 caracteres | `identificador '...' excede 30 caracteres (tem N)` |
| Identificador começando ou terminando com hífen | `identificador '-ABC' não pode iniciar com hífen` / `... não pode terminar com hífen` |
| Número com mais de um ponto decimal | `literal numérico inválido '1.2.3': mais de um ponto decimal` |
| Caractere inválido na coluna 7 | `indicador inválido na coluna 7: 'X'` |

---

## Decisões de projeto

### Gerais

| Decisão | Motivo |
|---|---|
| Implementação em **Go** | Performance. A especificação permite qualquer linguagem para o compilador. |
| Backend em **C** | O professor restringiu a geração a C ou JVM; C foi escolhido. |
| Layout em camadas (`cmd/` + `internal/`) | Cada fase do compilador fica em um pacote. As próximas fases (parser completo, semântica, geração de C) entram sem reescrever as anteriores. |
| O lexer **não para no primeiro erro** | Registra o erro, pula o trecho inválido e continua, mostrando todos os problemas do arquivo de uma vez. |
| Fluxo de tokens guardado na struct `TokenStream` | Os tokens ficam em ordem, junto com os erros, e o relatório é montado a partir dela. |
| Colunas contadas por caractere (*rune*), não por byte | Acentos dentro de strings (comuns em português) não deslocam as colunas 8–72. |

### Interpretações e divergências da especificação

Os programas de exemplo do professor, usados na validação, usam recursos que vão além do texto da especificação. Cada caso foi decidido assim:

| Decisão | Motivo |
|---|---|
| **`_` permitido só no nome do programa** (logo após `PROGRAM-ID.` e `END PROGRAM`) | O §2.2 não permite `_`, mas três exemplos do professor usam `_` no nome do programa (`PROG006_FILES`). Em qualquer outro lugar, `_` continua sendo erro. |
| **Categoria `OPERATOR`** (`+ - * / ** = > < >= <=`) | O §2.2 só lista `. , ; ( )`, mas quatro exemplos usam operadores (`COMPUTE X = ...`, `UNTIL IDX > 3`). |
| **Continuação (coluna 7 = `-`) só para literais** | É o que o §2.1 descreve. Uma linha `-` sem literal aberto na anterior é lida como linha normal, sem erro. |
| **PICTURE como token único** | Após `PIC`/`PICTURE`, a máscara inteira (`S9(05)V99`, `-ZZZZZZZZZ9.99`) vira um token. Lida palavra por palavra, `-ZZZ...` geraria um falso erro de hífen inicial. |
| **`PIC` e `PICTURE` como palavras reservadas** | Citadas no §3.3 e necessárias para reconhecer a máscara. |
| **Não diferencia maiúsculas de minúsculas** | `move`, `Move` e `MOVE` são a mesma palavra. Palavras e máscaras são convertidas para maiúsculas; o conteúdo dos literais é preservado. |
| **Linhas `D` ignoradas por padrão; `-debug` as inclui** | Segue o padrão COBOL: linha de depuração só é compilada em modo debug. |
| **Aspas duplicadas dentro de literais** | Regra do COBOL: `"ELE DISSE ""OI"""` é um único literal, e `""` representa uma aspa. |

---

## Testes

- **Entradas:** `tests/input_programs/`, com os 7 programas de exemplo fornecidos pelo professor (origem: [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos)).
- **Saídas:** `tests/outputs/`, com um `.txt` por programa, contendo o fluxo de tokens completo, os erros léxicos e a identificação do programa.
- **Resultado atual:** **7/7** programas processados sem erros léxicos.

| Programa | PROGRAM-ID | Tokens |
|---|---|---|
| `prog001_divisoes.cob` | `PROG0001` | 77 |
| `prog002_operacoes.cob` | `YOUR-PROGRAM-NAME` | 321 |
| `prog003_calculadora.cob` | `CALCULADORA` | 230 |
| `prog004_folha_pgto.cob` | `FOLHA-PGTO` | 223 |
| `prog005_strings.cob` | `PROG008_STRING` | 237 |
| `prog006_files.cob` | `PROG006_FILES` | 227 |
| `prog007_loja.cob` | `PROG007_LOJA` | 567 |

---

## Estrutura de pastas

```
.
├── cmd/cobolc/            # executável
├── internal/
│   ├── ast/               # árvore sintática
│   ├── diagnostic/        # erros e severidades
│   ├── lexer/             # análise léxica + TokenStream
│   ├── parser/            # análise sintática (§3.1)
│   ├── report/            # relatório em texto
│   ├── source/            # leitura do arquivo
│   └── token/             # tipos de token e palavras reservadas
├── examples/hello.cbl     # exemplo mínimo
├── scripts/run-tests.sh   # roda o cobolc em todos os testes
└── tests/
    ├── input_programs/    # programas COBOL de validação
    └── outputs/           # relatórios gerados
```

---

## Próximos passos

Conforme a especificação, as próximas fases são:

- **Análise sintática completa (§3):** ENVIRONMENT, DATA e PROCEDURE DIVISION.
- **Análise semântica (§4):** compatibilidade de `MOVE`, unicidade de identificadores, destinos de `PERFORM`/`GO TO` e associação de arquivos.
- **Geração de código C (§5).**
- **Tratamento de erros completo (§6):** todos os níveis `I`/`W`/`E`/`S` e a diretiva `COPY` (copybooks).

---

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE). Os programas de `tests/input_programs/` vêm de [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos) e mantêm a licença original (`tests/LICENSE`).
