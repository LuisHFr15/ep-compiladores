       IDENTIFICATION DIVISION.
       PROGRAM-ID. SEM002.
       DATA DIVISION.
       WORKING-STORAGE SECTION.
       77 CONT            PIC 9.
       77 CONT            PIC 9.
       01 GRUPO-1.
          05 CAMPO        PIC 9.
       01 GRUPO-2.
          05 CAMPO        PIC 9.
       PROCEDURE DIVISION.
       INICIO.
           MOVE 1 TO CAMPO.
           MOVE 1 TO CAMPO OF GRUPO-1.
           STOP RUN.
