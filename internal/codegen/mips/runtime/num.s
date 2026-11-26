# ===== num =====
# Num: byte 0 = sinal (0 +, 1 -), byte 1 = escala, bytes 4..43 = 40 digitos (indice 39 = menos significativo)
        .data
        .align 2
RT_SA:    .space 44
RT_SB:    .space 44
RT_SC:    .space 44
RT_SD:    .space 44
RT_PROD:  .space 320
RT_DVD:   .space 84
RT_DVS:   .space 84
RT_REM:   .space 84
RT_QUO:   .space 84
RT_EDITB: .space 48
        .text

# rt_num_clear(a0=num)
rt_num_clear:
        li   $t0, 0
rnc_l:
        li   $t1, 44
        bge  $t0, $t1, rnc_e
        addu $t2, $a0, $t0
        sw   $zero, 0($t2)
        addiu $t0, $t0, 4
        j    rnc_l
rnc_e:
        jr   $ra

# rt_copy(a0=dst, a1=src)
rt_copy:
        li   $t0, 0
rcp_l:
        li   $t1, 44
        bge  $t0, $t1, rcp_e
        addu $t2, $a1, $t0
        lw   $t3, 0($t2)
        addu $t2, $a0, $t0
        sw   $t3, 0($t2)
        addiu $t0, $t0, 4
        j    rcp_l
rcp_e:
        jr   $ra

# rt_iszero(a0=num) -> v0 = 1 se zero
rt_iszero:
        li   $t0, 0
rz_l:
        li   $t1, 40
        bge  $t0, $t1, rz_y
        addu $t2, $a0, $t0
        lbu  $t3, 4($t2)
        bnez $t3, rz_n
        addiu $t0, $t0, 1
        j    rz_l
rz_y:
        li   $v0, 1
        jr   $ra
rz_n:
        li   $v0, 0
        jr   $ra

# rt_fixzero(a0=num): sinal 0 se valor zero
rt_fixzero:
        addiu $sp, $sp, -8
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        jal  rt_iszero
        lw   $a0, 4($sp)
        beqz $v0, rfz_e
        sb   $zero, 0($a0)
rfz_e:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 8
        jr   $ra

# rt_rescale_up(a0=dst, a1=src, a2=novaescala >= escala de src)
rt_rescale_up:
        lbu  $t0, 0($a1)
        sb   $t0, 0($a0)
        lbu  $t1, 1($a1)
        subu $t1, $a2, $t1
        sb   $a2, 1($a0)
        li   $t2, 0
        li   $t3, 40
rsu_l:
        bge  $t2, $t3, rsu_e
        addu $t4, $t2, $t1
        li   $t5, 0
        bge  $t4, $t3, rsu_s
        addu $t6, $a1, $t4
        lbu  $t5, 4($t6)
rsu_s:
        addu $t6, $a0, $t2
        sb   $t5, 4($t6)
        addiu $t2, $t2, 1
        j    rsu_l
rsu_e:
        jr   $ra

# rt_shift_right(a0=num, a1=n): descarta n digitos menos significativos e reduz a escala
rt_shift_right:
        blez $a1, rsr_e
        lbu  $t0, 1($a0)
        subu $t0, $t0, $a1
        sb   $t0, 1($a0)
        li   $t1, 39
rsr_l:
        bltz $t1, rsr_e
        subu $t2, $t1, $a1
        li   $t3, 0
        bltz $t2, rsr_s
        addu $t4, $a0, $t2
        lbu  $t3, 4($t4)
rsr_s:
        addu $t4, $a0, $t1
        sb   $t3, 4($t4)
        addiu $t1, $t1, -1
        j    rsr_l
rsr_e:
        jr   $ra

# rt_magcmp(a0=a, a1=b) -> v0 = -1/0/1 (apenas digitos)
rt_magcmp:
        li   $t0, 0
rmc_l:
        li   $t1, 40
        bge  $t0, $t1, rmc_eq
        addu $t2, $a0, $t0
        lbu  $t3, 4($t2)
        addu $t2, $a1, $t0
        lbu  $t4, 4($t2)
        bne  $t3, $t4, rmc_d
        addiu $t0, $t0, 1
        j    rmc_l
rmc_d:
        sltu $t5, $t3, $t4
        li   $v0, 1
        beqz $t5, rmc_e
        li   $v0, -1
rmc_e:
        jr   $ra
rmc_eq:
        li   $v0, 0
        jr   $ra

# rt_magadd(a0=dst, a1=a, a2=b)
rt_magadd:
        li   $t0, 39
        li   $t1, 0
rma2_l:
        bltz $t0, rma2_e
        addu $t2, $a1, $t0
        lbu  $t3, 4($t2)
        addu $t2, $a2, $t0
        lbu  $t4, 4($t2)
        addu $t3, $t3, $t4
        addu $t3, $t3, $t1
        li   $t1, 0
        li   $t5, 10
        blt  $t3, $t5, rma2_s
        addiu $t3, $t3, -10
        li   $t1, 1
rma2_s:
        addu $t2, $a0, $t0
        sb   $t3, 4($t2)
        addiu $t0, $t0, -1
        j    rma2_l
rma2_e:
        jr   $ra

# rt_magsub(a0=dst, a1=a, a2=b), a >= b
rt_magsub:
        li   $t0, 39
        li   $t1, 0
rms_l:
        bltz $t0, rms_e
        addu $t2, $a1, $t0
        lbu  $t3, 4($t2)
        addu $t2, $a2, $t0
        lbu  $t4, 4($t2)
        subu $t3, $t3, $t4
        subu $t3, $t3, $t1
        li   $t1, 0
        bgez $t3, rms_s
        addiu $t3, $t3, 10
        li   $t1, 1
rms_s:
        addu $t2, $a0, $t0
        sb   $t3, 4($t2)
        addiu $t0, $t0, -1
        j    rms_l
rms_e:
        jr   $ra

# rt_add(a0=dst, a1=x, a2=y) / rt_sub
rt_add:
        li   $a3, 0
        j    rt_addsub
rt_sub:
        li   $a3, 1
rt_addsub:
        addiu $sp, $sp, -40
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        sw   $a3, 16($sp)
        lbu  $t0, 1($a1)
        lbu  $t1, 1($a2)
        move $t2, $t0
        bge  $t0, $t1, ras_1
        move $t2, $t1
ras_1:
        sw   $t2, 20($sp)
        la   $a0, RT_SA
        move $a2, $t2
        jal  rt_rescale_up
        la   $a0, RT_SB
        lw   $a1, 12($sp)
        lw   $a2, 20($sp)
        jal  rt_rescale_up
        la   $t0, RT_SA
        lbu  $t1, 0($t0)
        sw   $t1, 24($sp)
        la   $t0, RT_SB
        lbu  $t2, 0($t0)
        lw   $a3, 16($sp)
        xor  $t2, $t2, $a3
        sw   $t2, 28($sp)
        bne  $t1, $t2, ras_diff
        lw   $a0, 4($sp)
        la   $a1, RT_SA
        la   $a2, RT_SB
        jal  rt_magadd
        lw   $t1, 24($sp)
        j    ras_fin
ras_diff:
        la   $a0, RT_SA
        la   $a1, RT_SB
        jal  rt_magcmp
        beqz $v0, ras_zero
        bltz $v0, ras_neg
        lw   $a0, 4($sp)
        la   $a1, RT_SA
        la   $a2, RT_SB
        jal  rt_magsub
        lw   $t1, 24($sp)
        j    ras_fin
ras_neg:
        lw   $a0, 4($sp)
        la   $a1, RT_SB
        la   $a2, RT_SA
        jal  rt_magsub
        lw   $t1, 28($sp)
        j    ras_fin
ras_zero:
        lw   $a0, 4($sp)
        jal  rt_num_clear
        li   $t1, 0
ras_fin:
        lw   $a0, 4($sp)
        sb   $t1, 0($a0)
        lw   $t2, 20($sp)
        sb   $t2, 1($a0)
        jal  rt_fixzero
        lw   $ra, 0($sp)
        addiu $sp, $sp, 40
        jr   $ra

# rt_neg(a0=dst, a1=src)
rt_neg:
        addiu $sp, $sp, -8
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        jal  rt_copy
        lw   $a0, 4($sp)
        lbu  $t0, 0($a0)
        xori $t0, $t0, 1
        sb   $t0, 0($a0)
        jal  rt_fixzero
        lw   $ra, 0($sp)
        addiu $sp, $sp, 8
        jr   $ra

# rt_mul(a0=dst, a1=x, a2=y)
rt_mul:
        addiu $sp, $sp, -16
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        la   $t0, RT_PROD
        li   $t1, 0
rm_z:
        li   $t2, 320
        bge  $t1, $t2, rm_z2
        addu $t3, $t0, $t1
        sw   $zero, 0($t3)
        addiu $t1, $t1, 4
        j    rm_z
rm_z2:
        li   $t4, 39
rm_i:
        bltz $t4, rm_c
        addu $t5, $a1, $t4
        lbu  $t5, 4($t5)
        beqz $t5, rm_in
        li   $t6, 39
rm_j:
        bltz $t6, rm_in
        addu $t7, $a2, $t6
        lbu  $t7, 4($t7)
        mul  $t7, $t7, $t5
        addu $t8, $t4, $t6
        addiu $t8, $t8, 1
        sll  $t8, $t8, 2
        addu $t8, $t0, $t8
        lw   $t9, 0($t8)
        addu $t9, $t9, $t7
        sw   $t9, 0($t8)
        addiu $t6, $t6, -1
        j    rm_j
rm_in:
        addiu $t4, $t4, -1
        j    rm_i
rm_c:
        li   $t4, 79
        li   $t6, 10
rm_cl:
        blez $t4, rm_w
        sll  $t8, $t4, 2
        addu $t8, $t0, $t8
        lw   $t9, 0($t8)
        div  $t9, $t6
        mflo $t1
        mfhi $t2
        sw   $t2, 0($t8)
        lw   $t3, -4($t8)
        addu $t3, $t3, $t1
        sw   $t3, -4($t8)
        addiu $t4, $t4, -1
        j    rm_cl
rm_w:
        lw   $a0, 4($sp)
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        lbu  $t1, 0($a1)
        lbu  $t2, 0($a2)
        xor  $t1, $t1, $t2
        lbu  $t2, 1($a1)
        lbu  $t3, 1($a2)
        addu $t2, $t2, $t3
        sb   $t1, 0($a0)
        sb   $t2, 1($a0)
        li   $t4, 0
rm_wl:
        li   $t5, 40
        bge  $t4, $t5, rm_f
        addiu $t6, $t4, 40
        sll  $t6, $t6, 2
        addu $t6, $t0, $t6
        lw   $t7, 0($t6)
        addu $t8, $a0, $t4
        sb   $t7, 4($t8)
        addiu $t4, $t4, 1
        j    rm_wl
rm_f:
        jal  rt_fixzero
        lw   $ra, 0($sp)
        addiu $sp, $sp, 16
        jr   $ra

# rt_div(a0=dst, a1=x, a2=y, a3=escala do quociente) -> v0 = 1 se divisao por zero
# quociente truncado: dst = trunc(x / y * 10^a3) com escala a3
rt_div:
        addiu $sp, $sp, -64
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        sw   $a3, 16($sp)
        sw   $s1, 24($sp)
        sw   $s2, 28($sp)
        sw   $s3, 32($sp)
        sw   $s4, 36($sp)
        sw   $s5, 40($sp)
        sw   $s6, 44($sp)
        move $a0, $a2
        jal  rt_iszero
        beqz $v0, rd_go
        lw   $a0, 4($sp)
        jal  rt_num_clear
        lw   $a0, 4($sp)
        lw   $a3, 16($sp)
        sb   $a3, 1($a0)
        li   $v0, 1
        j    rd_ret
rd_go:
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        lw   $a3, 16($sp)
        lbu  $t0, 1($a1)
        lbu  $t1, 1($a2)
        addu $t2, $t1, $a3
        subu $t2, $t2, $t0
        sw   $t2, 20($sp)
        la   $t3, RT_DVD
        la   $t4, RT_DVS
        li   $t5, 0
rd_zc:
        li   $t6, 84
        bge  $t5, $t6, rd_fill
        addu $t7, $t3, $t5
        sw   $zero, 0($t7)
        addu $t7, $t4, $t5
        sw   $zero, 0($t7)
        la   $t7, RT_REM
        addu $t7, $t7, $t5
        sw   $zero, 0($t7)
        addiu $t5, $t5, 4
        j    rd_zc
rd_fill:
        li   $t5, 0
rd_fl:
        li   $t6, 40
        bge  $t5, $t6, rd_sh
        addu $t7, $a1, $t5
        lbu  $t8, 4($t7)
        addu $t7, $t3, $t5
        addiu $t7, $t7, 40
        sb   $t8, 0($t7)
        addu $t7, $a2, $t5
        lbu  $t8, 4($t7)
        addu $t7, $t4, $t5
        addiu $t7, $t7, 40
        sb   $t8, 0($t7)
        addiu $t5, $t5, 1
        j    rd_fl
rd_sh:
        lw   $t2, 20($sp)
        move $a0, $t3
        bgez $t2, rd_shd
        subu $t2, $zero, $t2
        move $a0, $t4
rd_shd:
        move $a1, $t2
        jal  rd_shl
        la   $s4, RT_DVD
        la   $s3, RT_DVS
        la   $s2, RT_REM
        la   $s5, RT_QUO
        li   $s1, 0
rd_k:
        li   $t1, 80
        bge  $s1, $t1, rd_out
        li   $t1, 0
rd_rs:
        li   $t2, 80
        bge  $t1, $t2, rd_rd
        addu $t7, $s2, $t1
        lbu  $t8, 1($t7)
        sb   $t8, 0($t7)
        addiu $t1, $t1, 1
        j    rd_rs
rd_rd:
        addu $t7, $s4, $s1
        lbu  $t8, 0($t7)
        sb   $t8, 80($s2)
        li   $s6, 0
rd_sub:
        move $a0, $s2
        move $a1, $s3
        jal  rd_cmp81
        bltz $v0, rd_q
        move $a0, $s2
        move $a1, $s3
        jal  rd_sub81
        addiu $s6, $s6, 1
        j    rd_sub
rd_q:
        addu $t7, $s5, $s1
        sb   $s6, 0($t7)
        addiu $s1, $s1, 1
        j    rd_k
rd_out:
        lw   $a0, 4($sp)
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        lw   $a3, 16($sp)
        lbu  $t1, 0($a1)
        lbu  $t2, 0($a2)
        xor  $t1, $t1, $t2
        sb   $t1, 0($a0)
        sb   $a3, 1($a0)
        li   $t0, 0
rd_ol:
        li   $t1, 40
        bge  $t0, $t1, rd_of
        addiu $t2, $t0, 40
        addu $t2, $s5, $t2
        lbu  $t3, 0($t2)
        addu $t4, $a0, $t0
        sb   $t3, 4($t4)
        addiu $t0, $t0, 1
        j    rd_ol
rd_of:
        jal  rt_fixzero
        li   $v0, 0
rd_ret:
        lw   $s1, 24($sp)
        lw   $s2, 28($sp)
        lw   $s3, 32($sp)
        lw   $s4, 36($sp)
        lw   $s5, 40($sp)
        lw   $s6, 44($sp)
        lw   $ra, 0($sp)
        addiu $sp, $sp, 64
        jr   $ra

# rd_shl(a0=array de 80 bytes, a1=n): desloca n digitos para a esquerda (multiplica por 10^n)
rd_shl:
        blez $a1, rds_e
        li   $t0, 0
rds_l:
        li   $t1, 80
        bge  $t0, $t1, rds_e
        addu $t2, $t0, $a1
        li   $t3, 0
        bge  $t2, $t1, rds_s
        addu $t4, $a0, $t2
        lbu  $t3, 0($t4)
rds_s:
        addu $t4, $a0, $t0
        sb   $t3, 0($t4)
        addiu $t0, $t0, 1
        j    rds_l
rds_e:
        jr   $ra

# rd_cmp81(a0=R[81 bytes], a1=D[80 bytes alinhados a direita]) -> v0 = -1/0/1
rd_cmp81:
        lbu  $t0, 0($a0)
        bnez $t0, rdc_gt
        li   $t1, 0
rdc_l:
        li   $t2, 80
        bge  $t1, $t2, rdc_eq
        addu $t3, $a0, $t1
        lbu  $t4, 1($t3)
        addu $t3, $a1, $t1
        lbu  $t5, 0($t3)
        bne  $t4, $t5, rdc_d
        addiu $t1, $t1, 1
        j    rdc_l
rdc_d:
        sltu $t6, $t4, $t5
        li   $v0, 1
        beqz $t6, rdc_e
        li   $v0, -1
rdc_e:
        jr   $ra
rdc_gt:
        li   $v0, 1
        jr   $ra
rdc_eq:
        li   $v0, 0
        jr   $ra

# rd_sub81(a0=R[81], a1=D[80]): R -= D
rd_sub81:
        li   $t0, 79
        li   $t1, 0
rds2_l:
        bltz $t0, rds2_h
        addu $t2, $a0, $t0
        lbu  $t3, 1($t2)
        addu $t4, $a1, $t0
        lbu  $t5, 0($t4)
        subu $t3, $t3, $t5
        subu $t3, $t3, $t1
        li   $t1, 0
        bgez $t3, rds2_s
        addiu $t3, $t3, 10
        li   $t1, 1
rds2_s:
        sb   $t3, 1($t2)
        addiu $t0, $t0, -1
        j    rds2_l
rds2_h:
        lbu  $t3, 0($a0)
        subu $t3, $t3, $t1
        sb   $t3, 0($a0)
        jr   $ra

# rt_num_to_int(a0=num) -> v0 = parte inteira com sinal
rt_num_to_int:
        lbu  $t0, 1($a0)
        li   $t1, 39
        subu $t1, $t1, $t0
        li   $v0, 0
        bltz $t1, rni_s
        li   $t2, 0
        li   $t5, 10
rni_l:
        bgt  $t2, $t1, rni_s
        addu $t3, $a0, $t2
        lbu  $t4, 4($t3)
        mul  $v0, $v0, $t5
        addu $v0, $v0, $t4
        addiu $t2, $t2, 1
        j    rni_l
rni_s:
        lbu  $t0, 0($a0)
        beqz $t0, rni_e
        subu $v0, $zero, $v0
rni_e:
        jr   $ra

# rt_num_from_int(a0=num, a1=valor, a2=escala)
rt_num_from_int:
        addiu $sp, $sp, -16
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        jal  rt_num_clear
        lw   $a0, 4($sp)
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        sb   $a2, 1($a0)
        bgez $a1, rnf_p
        li   $t0, 1
        sb   $t0, 0($a0)
        subu $a1, $zero, $a1
rnf_p:
        li   $t0, 39
        li   $t1, 10
rnf_l:
        bltz $t0, rnf_e
        beqz $a1, rnf_e
        div  $a1, $t1
        mfhi $t2
        mflo $a1
        addu $t3, $a0, $t0
        sb   $t2, 4($t3)
        addiu $t0, $t0, -1
        j    rnf_l
rnf_e:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 16
        jr   $ra

# rt_pow(a0=dst, a1=base, a2=expoente) -> v0 = 1 se expoente invalido
rt_pow:
        addiu $sp, $sp, -24
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        move $a0, $a2
        jal  rt_num_to_int
        sw   $v0, 16($sp)
        bgez $v0, rp_ok
        lw   $a0, 4($sp)
        jal  rt_num_clear
        li   $v0, 1
        j    rp_ret
rp_ok:
        la   $a0, RT_SC
        li   $a1, 1
        li   $a2, 0
        jal  rt_num_from_int
rp_l:
        lw   $t0, 16($sp)
        blez $t0, rp_d
        addiu $t0, $t0, -1
        sw   $t0, 16($sp)
        la   $a0, RT_SC
        la   $a1, RT_SC
        lw   $a2, 8($sp)
        jal  rt_mul
        la   $a0, RT_SC
        lbu  $t1, 1($a0)
        li   $t2, 18
        ble  $t1, $t2, rp_l
        subu $a1, $t1, $t2
        jal  rt_shift_right
        j    rp_l
rp_d:
        lw   $a0, 4($sp)
        la   $a1, RT_SC
        jal  rt_copy
        li   $v0, 0
rp_ret:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 24
        jr   $ra

# rt_cmp(a0=x, a1=y) -> v0 = -1/0/1
rt_cmp:
        addiu $sp, $sp, -24
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        lbu  $t0, 1($a0)
        lbu  $t1, 1($a1)
        move $t2, $t0
        bge  $t0, $t1, rc_1
        move $t2, $t1
rc_1:
        sw   $t2, 12($sp)
        move $a1, $a0
        la   $a0, RT_SA
        move $a2, $t2
        jal  rt_rescale_up
        la   $a0, RT_SB
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        jal  rt_rescale_up
        la   $a0, RT_SA
        jal  rt_fixzero
        la   $a0, RT_SB
        jal  rt_fixzero
        la   $t0, RT_SA
        lbu  $t1, 0($t0)
        la   $t0, RT_SB
        lbu  $t2, 0($t0)
        bne  $t1, $t2, rc_sg
        sw   $t1, 16($sp)
        la   $a0, RT_SA
        la   $a1, RT_SB
        jal  rt_magcmp
        lw   $t1, 16($sp)
        beqz $t1, rc_ret
        subu $v0, $zero, $v0
        j    rc_ret
rc_sg:
        li   $v0, 1
        beqz $t1, rc_ret
        li   $v0, -1
rc_ret:
        lw   $ra, 0($sp)
        addiu $sp, $sp, 24
        jr   $ra

# rt_load(a0=num, a1=addr, a2=size, a3=flags: bit8 sinal, byte0 escala)
rt_load:
        li   $t0, 0
rld_c:
        li   $t1, 44
        bge  $t0, $t1, rld_s
        addu $t2, $a0, $t0
        sw   $zero, 0($t2)
        addiu $t0, $t0, 4
        j    rld_c
rld_s:
        andi $t0, $a3, 255
        sb   $t0, 1($a0)
        li   $t0, 40
        subu $t0, $t0, $a2
        addu $t0, $a0, $t0
        addiu $t0, $t0, 4
        li   $t1, 0
        li   $t6, 0
rld_l:
        bge  $t1, $a2, rld_n
        addu $t2, $a1, $t1
        lbu  $t3, 0($t2)
        andi $t4, $a3, 256
        beqz $t4, rld_d
        addiu $t5, $a2, -1
        bne  $t1, $t5, rld_d
        li   $t5, 112
        blt  $t3, $t5, rld_d
        addiu $t3, $t3, -64
        li   $t6, 1
rld_d:
        addiu $t3, $t3, -48
        bltz $t3, rld_z
        li   $t5, 9
        ble  $t3, $t5, rld_w
rld_z:
        li   $t3, 0
rld_w:
        addu $t2, $t0, $t1
        sb   $t3, 0($t2)
        addiu $t1, $t1, 1
        j    rld_l
rld_n:
        sb   $t6, 0($a0)
        jr   $ra

# rt_store(a0=num, a1=addr, a2=size, a3=flags: bit9 nao grava em overflow, bit8 sinal, byte0 escala) -> v0 = overflow
rt_store:
        lbu  $t0, 1($a0)
        andi $t1, $a3, 255
        subu $t2, $t0, $t1
        addu $t3, $a2, $t2
        bgez $t3, rst_o
        li   $t3, 0
rst_o:
        li   $t4, 0
rst_ol:
        li   $t5, 39
        bgt  $t3, $t5, rst_od
        subu $t6, $t5, $t3
        addu $t6, $a0, $t6
        lbu  $t7, 4($t6)
        beqz $t7, rst_on
        li   $t4, 1
rst_on:
        addiu $t3, $t3, 1
        j    rst_ol
rst_od:
        andi $t5, $a3, 512
        beqz $t5, rst_w
        beqz $t4, rst_w
        move $v0, $t4
        jr   $ra
rst_w:
        li   $t3, 0
        li   $t8, 0
rst_wl:
        bge  $t3, $a2, rst_s
        addu $t5, $t3, $t2
        li   $t6, 0
        bltz $t5, rst_d
        li   $t7, 39
        bgt  $t5, $t7, rst_d
        subu $t7, $t7, $t5
        addu $t7, $a0, $t7
        lbu  $t6, 4($t7)
rst_d:
        beqz $t6, rst_z
        li   $t8, 1
rst_z:
        addiu $t6, $t6, 48
        subu $t7, $a2, $t3
        addiu $t7, $t7, -1
        addu $t7, $a1, $t7
        sb   $t6, 0($t7)
        addiu $t3, $t3, 1
        j    rst_wl
rst_s:
        andi $t5, $a3, 256
        beqz $t5, rst_f
        lbu  $t6, 0($a0)
        beqz $t6, rst_f
        beqz $t8, rst_f
        addiu $t7, $a2, -1
        addu $t7, $a1, $t7
        lbu  $t6, 0($t7)
        addiu $t6, $t6, 64
        sb   $t6, 0($t7)
rst_f:
        move $v0, $t4
        jr   $ra

# rt_accept_num(a0=dst, a1=size, a2=flags: bit10 virgula decimal, bit9 ignorado, bit8 sinal, byte0 escala)
rt_accept_num:
        addiu $sp, $sp, -24
        sw   $ra, 0($sp)
        sw   $a0, 4($sp)
        sw   $a1, 8($sp)
        sw   $a2, 12($sp)
        jal  rt_readline
        lw   $a1, 8($sp)
        lw   $a2, 12($sp)
        move $t9, $v0
        ble  $t9, $a1, ran_n
        move $t9, $a1
ran_n:
        la   $a0, RT_SC
        jal  rt_num_clear
        la   $t0, RT_INBUF
        li   $t1, 0
        li   $t2, 0
        li   $t3, 0
        li   $t4, 0
        li   $t8, 46
        andi $t5, $a2, 1024
        beqz $t5, ran_s
        li   $t8, 44
ran_s:
        bge  $t1, $t9, ran_p
        lbu  $t5, 0($t0)
        li   $t6, 45
        bne  $t5, $t6, ran_pl
        li   $t2, 1
        addiu $t1, $t1, 1
        j    ran_l
ran_pl:
        li   $t6, 43
        bne  $t5, $t6, ran_l
        addiu $t1, $t1, 1
ran_l:
        bge  $t1, $t9, ran_p
        addu $t5, $t0, $t1
        lbu  $t5, 0($t5)
        beq  $t5, $t8, ran_dp
        li   $t6, 48
        blt  $t5, $t6, ran_p
        li   $t6, 57
        bgt  $t5, $t6, ran_p
        li   $t6, 40
        bge  $t3, $t6, ran_nx
        la   $t7, RT_EDITB
        addu $t7, $t7, $t3
        addiu $t5, $t5, -48
        sb   $t5, 0($t7)
        addiu $t3, $t3, 1
        beqz $t4, ran_nx
        addiu $t4, $t4, 1
        j    ran_nx
ran_dp:
        bnez $t4, ran_p
        li   $t4, 1
ran_nx:
        addiu $t1, $t1, 1
        j    ran_l
ran_p:
        la   $a0, RT_SC
        beqz $t4, ran_nf
        addiu $t4, $t4, -1
ran_nf:
        sb   $t4, 1($a0)
        sb   $t2, 0($a0)
        li   $t5, 0
ran_cp:
        bge  $t5, $t3, ran_st
        la   $t7, RT_EDITB
        addu $t7, $t7, $t5
        lbu  $t6, 0($t7)
        li   $t7, 40
        subu $t7, $t7, $t3
        addu $t7, $t7, $t5
        addu $t7, $a0, $t7
        sb   $t6, 4($t7)
        addiu $t5, $t5, 1
        j    ran_cp
ran_st:
        jal  rt_fixzero
        la   $a0, RT_SC
        lw   $a1, 4($sp)
        lw   $a2, 8($sp)
        lw   $a3, 12($sp)
        andi $a3, $a3, 511
        jal  rt_store
        lw   $ra, 0($sp)
        addiu $sp, $sp, 24
        jr   $ra
