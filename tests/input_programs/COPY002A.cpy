       01  COPY002A-REGISTRO.
           05 COPY002A-QUANT-REG           PIC 9(02) VALUE ZEROS.
           05 COPY002A-TABELA OCCURS 50 TIMES.
              10 COPY002A-COD              PIC 9(02).
              10 COPY002A-NOME             PIC X(30).
              10 COPY002A-IDADE            PIC 9(02).
              10 COPY002A-DATA-NASC        PIC 9(08).
              10 COPY002A-CARGO            PIC X(20).
              10 COPY002A-EMAIL            PIC X(50).
              10 COPY002A-TELEFONE         PIC 9(09).
              10 COPY002A-RUA              PIC X(50).
              10 COPY002A-CIDADE           PIC X(30).
              10 COPY002A-ESTADO           PIC X(02).
              10 COPY002A-CEP              PIC 9(08).
