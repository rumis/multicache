package utils

// IfExpr 模拟三目表达式
func IfExpr[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}
	return b
}
