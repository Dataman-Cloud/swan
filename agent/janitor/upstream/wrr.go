package upstream

type wrrBalancer struct {
	index int
	cw    int
}

func (b *wrrBalancer) Next(bs []*Backend) *Backend {
	if len(bs) == 0 {
		return nil
	}

	gcd := getGcd(bs)

	max := getMaxWeight(bs)

	for {
		b.index = (b.index + 1) % len(bs)
		if b.index == 0 {
			b.cw = b.cw - gcd
			if b.cw <= 0 {
				b.cw = max
				if b.cw == 0 {
					return nil
				}
			}
		}

		if weight := bs[b.index].Weight; int(weight) >= b.cw {
			return bs[b.index]
		}
	}
}

func getMaxWeight(backends []*Backend) int {
	max := 0
	for _, w := range backends {
		if weight := w.Weight; int(weight) >= max {
			max = int(weight)
		}
	}

	return max
}

func getGcd(backends []*Backend) int {
	divisor := -1
	for _, b := range backends {
		if divisor == -1 {
			divisor = int(b.Weight)
		} else {
			divisor = gcd(divisor, int(b.Weight))
		}
	}
	return divisor
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
