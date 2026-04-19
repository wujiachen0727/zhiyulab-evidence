package internal

// ProcessData 是导出函数，编译器在跨包调用时可能放弃逃逸分析
func ProcessData(data *int) int {
	*data = *data * 2
	return *data
}

// processData 是未导出函数，同一包内调用时编译器有更多优化空间
func processData(data *int) int {
	*data = *data + 1
	return *data
}
