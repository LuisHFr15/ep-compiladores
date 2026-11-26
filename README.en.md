# cobolc — COBOL → MIPS compiler

[Português](README.md) | **English**

Course project (EP) for **Introduction to Compilers**. `cobolc` is a compiler written in Go that translates fixed-format COBOL into **MIPS assembly** to run on the **MARS** simulator. It follows the course specification (*COBOL Compiler Textual Specification*) and implements every phase: lexical, syntax and semantic analysis, plus final code generation, with no intermediate representation.

> Note: the compiler's messages and the generated reports are in Portuguese.

---

## Contents

- [What it does](#what-it-does)
- [Usage](#usage)
- [Architecture](#architecture)
- [Lexical analysis](#lexical-analysis)
- [Syntax and semantic analysis](#syntax-and-semantic-analysis)
- [MIPS generation](#mips-generation)
- [Design decisions](#design-decisions)
- [Tests](#tests)
- [Limitations](#limitations)
- [Folder layout](#folder-layout)
- [License](#license)

---

## What it does

Given a COBOL file, `cobolc`:

1. **Preprocesses** `COPY name.`, expanding the copybook (searched in the source folder and in `-I`).
2. **Produces the token stream**, honoring the column layout, and reports lexical errors.
3. **Parses** the 4 divisions and the statements, reporting every syntax error it finds.
4. **Analyzes semantics**: symbol table and the 4 checks of §4 of the specification.
5. **Generates MIPS `.asm`**, which MARS runs with the same output as GnuCOBOL.

If any phase fails, the program lists the errors (line, column, message), **does not write the `.asm`** and exits with code 1.

---

## Usage

### Requirements

- Go **1.26** or newer.
- To run the result: Java and **MARS** (`Mars4_5.jar`).
- For the tests: `cobc` (GnuCOBOL 3.2) and MARS at `tools/mars/Mars4_5.jar`.

### Build and run

```bash
go build -o cobolc ./cmd/cobolc
./cobolc -asm out.asm tests/input_programs/prog002_operacoes.cob
java -jar Mars4_5.jar nc sm out.asm
```

If `-asm` is omitted, the file is written as `<source name>.asm` in the current folder.

| Flag | Description |
|---|---|
| `-asm <file.asm>` | MIPS output file. |
| `-o <file.txt>` | Writes the full report (token stream, COPY, lexical, syntax and semantic errors). Written even on errors. |
| `-tokens` | Prints the token stream to the terminal (by default only errors and the result are shown). |
| `-I <folder>` | Extra folder to search copybooks (repeatable). |
| `-debug` | Includes debugging lines (column 7 = `D`). Without it they are ignored as comments. |

**Exit code:** `0` on success; `1` on an error in any phase or a read/write failure.

### Running on MARS

The code only uses syscalls 8, 10, 11, 13, 14, 15 and 16. For output identical to GnuCOBOL (accented bytes), run:

```bash
java -Dfile.encoding=ISO-8859-1 -Dstdout.encoding=ISO-8859-1 -jar Mars4_5.jar nc me sm out.asm < input.txt
```

`me` sends MARS messages to stderr; without it MARS appends an empty line to the output.

---

## Architecture

```
source .cob/.cbl ─► preprocess ─► lexer ─► parser ─► semantic ─► codegen/mips ─► .asm
                    (COPY)         tokens    AST      symbols      + runtime
```

| Package | Role |
|---|---|
| `cmd/cobolc` | Entry point: flags, phase orchestration, report. |
| `internal/source` | Source file reading. |
| `internal/preprocess` | `COPY` expansion. |
| `internal/token`, `internal/lexer` | Tokens and lexical analysis (`TokenStream`: ordered tokens + errors). |
| `internal/diagnostic` | Errors with the §6 severities (`I`, `W`, `E`, `S`). |
| `internal/ast` | Syntax tree nodes: data, statements, expressions, conditions. |
| `internal/parser` | Recursive-descent parser with per-sentence error recovery. |
| `internal/semantic` | PIC, symbol table, memory layout and §4 checks. |
| `internal/codegen/mips` | MIPS generator and `runtime/*.s` (routines embedded with `go:embed`). |
| `internal/report` | Text report (terminal and `.txt`). |

---

## Lexical analysis

### Column layout (§2.1)

| Columns | Area | Handling |
|---|---|---|
| 1–6 | Sequence | Ignored. |
| 7 | Indicator | `*` or `/` = comment · `-` = continuation · `D` = debugging · space = normal line. Any other character is an error. |
| 8–72 | Areas A and B | Read as code. |
| 73–80 | Program | Ignored. |

Continuation applies to **literals** and to **words** split at the end of column 72 (`T` + `-AB-TRABALHADOR` = `TAB-TRABALHADOR`).

### Token categories

`KEYWORD`, `IDENT`, `NUMBER`, `STRING`, `PICTURE-STRING`, `OPERATOR`, `PERIOD`, `COMMA`, `SEMICOLON`, `LPAREN`, `RPAREN` and `EOF`. Verbs and clauses the specification does not list (`ADD`, `INSPECT`, `SEARCH`…) come out as `IDENT` and are recognized by the parser.

### Lexical errors

All `E`, formatted as `[E] linha L, col C: message`: invalid character, `_` outside the program name, unterminated literal, continuation without a quote, identifier over 30 characters or with a leading/trailing hyphen, number with two decimal points, invalid column-7 indicator and **git merge conflict marker**. The free text of the `AUTHOR`, `INSTALLATION`, `DATE-WRITTEN`, `DATE-COMPILED`, `SECURITY` and `REMARKS` paragraphs (IDENTIFICATION) does not raise invalid-character errors.

---

## Syntax and semantic analysis

### Syntax

Recognizes `IDENTIFICATION` (including `AUTHOR`, `REMARKS`…), `ENVIRONMENT` (`SPECIAL-NAMES`, `SELECT`), `DATA` (`FD`, levels `01`–`49` and `77`, `PIC`, `VALUE(S)`, `OCCURS`, `INDEXED BY`, `USAGE`) and `PROCEDURE` with sections and paragraphs. Statements: `DISPLAY`, `ACCEPT`, `MOVE`, `ADD`, `SUBTRACT`, `MULTIPLY`, `DIVIDE`, `COMPUTE` (with `GIVING`, `REMAINDER`, `ON SIZE ERROR`), `IF/ELSE` (with or without `END-IF`), `EVALUATE`, `PERFORM` (plain, `UNTIL`, `THRU`), `GO TO`, `SET`, `SEARCH`, `INSPECT`, `OPEN`, `CLOSE`, `READ`, `WRITE`, `START`, `STOP RUN`, `GOBACK`, `EXIT`, `CONTINUE` and `FUNCTION UPPER-CASE`. Expressions with `**`, `* /`, `+ -`; conditions with `AND/OR/NOT`, symbolic or word relations (`EQUAL`, `GREATER`…), abbreviated conditions and `IS NUMERIC`.

### Semantics (§4)

| Check | Example message |
|---|---|
| `MOVE` from `PIC A` to `PIC 9` | `MOVE inválido: item alfabético (PIC A) não pode ser movido para o item numérico 'NUM'` |
| Duplicate identifier, ambiguous reference (resolved with `OF`) | `identificador 'CONT' declarado mais de uma vez no mesmo escopo` · `referência ambígua a 'CAMPO'; qualifique com OF` |
| Missing `PERFORM`/`GO TO` target | `parágrafo ou seção 'FALTA' não declarado na PROCEDURE DIVISION` |
| File operation without `SELECT` and `FD` | `arquivo 'X' não declarado em FILE-CONTROL (SELECT)` · `arquivo 'Y' sem descritor FD na FILE SECTION` |

It also checks undeclared identifiers, missing or extra subscripts, `SEARCH` without `INDEXED BY` and `VALUE` incompatible with the `PIC`.

---

## MIPS generation

- **Data:** one static block whose initial image is computed at compile time (`VALUE`, zeros for numerics, spaces for text). Groups, `OCCURS` and subscripts become `base + (i-1)*size`.
- **Numbers:** kept as ASCII digits, as in COBOL (`DISPLAY` usage), with sign overpunch. Arithmetic works on 40-digit decimal vectors: add, subtract, multiply, long (truncating) division, power with an integer exponent. No floating point, no 32-bit limit.
- **`PERFORM`:** software return stack; every paragraph or section end checks the top of the stack. Works with recursion and sections.
- **`ON SIZE ERROR`:** target untouched on overflow; without the clause, high-order digits are truncated.
- **Edited pictures:** `PIC -ZZZZZZZZZ9.99`, `ZZZ.ZZZ.ZZZ,ZZ` (with `DECIMAL-POINT IS COMMA`), `Z`, `*`, `9`, `+`, `-`, `CR`, `DB`, `B`, `0`, `/`, `$` and separators.
- **Files:** `SEQUENTIAL` (fixed records), `LINE SEQUENTIAL` and `INDEXED` (in-memory table looked up by key, written on `CLOSE`), with `FILE STATUS` 00, 10, 22, 23, 34, 35, 37, 41, 42, 47 and 48.
- **`INSPECT`:** `TALLYING`, `REPLACING` and `CONVERTING`, with `BEFORE`/`AFTER INITIAL`.

---

## Design decisions

| Decision | Reason |
|---|---|
| Implementation in **Go** | Performance. |
| **MIPS (MARS)** backend | Professor's recommendation, for simplicity. Replaces the C target in §5. |
| Hand-written lexer and parser | Many rules depend on columns and context. |
| COBOL words are contextual in the parser | Keeps the lexer stable and avoids reserving names such as `NAME`. |
| `_` only in the program name | §2.2 does not allow `_`, but three examples use it in `PROGRAM-ID`. |
| `OPERATOR` category | §2.2 only lists `. , ; ( )`, but the examples use `=`, `>`, `**`. |
| Word continuation | Without it, `TAB-TRABALHADOR` (`prog004`/`prog006`) would not exist. |
| Case-insensitive | COBOL standard; literals are preserved. |
| `D` lines ignored; `-debug` includes them | COBOL standard. |
| **GnuCOBOL** reference outputs (`-std=ibm`) | Source of truth for signed `DISPLAY`, numeric `ACCEPT` and files. |
| `CAD0002A`, `CAD0005A`, `PROG024A` rejected | `COPY` of missing copybooks and `CALL 'PROGDATA'`; merge markers in the source. |

---

## Tests

Validation programs live in `tests/input_programs/`: the professor's 7 (from [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos)) and 7 `.cbl` files. Inputs are in `tests/stdin/`, fixtures in `tests/data/`, expected outputs (produced by GnuCOBOL) in `tests/expected/`.

| Script | What it does |
|---|---|
| `scripts/run-mips-tests.sh [program]` | Compiles each program in `tests/programs.txt`, runs it on MARS and compares the output (and generated files) with `tests/expected/`. The `.asm` files go to `tests/asm/`. |
| `scripts/run-tests.sh` | Runs the compiler on each program and writes the report to `tests/outputs/`. |
| `scripts/run-error-tests.sh` | Checks the inserted lexical errors (`tests/error_programs/`), the semantic errors (`tests/semantic_programs/`) and the rejection of the 3 programs in `tests/rejected.txt`. |
| `scripts/gen-reference.sh` | Regenerates `tests/expected/` with `cobc`. |

Current result: **11/11** programs with output identical to GnuCOBOL; **7/7** programs with lexical errors detected; **11/11** expected rejections.

---

## Limitations

- `**` only with an integer exponent.
- `PERFORM VARYING`, inline `PERFORM`, `INITIALIZE`, `CALL`, reference modification (`X(1:2)`), `REDEFINES`, `COMP`/`COMP-3` and `SCREEN SECTION` are not supported and give a clear error.
- The indexed file is simulated (up to 256 records) in its own format, not ISAM-compatible.
- After an expanded `COPY`, line numbers in diagnostics are shifted.
- MARS blocks a file name after a failed read `OPEN`; the runtime works around it by prefixing `./` to the path.

---

## Folder layout

```
.
├── cmd/cobolc/              # executable
├── internal/
│   ├── ast/  codegen/mips/  diagnostic/  lexer/  parser/
│   ├── preprocess/  report/  semantic/  source/  token/
│   └── codegen/mips/runtime/   # MIPS routines (core, str, num, edit, inspect, file)
├── scripts/                 # run-mips-tests, run-tests, run-error-tests, gen-reference
├── examples/hello.cbl
└── tests/
    ├── input_programs/  stdin/  data/  expected/  asm/  outputs/
    └── error_programs/  error_outputs/  semantic_programs/  rejected.txt  programs.txt
```

---

## License

Distributed under the MIT license. See [LICENSE](LICENSE). The programs in `tests/input_programs/` come from [ebasso/cobol-exemplos](https://github.com/ebasso/cobol-exemplos) and keep their original license (`tests/LICENSE`).
