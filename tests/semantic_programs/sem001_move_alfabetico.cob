       IDENTIFICATION DIVISION.
       PROGRAM-ID. SEM001.
       DATA DIVISION.
       WORKING-STORAGE SECTION.
       77 NOME            PIC A(10) VALUE "ABC".
       77 NUM             PIC 9(5)  VALUE 0.
       PROCEDURE DIVISION.
       INICIO.
           MOVE NOME TO NUM.
           STOP RUN.
