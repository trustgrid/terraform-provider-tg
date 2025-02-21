package hcl

type HCL[T any] interface {
	ToTG() T
	UpdateFromTG(T) HCL[T]
}
