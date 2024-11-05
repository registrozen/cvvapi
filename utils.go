package cvvapi

import "iter"

func rangeRef[T any](s []T) iter.Seq[*T] {
	return func(yield func(item *T) bool) {
		for idx := range s {
			if !yield(&s[idx]) {
				return
			}
		}
	}
}

func rangeRef2[T any](s []T) iter.Seq2[int, *T] {
	return func(yield func(key int, item *T) bool) {
		for idx := range s {
			if !yield(idx, &s[idx]) {
				return
			}
		}
	}
}

func filter[T any](s []T, p func(i T) bool) []T {
	res := make([]T, 0)
	for i := range s {
		if p(s[i]) {
			res = append(res, s[i])
		}
	}
	return res
}

func rangeFilter[T any](s []T, p func(i T) bool) iter.Seq[T] {
	return func(yield func(item T) bool) {
		for idx := range s {
			if p(s[idx]) {
				if !yield(s[idx]) {
					return
				}
			}
		}
	}
}

func mapFunc[T any, D any](s []T, p func(i T) D) []D {
	res := make([]D, len(s))
	for i := range s {
		res[i] = p(s[i])
	}
	return res
}

func asRefs[T any](s []T) []*T {
	res := make([]*T, len(s))
	for i := range s {
		res[i] = &s[i]
	}
	return res
}

func findFirst[T any](s []T, p func(i *T) bool) *T {
	for idx := range s {
		if p(&s[idx]) {
			return &s[idx]
		}
	}

	return nil
}

func findFirst2[T any](s []T, p func(i *T) bool) (int, *T) {
	for idx := range s {
		if p(&s[idx]) {
			return idx, &s[idx]
		}
	}

	return -1, nil
}

func ifExpr[T any](expr bool, ifTrue T, ifFalse T) T {
	if expr {
		return ifTrue
	} else {
		return ifFalse
	}
}