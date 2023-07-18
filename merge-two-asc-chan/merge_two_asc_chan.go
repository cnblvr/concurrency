package merge_two_asc_chan

func MergeTwoAscChan(c1, c2 <-chan int) <-chan int {
	sink := make(chan int, 2)

	go func() {
		defer close(sink)

		pour := func(dst chan int, src <-chan int) {
			for v := range src {
				dst <- v
			}
		}

		var m1, m2 *int

		for {
			if m1 == nil {
				v1, ok := <-c1
				if ok {
					m1 = &v1
				} else {
					if m2 != nil {
						sink <- *m2
					}
					pour(sink, c2)
					return
				}
			}
			if m2 == nil {
				v2, ok := <-c2
				if ok {
					m2 = &v2
				} else {
					if m1 != nil {
						sink <- *m1
					}
					pour(sink, c1)
					return
				}
			}
			if *m1 < *m2 {
				sink <- *m1
				m1 = nil
			} else {
				sink <- *m2
				m2 = nil
			}
		}
	}()

	return sink
}
