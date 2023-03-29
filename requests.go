package main

import "sync/atomic"

type MatrixRequestHandler interface {
	Process()
}

type MatrixResult interface {
	Completed() bool
	Result() (matrix, error)
}

type MatMulAddReq struct {
	A []matrix
	B []matrix

	ReqId uint64

	Done   uint32
	result matrix
	err    error
}

func (m *MatMulAddReq) Process() {
	// Returns:
	// <rows, cols>
	// A1*B1 + A2*B2 + ... + An*Bn
	// TODO: run in goroutine
	m.result, m.err = MatrixMultiplyAndAdd(m.A, m.B)
	m.A = nil
	m.B = nil
	atomic.StoreUint32(&m.Done, 1)
}

func (m *MatMulAddReq) Completed() bool {
	return atomic.LoadUint32(&m.Done) == 1
}

func (m *MatMulAddReq) Result() (matrix, error) {
	return m.result, m.err
}

type MatAddReq struct {
	Res matrix
	A []matrix

	Done   uint32
	err    error
}

func (m *MatAddReq) Completed() bool {
	return atomic.LoadUint32(&m.Done) == 1
}

func (m *MatAddReq) Result() (matrix, error) {
	return m.Res, m.err
}

func (m *MatAddReq) Process() {
	// Returns:
	// <rows, cols>
	// A1*B1 + A2*B2 + ... + An*Bn
	// TODO: run in goroutine
	m.err = MatrixAddAgg(&m.Res, m.A)
	atomic.StoreUint32(&m.Done, 1)
	m.A = nil
}

type MatStoreReq struct {
	A matrix
}

func (m *MatStoreReq) Completed() bool {
	return true
}

func (m *MatStoreReq) Result() (matrix, error) {
	return m.A, nil
}
