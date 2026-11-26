# ===== str =====
        .text

# rt_move_alnum(a0=dst, a1=dlen, a2=src, a3=slen)
rt_move_alnum:
        move $t3, $a1
        ble  $a1, $a3, rma_n
        move $t3, $a3
rma_n:
        sltu $t0, $a0, $a2
        beqz $t0, rma_back
        li   $t0, 0
rma_f:
        bge  $t0, $t3, rma_pad
        addu $t1, $a2, $t0
        lbu  $t2, 0($t1)
        addu $t1, $a0, $t0
        sb   $t2, 0($t1)
        addiu $t0, $t0, 1
        j    rma_f
rma_back:
        move $t0, $t3
rma_b:
        blez $t0, rma_pad
        addiu $t0, $t0, -1
        addu $t1, $a2, $t0
        lbu  $t2, 0($t1)
        addu $t1, $a0, $t0
        sb   $t2, 0($t1)
        j    rma_b
rma_pad:
        move $t0, $t3
        li   $t2, 32
rma_p:
        bge  $t0, $a1, rma_e
        addu $t1, $a0, $t0
        sb   $t2, 0($t1)
        addiu $t0, $t0, 1
        j    rma_p
rma_e:
        jr   $ra

# rt_fill(a0=dst, a1=len, a2=byte)
rt_fill:
        li   $t0, 0
rf_l:
        bge  $t0, $a1, rf_e
        addu $t1, $a0, $t0
        sb   $a2, 0($t1)
        addiu $t0, $t0, 1
        j    rf_l
rf_e:
        jr   $ra

# rt_cmp_alnum(a0=p1, a1=l1, a2=p2, a3=l2) -> v0 = -1/0/1 (menor com espacos)
rt_cmp_alnum:
        li   $t0, 0
        move $t4, $a1
        bge  $a1, $a3, rca_1
        move $t4, $a3
rca_1:
        bge  $t0, $t4, rca_eq
        li   $t5, 32
        li   $t6, 32
        bge  $t0, $a1, rca_2
        addu $t1, $a0, $t0
        lbu  $t5, 0($t1)
rca_2:
        bge  $t0, $a3, rca_3
        addu $t1, $a2, $t0
        lbu  $t6, 0($t1)
rca_3:
        bne  $t5, $t6, rca_d
        addiu $t0, $t0, 1
        j    rca_1
rca_d:
        sltu $t7, $t5, $t6
        li   $v0, 1
        beqz $t7, rca_e
        li   $v0, -1
rca_e:
        jr   $ra
rca_eq:
        li   $v0, 0
        jr   $ra

# rt_cmp_fill(a0=p, a1=len, a2=byte) -> v0 = -1/0/1
rt_cmp_fill:
        li   $t0, 0
rcf_l:
        bge  $t0, $a1, rcf_eq
        addu $t1, $a0, $t0
        lbu  $t5, 0($t1)
        bne  $t5, $a2, rcf_d
        addiu $t0, $t0, 1
        j    rcf_l
rcf_d:
        sltu $t7, $t5, $a2
        li   $v0, 1
        beqz $t7, rcf_e
        li   $v0, -1
rcf_e:
        jr   $ra
rcf_eq:
        li   $v0, 0
        jr   $ra

# rt_upper(a0=dst, a1=src, a2=len)
rt_upper:
        li   $t0, 0
ru_l:
        bge  $t0, $a2, ru_e
        addu $t1, $a1, $t0
        lbu  $t2, 0($t1)
        li   $t3, 97
        blt  $t2, $t3, ru_s
        li   $t3, 122
        bgt  $t2, $t3, ru_s
        addiu $t2, $t2, -32
ru_s:
        addu $t1, $a0, $t0
        sb   $t2, 0($t1)
        addiu $t0, $t0, 1
        j    ru_l
ru_e:
        jr   $ra

# rt_accept_alnum(a0=dst, a1=len)
rt_accept_alnum:
        addiu $sp, $sp, -12
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        jal  rt_readline
        move $a3, $v0
        lw   $a0, 4($sp)
        lw   $a1, 8($sp)
        la   $a2, RT_INBUF
        jal  rt_move_alnum
        lw   $ra, 0($sp)
        addiu $sp, $sp, 12
        jr   $ra

# rt_is_numeric(a0=addr, a1=len) -> v0 = 1 se todos os bytes sao digitos
rt_is_numeric:
        li   $v0, 0
        blez $a1, rin_e
        li   $t0, 0
rin_l:
        bge  $t0, $a1, rin_ok
        addu $t1, $a0, $t0
        lbu  $t2, 0($t1)
        li   $t3, 48
        blt  $t2, $t3, rin_e
        li   $t3, 57
        bgt  $t2, $t3, rin_e
        addiu $t0, $t0, 1
        j    rin_l
rin_ok:
        li   $v0, 1
rin_e:
        jr   $ra
