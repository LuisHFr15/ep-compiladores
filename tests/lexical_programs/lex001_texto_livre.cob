       IDENTIFICATION DIVISION.
       PROGRAM-ID. LEX001.
       AUTHOR. Fulano: de Tal @ $ & #.
       REMARKS. Observacao: texto livre; com simbolos : @ $ ?
                 e continua na linha seguinte.
       DATA DIVISION.
       WORKING-STORAGE SECTION.
       77 A PIC 9 VALUE 1.
       PROCEDURE DIVISION.
       INICIO.
           DISPLAY "A:B " A.
           STOP RUN.
