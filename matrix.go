package main

import "fmt"

func MatrixMul(res *matrix, A, B matrix) error {
	if A.c != B.r {
		return fmt.Errorf("rix multiplication: incompatible")
	}
	if A.r != res.r || B.c != res.c {
		return fmt.Errorf("matrix multiplication: doesn't have enough space in result matrix")
	}

	// TODO: transpose matrix for faster speed
	for i := 0; i < A.r; i++ {
		for j := 0; j < B.c; j++ {
			res.nums[i*res.c+j] = 0
			for k := 0; k < A.c; k++ {
				res.nums[i*res.c+j] += A.nums[i*A.c+k] * B.nums[k*B.c+j]
			}
		}
	}

	return nil
}

func MatrixAdd(res *matrix, A, B matrix) error {
	if A.r != B.r || A.c != B.c {
		return fmt.Errorf("matrix addition: incompatible")
	}
	if A.r != res.r || A.c != res.c {
		return fmt.Errorf("matrix addition: doesn't have enough space in result matrix")
	}

	// TODO: transpose matrix for faster speed
	for i := 0; i < A.r * A.c; i++ {
		res.nums[i] = A.nums[i] + B.nums[i]
	}

	return nil
}

func MatrixAddAgg(res *matrix, A []matrix) error {
	for _, a := range A {
		err := MatrixAdd(res, *res, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func MatrixMultiplyAndAdd(A, B []matrix) (matrix, error) {
	var res matrix
	if len(A) != len(B) {
		return res, fmt.Errorf("Sizes doesn't match")
	}

	res.r = A[0].r
	res.c = B[0].c
	res.nums = make([]int, res.r * res.c)

	var tmp matrix
	tmp.r = A[0].r
	tmp.c = B[0].c
	tmp.nums = make([]int, tmp.r * tmp.c)

	for idx := 0; idx < len(A); idx++ {
		matA := A[idx]
		matB := B[idx]

		var err error
		err = MatrixMul(&tmp, matA, matB)
		if err != nil {
			return res, err
		}

		err = MatrixAdd(&res, res, tmp)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}
