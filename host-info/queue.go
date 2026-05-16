package hostinfo

// 定义一个环形队列，方便对数组进行追加和覆盖
type CircularQueue struct {
	data      []uint64 // 固定长度数组
	lastValue uint64   // 上一次添加的真实值，计算差值时使用
	head      int      // 队首索引
	next      int      // 下一个写入位置
	count     int      // 当前元素个数
}

// Add 添加元素到队列
func (cq *CircularQueue) Add(value uint64, diffValue ...bool) {

	if len(diffValue) > 0 && diffValue[0] {

		if cq.lastValue == 0 {
			cq.data[cq.next] = 0
		} else {
			cq.data[cq.next] = value - cq.lastValue

		}
		cq.lastValue = value
	} else {
		cq.data[cq.next] = value
	}

	cq.next = (cq.next + 1) % len(cq.data)
	if cq.count < len(cq.data) {
		cq.count++
	} else {
		cq.head = (cq.head + 1) % len(cq.data) // 覆盖最旧元素时移动头部
	}

}

// GetAll 获取队列中所有元素，按照添加顺序返回 从最旧到最新
func (cq *CircularQueue) GetAll() []uint64 {

	// 空位用0填充
	if cq.count < len(cq.data) {
		for i := 0; i < (len(cq.data) - cq.count); i++ {
			cq.data[i] = 0
		}
	}

	result := make([]uint64, cq.count)
	for i := (len(cq.data) - cq.count); i < len(cq.data); i++ {
		index := (cq.head + i) % len(cq.data)
		result[i] = cq.data[index]
	}

	return result
}

func (cq *CircularQueue) Length() int {
	return len(cq.data)
}

func (cq *CircularQueue) Count() int {
	return cq.count
}

func (cq *CircularQueue) LastValue() uint64 {
	return cq.data[(cq.next-1+len(cq.data))%len(cq.data)]
}

// Clear 清空队列
func (cq *CircularQueue) Clear() {
	cq.head = 0
	cq.next = 0
	cq.count = 0
}
