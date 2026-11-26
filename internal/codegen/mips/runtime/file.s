# ===== file =====
# FCB: 0 fd, 4 modo (0 fechado,1 input,2 output,3 i-o,4 extend), 8 status, 12 org (1 seq,2 linha,3 indexado),
#      16 tam registro, 20 area do registro, 24 deslocamento da chave, 28 tam da chave, 32 tabela,
#      36 qtd, 40 posicao atual, 44 sujo, 64 caminho (128 bytes)
        .data
        .align 2
RT_F1:  .space 8
RT_FNL: .byte 10
        .text

# rt_f_open(a0=fcb, a1=nome, a2=tam nome, a3=modo)
rt_f_open:
        addiu $sp, $sp, -16
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a3, 8($sp)
        lw   $t0, 4($a0)
        beqz $t0, rfo_ok
        li   $t1, 41
        sw   $t1, 8($a0)
        j    rfo_ret
rfo_ok:
        move $t0, $a2
rfo_trim:
        blez $t0, rfo_cp
        addu $t1, $a1, $t0
        lbu  $t1, -1($t1)
        li   $t2, 32
        bne  $t1, $t2, rfo_cp
        addiu $t0, $t0, -1
        j    rfo_trim
rfo_cp:
        addiu $t3, $a0, 64
        lw   $t7, 48($a0)
rfo_pf:
        blez $t7, rfo_pd
        li   $t8, 46
        sb   $t8, 0($t3)
        li   $t8, 47
        sb   $t8, 1($t3)
        addiu $t3, $t3, 2
        addiu $t7, $t7, -1
        j    rfo_pf
rfo_pd:
        li   $t4, 0
rfo_cl:
        bge  $t4, $t0, rfo_cd
        addu $t5, $a1, $t4
        lbu  $t5, 0($t5)
        addu $t6, $t3, $t4
        sb   $t5, 0($t6)
        addiu $t4, $t4, 1
        j    rfo_cl
rfo_cd:
        addu $t6, $t3, $t4
        sb   $zero, 0($t6)
        li   $a1, 0
        li   $t5, 2
        beq  $a3, $t5, rfo_fo
        li   $t5, 4
        bne  $a3, $t5, rfo_fl
        li   $a1, 9
        j    rfo_op
rfo_fo:
        li   $a1, 1
rfo_fl:
rfo_op:
        addiu $a0, $a0, 64
        li   $a2, 0
        li   $v0, 13
        syscall
        lw   $a0, 4($sp)
        lw   $a3, 8($sp)
        bltz $v0, rfo_err
        sw   $v0, 0($a0)
        sw   $a3, 4($a0)
        sw   $zero, 8($a0)
        lw   $t0, 12($a0)
        li   $t1, 3
        bne  $t0, $t1, rfo_ret
        li   $t2, 2
        beq  $a3, $t2, rfo_idxo
        sw   $zero, 36($a0)
        sw   $zero, 40($a0)
        sw   $zero, 44($a0)
rfo_ld:
        lw   $t3, 36($a0)
        lw   $t4, 16($a0)
        mul  $t5, $t3, $t4
        lw   $t6, 32($a0)
        addu $a1, $t6, $t5
        lw   $a2, 16($a0)
        lw   $a0, 0($a0)
        li   $v0, 14
        syscall
        lw   $a0, 4($sp)
        lw   $t4, 16($a0)
        bne  $v0, $t4, rfo_ldd
        lw   $t3, 36($a0)
        addiu $t3, $t3, 1
        sw   $t3, 36($a0)
        li   $t5, 256
        blt  $t3, $t5, rfo_ld
rfo_ldd:
        lw   $a0, 0($a0)
        li   $v0, 16
        syscall
        lw   $a0, 4($sp)
        li   $t0, -1
        sw   $t0, 0($a0)
        j    rfo_ret
rfo_idxo:
        sw   $zero, 36($a0)
        sw   $zero, 40($a0)
        li   $t3, 1
        sw   $t3, 44($a0)
        lw   $a0, 0($a0)
        li   $v0, 16
        syscall
        lw   $a0, 4($sp)
        li   $t0, -1
        sw   $t0, 0($a0)
        j    rfo_ret
rfo_err:
        li   $t0, 35
        li   $t1, 2
        beq  $a3, $t1, rfo_e37
        li   $t1, 4
        bne  $a3, $t1, rfo_es
rfo_e37:
        li   $t0, 37
rfo_es:
        sw   $t0, 8($a0)
        lw   $t2, 48($a0)
        addiu $t2, $t2, 1
        sw   $t2, 48($a0)
rfo_ret:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 16
        jr   $ra

# rt_f_close(a0=fcb)
rt_f_close:
        lw   $t0, 4($a0)
        bnez $t0, rfc_ok
        li   $t1, 42
        sw   $t1, 8($a0)
        jr   $ra
rfc_ok:
        lw   $t0, 12($a0)
        li   $t1, 3
        bne  $t0, $t1, rfc_cl
        lw   $t2, 44($a0)
        beqz $t2, rfc_end
        move $t9, $a0
        addiu $a0, $t9, 64
        li   $a1, 1
        li   $a2, 0
        li   $v0, 13
        syscall
        move $t8, $v0
        bltz $t8, rfc_end2
        move $a0, $t8
        lw   $a1, 32($t9)
        lw   $t3, 36($t9)
        lw   $t4, 16($t9)
        mul  $a2, $t3, $t4
        li   $v0, 15
        syscall
        move $a0, $t8
        li   $v0, 16
        syscall
rfc_end2:
        move $a0, $t9
        j    rfc_end
rfc_cl:
        move $t9, $a0
        lw   $a0, 0($t9)
        li   $v0, 16
        syscall
        move $a0, $t9
rfc_end:
        li   $t0, -1
        sw   $t0, 0($a0)
        sw   $zero, 4($a0)
        sw   $zero, 8($a0)
        sw   $zero, 44($a0)
        jr   $ra

# rt_f_read(a0=fcb): leitura sequencial
rt_f_read:
        lw   $t0, 4($a0)
        li   $t1, 1
        beq  $t0, $t1, rfr_m
        li   $t1, 3
        beq  $t0, $t1, rfr_m
        li   $t1, 47
        sw   $t1, 8($a0)
        jr   $ra
rfr_m:
        lw   $t0, 12($a0)
        li   $t1, 2
        beq  $t0, $t1, rfr_line
        li   $t1, 3
        beq  $t0, $t1, rfr_idx
        move $t9, $a0
        lw   $a0, 0($t9)
        lw   $a1, 20($t9)
        lw   $a2, 16($t9)
        li   $v0, 14
        syscall
        move $a0, $t9
        blez $v0, rfr_eof
        lw   $t2, 16($a0)
        bge  $v0, $t2, rfr_ok
        lw   $t3, 20($a0)
        li   $t4, 32
rfr_pf:
        bge  $v0, $t2, rfr_ok
        addu $t5, $t3, $v0
        sb   $t4, 0($t5)
        addiu $v0, $v0, 1
        j    rfr_pf
rfr_eof:
        li   $t1, 10
        sw   $t1, 8($a0)
        jr   $ra
rfr_ok:
        sw   $zero, 8($a0)
        jr   $ra
rfr_line:
        move $t9, $a0
        li   $t6, 0
        li   $t7, 0
        lw   $t8, 16($t9)
rfr_ll:
        lw   $a0, 0($t9)
        la   $a1, RT_F1
        li   $a2, 1
        li   $v0, 14
        syscall
        blez $v0, rfr_le
        la   $t0, RT_F1
        lbu  $t0, 0($t0)
        li   $t1, 10
        beq  $t0, $t1, rfr_pad
        li   $t1, 13
        beq  $t0, $t1, rfr_ll
        li   $t7, 1
        bge  $t6, $t8, rfr_ll
        lw   $t2, 20($t9)
        addu $t2, $t2, $t6
        sb   $t0, 0($t2)
        addiu $t6, $t6, 1
        j    rfr_ll
rfr_le:
        bnez $t7, rfr_pad
        move $a0, $t9
        j    rfr_eof
rfr_pad:
        lw   $t2, 20($t9)
        li   $t4, 32
rfr_lp:
        bge  $t6, $t8, rfr_ld
        addu $t5, $t2, $t6
        sb   $t4, 0($t5)
        addiu $t6, $t6, 1
        j    rfr_lp
rfr_ld:
        move $a0, $t9
        sw   $zero, 8($a0)
        jr   $ra
rfr_idx:
        lw   $t0, 40($a0)
        lw   $t1, 36($a0)
        blt  $t0, $t1, rfr_ix
        j    rfr_eof
rfr_ix:
        lw   $t2, 16($a0)
        mul  $t3, $t0, $t2
        lw   $t4, 32($a0)
        addu $t4, $t4, $t3
        lw   $t5, 20($a0)
        li   $t6, 0
rfr_ic:
        bge  $t6, $t2, rfr_id
        addu $t7, $t4, $t6
        lbu  $t7, 0($t7)
        addu $t8, $t5, $t6
        sb   $t7, 0($t8)
        addiu $t6, $t6, 1
        j    rfr_ic
rfr_id:
        addiu $t0, $t0, 1
        sw   $t0, 40($a0)
        sw   $zero, 8($a0)
        jr   $ra

# rf_find(a0=fcb) -> v0 = indice do registro com a chave igual a da area do registro, ou -1
rf_find:
        lw   $t0, 36($a0)
        lw   $t1, 16($a0)
        lw   $t2, 32($a0)
        lw   $t3, 20($a0)
        lw   $t4, 24($a0)
        lw   $t5, 28($a0)
        li   $t6, 0
rff_l:
        bge  $t6, $t0, rff_nf
        mul  $t7, $t6, $t1
        addu $t7, $t2, $t7
        addu $t7, $t7, $t4
        addu $t8, $t3, $t4
        li   $t9, 0
rff_c:
        bge  $t9, $t5, rff_f
        addu $v0, $t7, $t9
        lbu  $v0, 0($v0)
        addu $v1, $t8, $t9
        lbu  $v1, 0($v1)
        bne  $v0, $v1, rff_n
        addiu $t9, $t9, 1
        j    rff_c
rff_n:
        addiu $t6, $t6, 1
        j    rff_l
rff_f:
        move $v0, $t6
        jr   $ra
rff_nf:
        li   $v0, -1
        jr   $ra

# rt_f_readkey(a0=fcb)
rt_f_readkey:
        addiu $sp, $sp, -8
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        lw   $t0, 4($a0)
        li   $t1, 1
        beq  $t0, $t1, rfk_m
        li   $t1, 3
        beq  $t0, $t1, rfk_m
        li   $t1, 47
        sw   $t1, 8($a0)
        j    rfk_r
rfk_m:
        jal  rf_find
        lw   $a0, 4($sp)
        bgez $v0, rfk_f
        li   $t1, 23
        sw   $t1, 8($a0)
        j    rfk_r
rfk_f:
        addiu $t0, $v0, 1
        sw   $t0, 40($a0)
        lw   $t2, 16($a0)
        mul  $t3, $v0, $t2
        lw   $t4, 32($a0)
        addu $t4, $t4, $t3
        lw   $t5, 20($a0)
        li   $t6, 0
rfk_c:
        bge  $t6, $t2, rfk_d
        addu $t7, $t4, $t6
        lbu  $t7, 0($t7)
        addu $t8, $t5, $t6
        sb   $t7, 0($t8)
        addiu $t6, $t6, 1
        j    rfk_c
rfk_d:
        sw   $zero, 8($a0)
rfk_r:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 8
        jr   $ra

# rt_f_start(a0=fcb)
rt_f_start:
        addiu $sp, $sp, -8
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        lw   $t0, 4($a0)
        li   $t1, 1
        beq  $t0, $t1, rfs_m
        li   $t1, 3
        beq  $t0, $t1, rfs_m
        li   $t1, 47
        sw   $t1, 8($a0)
        j    rfs_r
rfs_m:
        jal  rf_find
        lw   $a0, 4($sp)
        bgez $v0, rfs_f
        li   $t1, 23
        sw   $t1, 8($a0)
        j    rfs_r
rfs_f:
        sw   $v0, 40($a0)
        sw   $zero, 8($a0)
rfs_r:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 8
        jr   $ra

# rt_f_write(a0=fcb, a1=registro, a2=tam)
rt_f_write:
        addiu $sp, $sp, -16
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        lw   $t0, 4($a0)
        li   $t1, 2
        beq  $t0, $t1, rfw_m
        li   $t1, 3
        beq  $t0, $t1, rfw_m
        li   $t1, 4
        beq  $t0, $t1, rfw_m
        li   $t1, 48
        sw   $t1, 8($a0)
        j    rfw_r
rfw_m:
        lw   $t0, 12($a0)
        li   $t1, 3
        beq  $t0, $t1, rfw_idx
        li   $t1, 2
        beq  $t0, $t1, rfw_line
        lw   $a0, 0($a0)
        li   $v0, 15
        syscall
        j    rfw_ok
rfw_line:
rfw_tr:
        blez $a2, rfw_lw
        addu $t2, $a1, $a2
        lbu  $t2, -1($t2)
        li   $t3, 32
        bne  $t2, $t3, rfw_lw
        addiu $a2, $a2, -1
        j    rfw_tr
rfw_lw:
        lw   $a0, 0($a0)
        li   $v0, 15
        syscall
        lw   $a0, 4($sp)
        lw   $a0, 0($a0)
        la   $a1, RT_FNL
        li   $a2, 1
        li   $v0, 15
        syscall
        j    rfw_ok
rfw_idx:
        jal  rf_find
        lw   $a0, 4($sp)
        bltz $v0, rfw_ni
        li   $t1, 22
        sw   $t1, 8($a0)
        j    rfw_r
rfw_ni:
        lw   $t0, 36($a0)
        li   $t1, 256
        blt  $t0, $t1, rfw_ad
        li   $t1, 34
        sw   $t1, 8($a0)
        j    rfw_r
rfw_ad:
        lw   $t2, 16($a0)
        mul  $t3, $t0, $t2
        lw   $t4, 32($a0)
        addu $t4, $t4, $t3
        lw   $t5, 20($a0)
        li   $t6, 0
rfw_ic:
        bge  $t6, $t2, rfw_id
        addu $t7, $t5, $t6
        lbu  $t7, 0($t7)
        addu $t8, $t4, $t6
        sb   $t7, 0($t8)
        addiu $t6, $t6, 1
        j    rfw_ic
rfw_id:
        addiu $t0, $t0, 1
        sw   $t0, 36($a0)
        li   $t1, 1
        sw   $t1, 44($a0)
rfw_ok:
        lw   $a0, 4($sp)
        sw   $zero, 8($a0)
rfw_r:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 16
        jr   $ra
