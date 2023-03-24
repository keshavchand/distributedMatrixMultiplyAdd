package main

import "testing"

func TestMatMulAddSingle(t *testing.T) {
	var A matrix
	var B matrix

	A.r = 4
	A.c = 4
	A.nums = make([]int, 16)
	for i := 0; i < 16; i++ {
		A.nums[i] = i
	}

	B.r = 4
	B.c = 4
	B.nums = make([]int, 16)
	for i := 0; i < 16; i++ {
		B.nums[i] = i
	}

	res, err := MatrixMultiplyAndAdd([]matrix{A}, []matrix{B})
	if err != nil {
		t.Errorf("Error in matmul_add: %v", err)
	}

	actualResult := []int{
			56,  62,  68,  74,
			152, 174, 196, 218,
			248, 286, 324, 362,
			344, 398, 452, 506,
		}

	for i := 0; i < 16; i++ {
		if res.nums[i] != actualResult[i] {
			t.Errorf("error in multiplication")
			return
		}
	}
}

func TestMatMulAddMultiple(t *testing.T) {
	var A matrix
	var B matrix

	A.r = 4
	A.c = 4
	A.nums = make([]int, 16)
	for i := 0; i < 16; i++ {
		A.nums[i] = i
	}

	B.r = 4
	B.c = 4
	B.nums = make([]int, 16)
	for i := 0; i < 16; i++ {
		B.nums[i] = i
	}

	res, err := MatrixMultiplyAndAdd([]matrix{A, A}, []matrix{B, B})
	if err != nil {
		t.Errorf("Error in matmul_add: %v", err)
	}

	actualResult := []int{
				112,  124,  136,  148,
				304,  348,  392,  436,
				496,  572,  648,  724,
				688,  796,  904,  1012,
		}

	for i := 0; i < 16; i++ {
		if res.nums[i] != actualResult[i] {
			t.Errorf("error in multiplication")
			return
		}
	}
}
