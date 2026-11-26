# ===== edit =====
        .text

# rt_edit(a0=num, a1=dst, a2=mascara, a3=tamanho | (virgula decimal << 16))
rt_edit:
        addiu $sp, $sp, -48
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        sw   $a3, 16($sp)
        li   $t8, 46
        li   $t9, 44
        srl  $t0, $a3, 16
        andi $t0, $t0, 1
        beqz $t0, re_1
        li   $t8, 44
        li   $t9, 46
re_1:
        sw   $t8, 20($sp)
        sw   $t9, 24($sp)
        andi $t0, $a3, 65535
        sw   $t0, 28($sp)
        li   $t1, 0
        li   $t2, 0
        li   $t3, 0
        li   $t4, 0
re_p1:
        bge  $t1, $t0, re_p1d
        addu $t5, $a2, $t1
        lbu  $t5, 0($t5)
        li   $t6, 57
        beq  $t5, $t6, re_dg
        li   $t6, 90
        beq  $t5, $t6, re_dg
        li   $t6, 42
        beq  $t5, $t6, re_dg
        j    re_nd
re_dg:
        bnez $t4, re_df
        addiu $t2, $t2, 1
        j    re_nd
re_df:
        addiu $t3, $t3, 1
re_nd:
        bne  $t5, $t8, re_nx
        li   $t4, 1
re_nx:
        addiu $t1, $t1, 1
        j    re_p1
re_p1d:
        addu $t6, $t2, $t3
        sw   $t6, 32($sp)
        lw   $a0, 4($sp)
        la   $a1, RT_EDITB
        move $a2, $t6
        move $a3, $t3
        jal  rt_store
        lw   $t0, 4($sp)
        lbu  $t7, 0($t0)
        beqz $t7, re_ng
        li   $t1, 0
        lw   $t6, 32($sp)
re_nz:
        bge  $t1, $t6, re_nz0
        la   $t2, RT_EDITB
        addu $t2, $t2, $t1
        lbu  $t3, 0($t2)
        li   $t4, 48
        bne  $t3, $t4, re_ng
        addiu $t1, $t1, 1
        j    re_nz
re_nz0:
        li   $t7, 0
re_ng:
        sw   $t7, 36($sp)
        li   $t1, 0
        li   $t2, 0
        li   $t3, 0
        lw   $t0, 28($sp)
re_w:
        bge  $t1, $t0, re_done
        lw   $a2, 12($sp)
        addu $t5, $a2, $t1
        lbu  $t5, 0($t5)
        lw   $t8, 20($sp)
        lw   $t9, 24($sp)
        lw   $a1, 8($sp)
        addu $t6, $a1, $t1
        li   $t7, 57
        beq  $t5, $t7, re_9
        li   $t7, 90
        beq  $t5, $t7, re_z
        li   $t7, 42
        beq  $t5, $t7, re_s
        beq  $t5, $t8, re_dp
        beq  $t5, $t9, re_sep
        li   $t7, 45
        beq  $t5, $t7, re_mi
        li   $t7, 43
        beq  $t5, $t7, re_pl
        li   $t7, 67
        beq  $t5, $t7, re_c
        li   $t7, 82
        beq  $t5, $t7, re_r
        li   $t7, 68
        beq  $t5, $t7, re_d
        li   $t7, 66
        beq  $t5, $t7, re_b
        move $t4, $t5
        j    re_out
re_9:
        la   $t4, RT_EDITB
        addu $t4, $t4, $t2
        lbu  $t4, 0($t4)
        addiu $t2, $t2, 1
        li   $t3, 1
        j    re_out
re_z:
        li   $t9, 32
        j    re_sup
re_s:
        li   $t9, 42
re_sup:
        la   $t4, RT_EDITB
        addu $t4, $t4, $t2
        lbu  $t4, 0($t4)
        addiu $t2, $t2, 1
        li   $t7, 48
        beq  $t4, $t7, re_sup2
        li   $t3, 1
re_sup2:
        bnez $t3, re_out
        move $t4, $t9
        j    re_out
re_dp:
        move $t4, $t5
        li   $t3, 1
        j    re_out
re_sep:
        li   $t4, 32
        beqz $t3, re_out
        move $t4, $t5
        j    re_out
re_mi:
        lw   $t7, 36($sp)
        li   $t4, 32
        beqz $t7, re_out
        li   $t4, 45
        j    re_out
re_pl:
        lw   $t7, 36($sp)
        li   $t4, 43
        beqz $t7, re_out
        li   $t4, 45
        j    re_out
re_c:
re_r:
re_d:
        lw   $t7, 36($sp)
        li   $t4, 32
        beqz $t7, re_out
        move $t4, $t5
        j    re_out
re_b:
        li   $t4, 32
        beqz $t1, re_out
        addu $t7, $a2, $t1
        lbu  $t7, -1($t7)
        li   $t8, 68
        bne  $t7, $t8, re_out
        lw   $t7, 36($sp)
        beqz $t7, re_out
        li   $t4, 66
re_out:
        sb   $t4, 0($t6)
        addiu $t1, $t1, 1
        j    re_w
re_done:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 48
        jr   $ra
