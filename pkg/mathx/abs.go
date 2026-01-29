package mathx

type signIntType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func AbsInt[T signIntType](v T) T {
	if v >= 0 {
		return v
	}
	return -v
}
