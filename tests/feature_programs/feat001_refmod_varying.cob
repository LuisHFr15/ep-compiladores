       IDENTIFICATION DIVISION.
       PROGRAM-ID. FEAT001.
       DATA DIVISION.
       WORKING-STORAGE SECTION.
       77 TXT       PIC X(10) VALUE "ABCDEFGHIJ".
       77 NUM       PIC 9(8)  VALUE 20250131.
       77 I         PIC 9(2).
       77 SOMA      PIC 9(3)  VALUE ZEROS.
       77 CAR        PIC X(1).
       77 IDADE     PIC 9(3)  VALUE ZEROS.
       01 GRP.
          05 G-A    PIC 9(2) VALUE 11.
          05 G-B    PIC X(3) VALUE "XYZ".
          05 FILLER PIC X(1) VALUE "-".
          05 G-C    PIC 9(2) VALUE 22.
       PROCEDURE DIVISION.
       INICIO.
           DISPLAY TXT(3:2) "|" TXT(8:) "|" TXT(1:1).
           MOVE 4 TO I.
           DISPLAY TXT(I:3).
           MOVE "ZZ" TO TXT(2:2).
           DISPLAY TXT.
           MOVE NUM(1:4) TO TXT(1:4).
           DISPLAY TXT.
           MOVE 99 TO NUM(5:2).
           DISPLAY NUM.
           MOVE "AB1CD2EF3G" TO TXT.
           PERFORM VARYING I FROM 1 BY 1 UNTIL I > LENGTH OF TXT
               MOVE TXT(I:1) TO CAR
               IF CAR IS NUMERIC
                   ADD 1 TO SOMA
               END-IF
           END-PERFORM.
           DISPLAY "DIGITOS=" SOMA.
           PERFORM UNTIL SOMA = 0
               SUBTRACT 1 FROM SOMA
           END-PERFORM.
           DISPLAY "SOMA=" SOMA.
           MOVE LENGTH OF GRP TO I.
           DISPLAY "LEN=" I.
           INITIALIZE GRP REPLACING ALPHANUMERIC BY "AAA"
                                    NUMERIC BY 7.
           DISPLAY GRP.
           INITIALIZE GRP.
           DISPLAY GRP.
           MOVE "ABC" TO TXT.
           INSPECT FUNCTION REVERSE(TXT) TALLYING SOMA
                   FOR LEADING SPACES.
           DISPLAY "SOMA=" SOMA.
           MOVE 30 TO IDADE.
           EVALUATE IDADE
              WHEN LESS 18
                 DISPLAY "MENOR"
              WHEN GREATER 65
                 DISPLAY "IDOSO"
              WHEN IS NOT NUMERIC
                 DISPLAY "NAN"
              WHEN ZEROS
                 DISPLAY "ZERO"
              WHEN OTHER
                 DISPLAY "ADULTO"
           END-EVALUATE.
           STOP RUN.
