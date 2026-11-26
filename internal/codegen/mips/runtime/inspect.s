# ===== inspect =====
# RT_INS: 0 alvo, 4 tam, 8 tipo, 12 padrao, 16 tam padrao, 20 substituto, 24 tam substituto,
#         28 after, 32 tam after (-1 ausente), 36 before, 40 tam before (-1 ausente)
        .data
        .align 2
RT_INS: .space 48
        .text

# ri_match(a0=p, a1=q, a2=len) -> v0 = 1 se iguais
ri_match:
        li   $t0, 0
rim_l:
        bge  $t0, $a2, rim_y
        addu $t1, $a0, $t0
        lbu  $t2, 0($t1)
        addu $t1, $a1, $t0
        lbu  $t3, 0($t1)
        bne  $t2, $t3, rim_n
        addiu $t0, $t0, 1
        j    rim_l
rim_y:
        li   $v0, 1
        jr   $ra
rim_n:
        li   $v0, 0
        jr   $ra

# ri_find(a0=base, a1=de, a2=ate, a3=padrao, v1=tam) -> v0 = posicao ou -1
ri_find:
        addiu $sp, $sp, -28
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        sw   $a3, 16($sp)
        sw   $v1, 20($sp)
rif_l:
        lw   $t0, 8($sp)
        lw   $t1, 12($sp)
        lw   $t3, 20($sp)
        addu $t2, $t0, $t3
        bgt  $t2, $t1, rif_nf
        lw   $a0, 4($sp)
        addu $a0, $a0, $t0
        lw   $a1, 16($sp)
        move $a2, $t3
        jal  ri_match
        bnez $v0, rif_f
        lw   $t0, 8($sp)
        addiu $t0, $t0, 1
        sw   $t0, 8($sp)
        j    rif_l
rif_f:
        lw   $v0, 8($sp)
        j    rif_r
rif_nf:
        li   $v0, -1
rif_r:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 28
        jr   $ra

# rt_inspect() -> v0 = contagem (tipos 1-3)
# tipos: 1 TALLY CHARACTERS, 2 TALLY ALL, 3 TALLY LEADING, 4 REPL CHARACTERS,
#        5 REPL ALL, 6 REPL FIRST, 7 REPL LEADING, 8 CONVERTING
rt_inspect:
        addiu $sp, $sp, -48
        sw   $ra, 0($sp)
        sw   $s1, 4($sp)
        sw   $s2, 8($sp)
        sw   $s3, 12($sp)
        sw   $s4, 16($sp)
        sw   $s5, 20($sp)
        sw   $s6, 24($sp)
        sw   $s7, 28($sp)
        la   $t9, RT_INS
        lw   $s1, 0($t9)
        lw   $s2, 4($t9)
        lw   $s5, 12($t9)
        lw   $s6, 16($t9)
        li   $s3, 0
        move $s4, $s2
        li   $s7, 0
        lw   $t0, 32($t9)
        bltz $t0, ri_nb
        move $a0, $s1
        li   $a1, 0
        move $a2, $s2
        lw   $a3, 28($t9)
        move $v1, $t0
        jal  ri_find
        bltz $v0, ri_noaf
        la   $t9, RT_INS
        lw   $t0, 32($t9)
        addu $s3, $v0, $t0
        j    ri_nb
ri_noaf:
        move $s3, $s2
        move $s4, $s2
ri_nb:
        la   $t9, RT_INS
        lw   $t0, 40($t9)
        bltz $t0, ri_k
        move $a0, $s1
        move $a1, $s3
        move $a2, $s2
        lw   $a3, 36($t9)
        move $v1, $t0
        jal  ri_find
        bltz $v0, ri_k
        move $s4, $v0
ri_k:
        bge  $s4, $s3, ri_k2
        move $s4, $s3
ri_k2:
        la   $t9, RT_INS
        lw   $t0, 8($t9)
        sw   $t0, 36($sp)
        li   $t1, 1
        beq  $t0, $t1, ri_k1
        li   $t1, 4
        beq  $t0, $t1, ri_k4
        li   $t1, 8
        beq  $t0, $t1, ri_k8
        sw   $s3, 32($sp)
ri_sc:
        lw   $t0, 32($sp)
        addu $t1, $t0, $s6
        bgt  $t1, $s4, ri_done
        addu $a0, $s1, $t0
        move $a1, $s5
        move $a2, $s6
        jal  ri_match
        bnez $v0, ri_hit
        lw   $t2, 36($sp)
        li   $t3, 3
        beq  $t2, $t3, ri_done
        li   $t3, 7
        beq  $t2, $t3, ri_done
        lw   $t0, 32($sp)
        addiu $t0, $t0, 1
        sw   $t0, 32($sp)
        j    ri_sc
ri_hit:
        lw   $t2, 36($sp)
        li   $t3, 5
        blt  $t2, $t3, ri_cnt
        la   $t9, RT_INS
        lw   $t4, 20($t9)
        lw   $t5, 24($t9)
        move $t6, $s6
        ble  $t6, $t5, ri_cp
        move $t6, $t5
ri_cp:
        lw   $t0, 32($sp)
        li   $t7, 0
ri_cl:
        bge  $t7, $t6, ri_ad
        addu $t8, $t4, $t7
        lbu  $t8, 0($t8)
        addu $t1, $t0, $t7
        addu $t1, $s1, $t1
        sb   $t8, 0($t1)
        addiu $t7, $t7, 1
        j    ri_cl
ri_cnt:
        addiu $s7, $s7, 1
ri_ad:
        lw   $t0, 32($sp)
        addu $t0, $t0, $s6
        sw   $t0, 32($sp)
        lw   $t2, 36($sp)
        li   $t3, 6
        bne  $t2, $t3, ri_sc
        j    ri_done
ri_k1:
        subu $s7, $s4, $s3
        j    ri_done
ri_k4:
        la   $t9, RT_INS
        lw   $t4, 20($t9)
        lbu  $t4, 0($t4)
        move $t0, $s3
ri_k4l:
        bge  $t0, $s4, ri_done
        addu $t1, $s1, $t0
        sb   $t4, 0($t1)
        addiu $t0, $t0, 1
        j    ri_k4l
ri_k8:
        la   $t9, RT_INS
        lw   $t4, 20($t9)
        move $t0, $s3
ri_k8l:
        bge  $t0, $s4, ri_done
        addu $t1, $s1, $t0
        lbu  $t2, 0($t1)
        li   $t3, 0
ri_k8k:
        bge  $t3, $s6, ri_k8n
        addu $t5, $s5, $t3
        lbu  $t5, 0($t5)
        bne  $t5, $t2, ri_k8m
        addu $t5, $t4, $t3
        lbu  $t5, 0($t5)
        sb   $t5, 0($t1)
        j    ri_k8n
ri_k8m:
        addiu $t3, $t3, 1
        j    ri_k8k
ri_k8n:
        addiu $t0, $t0, 1
        j    ri_k8l
ri_done:
        move $v0, $s7
        lw   $ra, 0($sp)
        lw   $s1, 4($sp)
        lw   $s2, 8($sp)
        lw   $s3, 12($sp)
        lw   $s4, 16($sp)
        lw   $s5, 20($sp)
        lw   $s6, 24($sp)
        lw   $s7, 28($sp)
        addiu $sp, $sp, 48
        jr   $ra
