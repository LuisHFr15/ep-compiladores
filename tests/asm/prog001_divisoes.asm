# gerado por cobolc
	.data
	.align 2
DATA:
	.byte 0
S1:	.byte 72,101,108,108,111,32,119,111,114,108,100
S2:	.byte 69,110,105,111,32,66,97,115,115,111,32,45,32,50
	.align 2

	.text
	.globl main
main:
	la $s0, DATA
P0:
	la $a0, S1
	li $a1, 11
	jal rt_print
	jal rt_print_nl
	la $a0, S2
	li $a1, 14
	jal rt_print
	jal rt_print_nl
	li $v0, 10
	syscall
E0:
	lw $t2, RT_PSP
	beqz $t2, K3
	la $t3, RT_PSTK
	addu $t3, $t3, $t2
	lw $t4, -4($t3)
	la $t5, E0
	bne $t4, $t5, K3
	lw $t4, -8($t3)
	addiu $t2, $t2, -8
	sw $t2, RT_PSP
	jr $t4
K3:
	li $v0, 10
	syscall

# ===== core =====
        .data
        .align 2
RT_INBUF: .space 264
RT_PSTK:  .space 8192
RT_PSP:   .word 0
        .text

# rt_print(a0=addr, a1=len)
rt_print:
        move $t0, $a0
        move $t1, $a1
rt_print_l:
        blez $t1, rt_print_e
        lbu  $a0, 0($t0)
        li   $v0, 11
        syscall
        addiu $t0, $t0, 1
        addiu $t1, $t1, -1
        j    rt_print_l
rt_print_e:
        jr   $ra

rt_print_nl:
        li   $a0, 10
        li   $v0, 11
        syscall
        jr   $ra

# rt_print_signed(a0=addr, a1=size): digitos sem overpunch + sinal final
rt_print_signed:
        move $t0, $a0
        move $t1, $a1
        li   $t2, 0
rps_l:
        blez $t1, rps_d
        lbu  $a0, 0($t0)
        li   $t3, 1
        bne  $t1, $t3, rps_p
        li   $t3, 112
        blt  $a0, $t3, rps_p
        addiu $a0, $a0, -64
        li   $t2, 1
rps_p:
        li   $v0, 11
        syscall
        addiu $t0, $t0, 1
        addiu $t1, $t1, -1
        j    rps_l
rps_d:
        li   $a0, 43
        beqz $t2, rps_s
        li   $a0, 45
rps_s:
        li   $v0, 11
        syscall
        jr   $ra

# rt_print_index(a0=addr): '+' e 9 digitos
rt_print_index:
        move $t0, $a0
        li   $a0, 43
        li   $v0, 11
        syscall
        li   $t1, 9
rpi_l:
        blez $t1, rpi_e
        lbu  $a0, 0($t0)
        li   $v0, 11
        syscall
        addiu $t0, $t0, 1
        addiu $t1, $t1, -1
        j    rpi_l
rpi_e:
        jr   $ra

# rt_readline() -> v0 = tamanho da linha em RT_INBUF (sem \n)
rt_readline:
        la   $a0, RT_INBUF
        li   $a1, 256
        li   $v0, 8
        syscall
        la   $t0, RT_INBUF
        li   $v0, 0
rl_l:
        lbu  $t1, 0($t0)
        beqz $t1, rl_e
        li   $t2, 10
        beq  $t1, $t2, rl_e
        li   $t2, 13
        beq  $t1, $t2, rl_e
        addiu $v0, $v0, 1
        addiu $t0, $t0, 1
        j    rl_l
rl_e:
        la   $t0, RT_INBUF
        addu $t0, $t0, $v0
        sb   $zero, 0($t0)
        jr   $ra

# rt_exit()
rt_exit:
        li   $v0, 10
        syscall

