# cobolc — COBOL → C compiler

[Português](README.md) | **English**

Course project (EP) for **Introduction to Compilers**. The goal is to build, in Go, a compiler for fixed-format COBOL that generates an equivalent C program, following the course specification (*Especificação Textual do Compilador COBOL*).

The project is delivered in phases. **The current phase is the first delivery: lexical analysis.** Besides the lexer, the program already recognizes the IDENTIFICATION DIVISION (§3.1 of the specification) and extracts the program name.

---

## Contents

- [What the program does today](#what-the-program-does-today)
- [Usage](#usage)
- [Sample output](#sample-output)
- [Architecture](#architecture)
- [Lexical analysis](#lexical-analysis)
- [Design decisions](#design-decisions)
- [Tests](#tests)
- [Project layout](#project-layout)
- [Next steps](#next-steps)
- [License](#license)

---

## What the program does today

Given a COBOL source file, `cobolc`:

1. **Reads the file** following COBOL's column layout.
2. **Produces the token stream**: turns the code into an ordered sequence of tokens (reserved words, identifiers, literals, operators, symbols), each with its line and column.
3. **Reports lexical errors** with line, column and a clear message. Analysis does not stop at the first error: all of them are listed at once.
4. **Identifies the program**: validates `IDENTIFICATION DIVISION.` + `PROGRAM-ID. <name>.` and extracts the name.
5. **Prints everything to the terminal** and, optionally, **writes the same report to a `.txt` file**.

There is **no** C code generation yet. It comes in the next phases (see [Next steps](#next-steps)).

> The report itself is written in Portuguese (the course language): e.g. `Fluxo de tokens` = token stream, `Erros léxicos` = lexical errors, `linha`/`col` = line/column.

---

## Usage

### Requirements

- Go **1.26** or later.

### Build

```bash
go build -o cobolc ./cmd/cobolc
```

### Run

```bash
./cobolc [-debug] [-o output.txt] <file.cob>
```

Or, without building first:

```bash
go run ./cmd/cobolc -o output.txt tests/input_programs/prog001_divisoes.cob
```

| Flag | Description |
|---|---|
| `-o <file.txt>` | Writes the report (token stream + errors) to this file. The directory is created if needed. The file is written even when there are errors. |
| `-debug` | Includes debugging lines (column 7 = `D`). Without it, they are skipped like comments. |

**Exit code:** `0` when there are no errors; `1` on a lexical error, a program identification error, or a failure reading/writing files.

### Run all tests

```bash
./scripts/run-tests.sh
```

Builds `cobolc` once and runs it on every program in `tests/input_programs/`, writing each result to `tests/outputs/<name>.txt`. It prints a summary at the end (`Resumo: 7/7 passaram`, i.e. 7/7 passed) and exits non-zero if any program fails. It can be run from any directory.

---

## Sample output

Real excerpt from `tests/outputs/prog001_divisoes.txt`:

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

When there are errors, the errors section lists each one with its line and column:

```
=== Erros léxicos (2) ===
[E] linha 3, col 16: caractere inválido '@'
[E] linha 4, col 20: literal alfanumérico não terminado: falta a aspa de fechamento (")

=== Identificação (§3.1) ===
PROGRAM-ID: EXEMPLO

Resultado: FALHA (2 erros léxicos)
```

---

## Architecture

The program is a layered pipeline, following the standard Go project layout (`cmd/` for the executable, `internal/` for packages):

```
.cob file ──► source ──► lexer ──► parser ──► report ──► terminal / .txt
                           │
                           └─► TokenStream (ordered tokens + errors)
```

| Package | Role |
|---|---|
| `cmd/cobolc` | Entry point: parses flags, runs the pipeline, prints and writes the report. |
| `internal/source` | Reads the source file. |
| `internal/token` | Token types, reserved words and the `Token` struct (type, lexeme, line, column). |
| `internal/lexer` | Lexical analysis. Produces the `TokenStream`, which holds the token stream **in order** and the errors found. |
| `internal/diagnostic` | Error representation, using the specification's severity levels (§6: `I`, `W`, `E`, `S`). |
| `internal/ast` | Syntax tree nodes (for now, `Program` and `IdentificationDivision`). |
| `internal/parser` | Parsing of the IDENTIFICATION DIVISION (§3.1). |
| `internal/report` | Builds the text report, the same for terminal and `.txt`. |

---

## Lexical analysis

### Column layout (§2.1)

| Columns | Area | Handling |
|---|---|---|
| 1–6 | Sequence | Ignored. |
| 7 | Indicator | `*` or `/` = comment · `-` = literal continuation · `D` = debugging line · space = normal line. Any other character is an error. |
| 8–11 | Area A | Read as code. |
| 12–72 | Area B | Read as code. |
| 73–80 | Program name area | Ignored. |

### Token categories

| Type | Meaning | Example |
|---|---|---|
| `KEYWORD` | Reserved word | `IDENTIFICATION`, `MOVE`, `PIC` |
| `IDENT` | Identifier (user-defined name) | `WRK-TOTAL` |
| `NUMBER` | Numeric literal, optional sign and decimal point | `10`, `-123.45` |
| `STRING` | Alphanumeric literal, in `"` or `'` | `"OLA, MUNDO!"` |
| `PICTURE-STRING` | PICTURE clause mask | `S9(05)V99` |
| `OPERATOR` | Arithmetic or relational operator | `=`, `>`, `**` |
| `PERIOD`, `COMMA`, `SEMICOLON`, `LPAREN`, `RPAREN` | Symbols `.` `,` `;` `(` `)` | `.` |
| `EOF` | End of file | |

**Reserved words:** those in §2.2 (`IDENTIFICATION`, `ENVIRONMENT`, `DATA`, `PROCEDURE`, `DIVISION`, `SECTION`, `DISPLAY`, `ACCEPT`, `PERFORM`, `STOP`, `RUN`, `IF`, `ELSE`, `MOVE`), plus `PROGRAM-ID` (§3.1) and `PIC`/`PICTURE` (§3.3).

### Lexical errors

All with severity `E`, in the format `[E] linha L, col C: <message>`:

| Situation | Message (Portuguese) |
|---|---|
| Character outside the language alphabet | `caractere inválido '@'` |
| `_` outside the program name | `caractere inválido '_'` |
| Literal missing its closing quote | `literal alfanumérico não terminado: falta a aspa de fechamento (")` |
| Literal continuation line without a quote | `continuação de literal deve iniciar com aspa (")` |
| Identifier longer than 30 characters | `identificador '...' excede 30 caracteres (tem N)` |
| Identifier starting or ending with a hyphen | `identificador '-ABC' não pode iniciar com hífen` / `... não pode terminar com hífen` |
| Number with more than one decimal point | `literal numérico inválido '1.2.3': mais de um ponto decimal` |
| Invalid character in column 7 | `indicador inválido na coluna 7: 'X'` |

---

## Design decisions

### General

| Decision | Reason |
|---|---|
| Implemented in **Go** | Performance. The specification allows any language for the compiler. |
| **C** backend | The professor restricted code generation to C or the JVM; C was chosen. |
| Layered layout (`cmd/` + `internal/`) | Each compiler phase lives in its own package. The next phases (full parser, semantics, C generation) plug in without rewriting the earlier ones. |
| The lexer **does not stop at the first error** | It records the error, skips the invalid text and keeps going, showing every problem in the file at once. |
| Token stream stored in the `TokenStream` struct | Tokens are kept in order together with the errors, and the report is built from it. |
| Columns counted by character (*rune*), not by byte | Accented characters inside strings (common in Portuguese) do not shift columns 8–72. |

### Interpretations and deviations from the specification

The professor's sample programs, used for validation, rely on features beyond the specification text. Each case was decided as follows:

| Decision | Reason |
|---|---|
| **`_` allowed only in the program name** (right after `PROGRAM-ID.` and `END PROGRAM`) | §2.2 does not allow `_`, but three of the professor's samples use it in the program name (`PROG006_FILES`). Anywhere else, `_` is still an error. |
| **`OPERATOR` category** (`+ - * / ** = > < >= <=`) | §2.2 only lists `. , ; ( )`, but four samples use operators (`COMPUTE X = ...`, `UNTIL IDX > 3`). |
| **Continuation (column 7 = `-`) only for literals** | That is what §2.1 describes. A `-` line with no open literal on the previous line is read as a normal line, without an error. |
| **PICTURE as a single token** | After `PIC`/`PICTURE`, the whole mask (`S9(05)V99`, `-ZZZZZZZZZ9.99`) becomes one token. Read word by word, `-ZZZ...` would raise a false leading-hyphen error. |
| **`PIC` and `PICTURE` as reserved words** | Mentioned in §3.3 and needed to recognize the mask. |
| **Case-insensitive** | `move`, `Move` and `MOVE` are the same word. Words and masks are uppercased; literal contents are preserved. |
| **`D` lines skipped by default; `-debug` includes them** | Follows the COBOL standard: debugging lines are only compiled in debug mode. |
| **Doubled quotes inside literals** | COBOL rule: `"ELE DISSE ""OI"""` is a single literal, and `""` stands for one quote. |

---

## Tests

- **Inputs:** `tests/input_programs/`, with the 7 sample programs provided by the professor (source: [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos)).
- **Outputs:** `tests/outputs/`, one `.txt` per program, with the full token stream, lexical errors and the program identification.
- **Current result:** **7/7** programs processed without lexical errors.

| Program | PROGRAM-ID | Tokens |
|---|---|---|
| `prog001_divisoes.cob` | `PROG0001` | 77 |
| `prog002_operacoes.cob` | `YOUR-PROGRAM-NAME` | 321 |
| `prog003_calculadora.cob` | `CALCULADORA` | 230 |
| `prog004_folha_pgto.cob` | `FOLHA-PGTO` | 223 |
| `prog005_strings.cob` | `PROG008_STRING` | 237 |
| `prog006_files.cob` | `PROG006_FILES` | 227 |
| `prog007_loja.cob` | `PROG007_LOJA` | 567 |

---

## Project layout

```
.
├── cmd/cobolc/            # executable
├── internal/
│   ├── ast/               # syntax tree
│   ├── diagnostic/        # errors and severities
│   ├── lexer/             # lexical analysis + TokenStream
│   ├── parser/            # parsing (§3.1)
│   ├── report/            # text report
│   ├── source/            # file reading
│   └── token/             # token types and reserved words
├── examples/hello.cbl     # minimal example
├── scripts/run-tests.sh   # runs cobolc on every test
└── tests/
    ├── input_programs/    # COBOL validation programs
    └── outputs/           # generated reports
```

---

## Next steps

Following the specification, the next phases are:

- **Full parsing (§3):** ENVIRONMENT, DATA and PROCEDURE DIVISION.
- **Semantic analysis (§4):** `MOVE` compatibility, identifier uniqueness, `PERFORM`/`GO TO` targets and file association.
- **C code generation (§5).**
- **Full error handling (§6):** all `I`/`W`/`E`/`S` levels and the `COPY` directive (copybooks).

---

## License

Released under the MIT License. See [LICENSE](LICENSE). The programs in `tests/input_programs/` come from [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos) and keep their original license (`tests/LICENSE`).
