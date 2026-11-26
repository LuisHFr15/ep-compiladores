# cobolc — compilador COBOL → MIPS

**Português** | [English](README.en.md)

Trabalho (EP) da disciplina de **Introdução à Compiladores**. O `cobolc` é um compilador, escrito em Go, de COBOL em formato fixo (*fixed format*) para **assembly MIPS**, que roda no simulador **MARS**. Ele segue a especificação do curso (*Especificação Textual do Compilador COBOL*) e faz todas as fases: análise léxica, sintática, semântica e geração do código final, sem representação intermediária.

---

## Sumário

- [O que o programa faz](#o-que-o-programa-faz)
- [Como usar](#como-usar)
- [Arquitetura](#arquitetura)
- [Análise léxica](#análise-léxica)
- [Análise sintática e semântica](#análise-sintática-e-semântica)
- [Geração de MIPS](#geração-de-mips)
- [Decisões de projeto](#decisões-de-projeto)
- [Testes](#testes)
- [Limitações](#limitações)
- [Estrutura de pastas](#estrutura-de-pastas)
- [Licença](#licença)

---

## O que o programa faz

Dado um arquivo COBOL, o `cobolc`:

1. **Pré-processa** `COPY nome.` e `COPY nome REPLACING ==a== BY ==b==.`, expandindo o copybook (procurado na pasta do fonte e em `-I`).
2. **Gera o fluxo de tokens**, respeitando o layout de colunas, e reporta os erros léxicos.
3. **Analisa a sintaxe** das 4 divisões e dos comandos, reportando todos os erros de sintaxe encontrados.
4. **Analisa a semântica**: tabela de símbolos e as 4 checagens do §4 da especificação.
5. **Gera o `.asm` em MIPS**, que o MARS executa com a mesma saída do GnuCOBOL.

Se qualquer fase tiver erro, o programa mostra a lista (linha, coluna e mensagem), **não gera o `.asm`** e sai com código 1.

---

## Como usar

### Requisitos

- Go **1.26** ou superior.
- Para executar o resultado: Java e o **MARS** (`Mars4_5.jar`).
- Para os testes: `cobc` (GnuCOBOL 3.2) e o MARS em `tools/mars/Mars4_5.jar`.

### Compilar e executar

```bash
go build -o cobolc ./cmd/cobolc
./cobolc -asm saida.asm tests/input_programs/prog002_operacoes.cob
java -jar Mars4_5.jar nc sm saida.asm
```

Se `-asm` não for informado, o arquivo é gravado como `<nome do fonte>.asm` na pasta atual.

| Flag | Descrição |
|---|---|
| `-asm <arquivo.asm>` | Destino do código MIPS. |
| `-o <arquivo.txt>` | Grava o relatório completo (fluxo de tokens, erros de COPY, léxicos, sintáticos e semânticos). É gravado mesmo com erros. |
| `-tokens` | Imprime o fluxo de tokens no terminal (por padrão só aparecem os erros e o resultado). |
| `-I <pasta>` | Pasta extra para procurar copybooks (pode repetir). |
| `-debug` | Inclui as linhas de depuração (coluna 7 = `D`). Sem a flag, elas são ignoradas como comentário. |

**Código de saída:** `0` se compilou; `1` se houve erro em qualquer fase ou falha ao ler/gravar arquivos.

### Executar no MARS

O código usa só as syscalls 8, 10, 11, 13, 14, 15 e 16. Para saída idêntica ao GnuCOBOL (bytes acentuados), rode com:

```bash
java -Dfile.encoding=ISO-8859-1 -Dstdout.encoding=ISO-8859-1 -jar Mars4_5.jar nc me sm saida.asm < entrada.txt
```

`me` manda as mensagens do MARS para o stderr; sem isso, ele acrescenta uma linha em branco ao fim da saída.

---

## Arquitetura

```
fonte .cob/.cbl ─► preprocess ─► lexer ─► parser ─► semantic ─► codegen/mips ─► .asm
                   (COPY)         tokens    AST      símbolos      + runtime
```

| Pacote | Papel |
|---|---|
| `cmd/cobolc` | Ponto de entrada: flags, orquestração das fases, relatório. |
| `internal/source` | Leitura do arquivo fonte. |
| `internal/preprocess` | Expansão de `COPY`. |
| `internal/token`, `internal/lexer` | Tokens e análise léxica (`TokenStream`: tokens em ordem + erros). |
| `internal/diagnostic` | Erros com as severidades do §6 (`I`, `W`, `E`, `S`). |
| `internal/ast` | Nós da árvore sintática: dados, comandos, expressões, condições. |
| `internal/parser` | Análise sintática por descida recursiva, com recuperação por sentença. |
| `internal/semantic` | PIC, tabela de símbolos, layout de memória e checagens do §4. |
| `internal/codegen/mips` | Gerador de MIPS e `runtime/*.s` (rotinas embutidas com `go:embed`). |
| `internal/report` | Relatório em texto (terminal e `.txt`). |

---

## Análise léxica

### Layout de colunas (§2.1)

| Colunas | Área | Tratamento |
|---|---|---|
| 1–6 | Sequência | Ignorada. |
| 7 | Indicador | `*` ou `/` = comentário · `-` = continuação · `D` = depuração · espaço = linha normal. Outro caractere é erro. |
| 8–72 | Áreas A e B | Lidas como código. |
| 73–80 | Programa | Ignorada. |

A continuação vale para **literais** e para **palavras** partidas no fim da coluna 72 (`T` + `-AB-TRABALHADOR` = `TAB-TRABALHADOR`).

### Categorias de token

`KEYWORD`, `IDENT`, `NUMBER`, `STRING`, `PICTURE-STRING`, `OPERATOR`, `PERIOD`, `COMMA`, `SEMICOLON`, `LPAREN`, `RPAREN` e `EOF`. Os verbos e cláusulas que a especificação não lista (`ADD`, `INSPECT`, `SEARCH`…) saem como `IDENT` e são reconhecidos pelo parser.

### Erros léxicos

Todos `E`, no formato `[E] linha L, col C: mensagem`: caractere inválido, `_` fora do nome do programa, literal não terminado, continuação sem aspa, identificador com mais de 30 caracteres ou com hífen no início/fim, número com dois pontos decimais, indicador inválido na coluna 7 e **marcador de conflito de merge do git**. O texto livre dos parágrafos `AUTHOR`, `INSTALLATION`, `DATE-WRITTEN`, `DATE-COMPILED`, `SECURITY` e `REMARKS` (IDENTIFICATION) não gera erro de caractere inválido.

---

## Análise sintática e semântica

### Sintaxe

Reconhece `IDENTIFICATION` (inclusive `AUTHOR`, `REMARKS`…), `ENVIRONMENT` (`SPECIAL-NAMES`, `SELECT`), `DATA` (`FD`, níveis `01`–`49` e `77`, `PIC`, `VALUE(S)`, `OCCURS`, `INDEXED BY`, `USAGE`) e `PROCEDURE` com seções e parágrafos. Comandos: `DISPLAY`, `ACCEPT`, `MOVE`, `ADD`, `SUBTRACT`, `MULTIPLY`, `DIVIDE`, `COMPUTE` (com `GIVING`, `REMAINDER`, `ON SIZE ERROR`), `IF/ELSE` (com ou sem `END-IF`), `EVALUATE`, `PERFORM` (simples, `UNTIL`, `THRU`, `VARYING` e em linha com `END-PERFORM`), `INITIALIZE` (com `REPLACING`), `GO TO`, `SET`, `SEARCH`, `INSPECT`, `OPEN`, `CLOSE`, `READ`, `WRITE`, `START`, `STOP RUN`, `GOBACK`, `EXIT`, `CONTINUE` e `FUNCTION UPPER-CASE`. Expressões com `**`, `* /`, `+ -`; condições com `AND/OR/NOT`, relacionais por símbolo ou palavra (`EQUAL`, `GREATER`…), condição abreviada e `IS NUMERIC`.

### Semântica (§4)

| Checagem | Mensagem de exemplo |
|---|---|
| `MOVE` de `PIC A` para `PIC 9` | `MOVE inválido: item alfabético (PIC A) não pode ser movido para o item numérico 'NUM'` |
| Identificador duplicado, referência ambígua (resolvida com `OF`) | `identificador 'CONT' declarado mais de uma vez no mesmo escopo` · `referência ambígua a 'CAMPO'; qualifique com OF` |
| Destino de `PERFORM`/`GO TO` inexistente | `parágrafo ou seção 'FALTA' não declarado na PROCEDURE DIVISION` |
| Operação de arquivo sem `SELECT` e `FD` | `arquivo 'X' não declarado em FILE-CONTROL (SELECT)` · `arquivo 'Y' sem descritor FD na FILE SECTION` |

Também são verificados: identificador não declarado, subscritos faltando ou sobrando, `SEARCH` sem `INDEXED BY` e `VALUE` incompatível com o `PIC`.

---

## Geração de MIPS

- **Dados:** um bloco estático com a imagem inicial calculada em compilação (`VALUE`, zeros nos numéricos, espaços nos textos). Grupos, `OCCURS` e subscritos viram `base + (i-1)*tamanho`.
- **Números:** guardados como dígitos ASCII, como no COBOL (`DISPLAY`), com sinal por *overpunch*. A aritmética usa vetores de 40 dígitos decimais: soma, subtração, multiplicação, divisão longa (truncada), potência com expoente inteiro. Sem ponto flutuante, sem limite de 32 bits.
- **`PERFORM`:** pilha de retorno em software; cada fim de parágrafo ou seção testa o topo. Funciona com recursão e com seções.
- **`ON SIZE ERROR`:** destino intacto no estouro; sem a cláusula, os dígitos altos são truncados.
- **Máscaras editadas:** `PIC -ZZZZZZZZZ9.99`, `ZZZ.ZZZ.ZZZ,ZZ` (com `DECIMAL-POINT IS COMMA`), `Z`, `*`, `9`, `+`, `-`, `CR`, `DB`, `B`, `0`, `/`, `$` e separadores.
- **Arquivos:** `SEQUENTIAL` (registros fixos), `LINE SEQUENTIAL` e `INDEXED` (tabela em memória com busca pela chave, gravada no `CLOSE`), com `FILE STATUS` 00, 10, 22, 23, 34, 35, 37, 41, 42, 47 e 48.
- **`INSPECT`:** `TALLYING`, `REPLACING` e `CONVERTING`, com `BEFORE`/`AFTER INITIAL`.

---

## Decisões de projeto

| Decisão | Motivo |
|---|---|
| Implementação em **Go** | Performance. |
| Backend **MIPS (MARS)** | Recomendação do professor, por simplicidade. Substitui o C previsto no §5. |
| Lexer e parser escritos à mão | Boa parte das regras depende de coluna e de contexto. |
| Palavras COBOL contextuais no parser | Mantém o léxico estável e evita reservar nomes como `NAME`. |
| `_` só no nome do programa | O §2.2 não prevê `_`, mas três exemplos o usam em `PROGRAM-ID`. |
| Categoria `OPERATOR` | O §2.2 só lista `. , ; ( )`, mas os exemplos usam `=`, `>`, `**`. |
| Continuação de palavra | Sem ela, `TAB-TRABALHADOR` (`prog004`/`prog006`) não existiria. |
| Não diferencia maiúsculas de minúsculas | Padrão COBOL; literais são preservados. |
| Linhas `D` ignoradas; `-debug` as inclui | Padrão COBOL. |
| Saídas de referência do **GnuCOBOL** (`-std=ibm`) | Fonte de verdade para `DISPLAY` de campos com sinal, `ACCEPT` numérico e arquivos. |
| `CAD0005A`, `PROG024A` rejeitados | Marcadores de merge no fonte; `COPY` de copybook inexistente e `CALL 'PROGDATA'`. |
| `COPY002A.cpy` inferido do uso | O copybook de `CAD0002A` não existe no repositório; foi criado a partir dos nomes e tamanhos que o programa usa. |

---

## Testes

Os programas de validação estão em `tests/input_programs/`: os 7 do professor (origem: [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos)) e 7 `.cbl`. Entradas em `tests/stdin/`, fixtures em `tests/data/` e saídas esperadas (geradas pelo GnuCOBOL) em `tests/expected/`.

| Script | O que faz |
|---|---|
| `scripts/run-mips-tests.sh [programa]` | Compila cada programa de `tests/programs.txt`, roda no MARS e compara a saída (e os arquivos gerados) com `tests/expected/`. Os `.asm` ficam em `tests/asm/`. |
| `scripts/run-tests.sh` | Roda o compilador em cada programa e grava o relatório em `tests/outputs/`. |
| `scripts/run-error-tests.sh` | Confere os erros léxicos inseridos (`tests/error_programs/`), os erros semânticos (`tests/semantic_programs/`) e a rejeição dos 3 programas de `tests/rejected.txt`. |
| `scripts/gen-reference.sh` | Regenera `tests/expected/` com o `cobc`. |

Resultado atual: **13/13** programas com saída idêntica à do GnuCOBOL; **7/7** programas com erros léxicos detectados; **12/12** rejeições esperadas; **2/2** programas aceitos (texto livre e `COPY REPLACING`).

| Programa | Recursos |
|---|---|
| `prog001`–`prog003` | Divisões, aritmética, `ACCEPT`, `IF` aninhado, máscara editada, `**` |
| `prog004`, `prog005` | `OCCURS`, `SEARCH`, `SET`, `INSPECT` |
| `prog006`, `CAD0004A`, `BUSCA01A` | Arquivos sequenciais e de linha |
| `prog007` | Arquivo indexado, `START`, `INVALID KEY`, `EVALUATE`, `FILE STATUS` |
| `CAIXAREG`, `CALCMD01` | Decimais, `COMPUTE`, `PERFORM` recursivo |
| `CAD0002A` | `COPY`, `LINKAGE`/`USING`/`GOBACK`, referência por posição `X(i:n)`, `PERFORM VARYING`, `INITIALIZE`, `EVALUATE` com `WHEN LESS`/`GREATER`/`IS NOT NUMERIC`, `INSPECT FUNCTION REVERSE` |
| `feat001_refmod_varying` | Os mesmos recursos novos em um programa curto (`tests/feature_programs/`) |

---

## Limitações

- `**` só com expoente inteiro.
- `PERFORM n TIMES`, `CALL`, `REDEFINES`, `COMP`/`COMP-3` e `SCREEN SECTION` não são suportados e geram erro claro.
- O arquivo indexado é simulado (até 256 registros) e tem formato próprio, não compatível com ISAM.
- Depois de um `COPY` expandido, a numeração das linhas nos diagnósticos fica deslocada.
- `COPY ... REPLACING` aceita `==pseudo-texto==` e palavras; `LEADING`/`TRAILING`, literais como operando e nome de copybook entre aspas não são suportados.
- O MARS bloqueia um nome de arquivo depois de um `OPEN` de leitura que falha; o runtime contorna prefixando `./` ao caminho.

---

## Estrutura de pastas

```
.
├── cmd/cobolc/              # executável
├── internal/
│   ├── ast/  codegen/mips/  diagnostic/  lexer/  parser/
│   ├── preprocess/  report/  semantic/  source/  token/
│   └── codegen/mips/runtime/   # rotinas MIPS (core, str, num, edit, inspect, file)
├── scripts/                 # run-mips-tests, run-tests, run-error-tests, gen-reference
├── examples/hello.cbl
└── tests/
    ├── input_programs/  stdin/  data/  expected/  asm/  outputs/
    └── error_programs/  error_outputs/  semantic_programs/  rejected.txt  programs.txt
```

---

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE). Os programas de `tests/input_programs/` vêm de [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos) e mantêm a licença original (`tests/LICENSE`).
